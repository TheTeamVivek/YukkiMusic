/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/TheTeamVivek/YukkiMusic
 */

package modules

import (
	"context"
	"strconv"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/skip"] = `<i>Skip the currently playing track and play the next in queue.</i>

<u>Usage:</u>
<b>/skip</b> — Skip current track

<b>⚙️ Behavior:</b>
• Downloads next track in queue
• Starts playback automatically
• If queue is empty and loop is 0, stops playback

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>⚠️ Notes:</b>
• Cannot be undone
• If no tracks in queue, playback stops
• Loop count affects skip behavior`
}

func skipHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	return handleSkip(c, m, false)
}

func cskipHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	return handleSkip(c, m, true)
}

func handleSkip(c *td.Client, m *td.Message, cplay bool) error {
	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	chatID := m.ChatID()
	if !r.IsActiveChat() {
		m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	mention := mentionOf(nil, m.SenderID())
	skipCount := 1

	if args := m.Args(); args != "" {
		parsed, parseErr := strconv.Atoi(args)
		if parseErr != nil {
			m.ReplyText(c, F(chatID, "skip_invalid_number"), nil)
			return nil
		}

		queuedTracks := len(r.Queue())
		if queuedTracks == 0 {
			m.ReplyText(c, F(chatID, "skip_queue_empty_for_count"), nil)
			return nil
		}

		if parsed < 1 || parsed > queuedTracks {
			m.ReplyText(c, F(chatID, "skip_count_exceeds_queue", locales.Arg{
				"requested": parsed,
				"available": queuedTracks,
			}), nil)
			return nil
		}

		// /skip N means: skip current + N queued tracks.
		skipCount = parsed + 1
	}

	if len(r.Queue()) == 0 {

		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		m.ReplyText(c, F(chatID, "skip_stopped", locales.Arg{
			"user": mention,
		}), nil)
		return nil
	}

	r.SetLoop(0)

	for i := 1; i < skipCount; i++ {
		if len(r.Queue()) == 0 {

			scheduleOldPlayingMessage(r)
			core.DeleteRoom(r.ID)
			m.ReplyText(c, F(chatID, "skip_stopped", locales.Arg{
				"user": mention,
			}), nil)
			return nil
		}
		_ = r.NextTrack()
	}

	if len(r.Queue()) == 0 {

		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		m.ReplyText(c, F(chatID, "skip_stopped", locales.Arg{
			"user": mention,
		}), nil)
		return nil
	}

	t := r.NextTrack()
	if t == nil {

		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		m.ReplyText(c, F(chatID, "skip_stopped", locales.Arg{
			"user": mention,
		}), nil)
		return nil
	}

	statusMsg, err := core.Bot.SendTextMessage(
		chatID,
		F(chatID, "stream_downloading_next"),
		nil,
	)
	if err != nil {
		logger.Errorf("[skip.go] err: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, statusMsg)
	if err != nil {
		txt := F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		})

		if statusMsg != nil {
			utils.EOR(c, statusMsg, txt, nil)
		} else {
			core.Bot.SendTextMessage(chatID, txt, nil)
		}

		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		return nil
	}

	if err := r.Play(t, path, true); err != nil {
		txt := F(chatID, "stream_play_fail")
		if statusMsg != nil {
			utils.EOR(c, statusMsg, txt, nil)
		} else {
			core.Bot.SendTextMessage(chatID, txt, nil)
		}
		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		return nil
	}

	statusMsg = sendNowPlaying(c, statusMsg, chatID, r, t)
	if statusMsg != nil {
		r.SetStatusMsg(statusMsg)
	}

	return nil
}

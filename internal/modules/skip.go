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
	"html"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
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

func skipHandler(m *telegram.NewMessage) error {
	return handleSkip(m, false)
}

func cskipHandler(m *telegram.NewMessage) error {
	return handleSkip(m, true)
}

func handleSkip(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	mention := utils.MentionHTML(m.Sender)

	if len(r.Queue()) == 0 {
		core.DeleteRoom(r.ChatID())
		m.Reply(F(chatID, "skip_stopped", locales.Arg{
			"user": mention,
		}))
		return telegram.ErrEndGroup
	}

	r.SetLoop(0)
	t := r.NextTrack()

	statusMsg, err := core.Bot.SendMessage(
		chatID,
		F(chatID, "stream_downloading_next"),
	)
	if err != nil {
		gologging.ErrorF("[skip.go] err: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, statusMsg)
	if err != nil {
		txt := F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		})

		if statusMsg != nil {
			utils.EOR(statusMsg, txt)
		} else {
			core.Bot.SendMessage(chatID, txt)
		}

		core.DeleteRoom(r.ChatID())
		return telegram.ErrEndGroup
	}

	if err := r.Play(t, path, true); err != nil {
		txt := F(chatID, "stream_play_fail")
		if statusMsg != nil {
			utils.EOR(statusMsg, txt)
		} else {
			core.Bot.SendMessage(chatID, txt)
		}
		core.DeleteRoom(r.ChatID())
		return telegram.ErrEndGroup
	}

	title := utils.ShortTitle(t.Title, 25)
	safeTitle := html.EscapeString(title)

	msg := F(chatID, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    safeTitle,
		"duration": formatDuration(t.Duration),
		"by":       t.Requester,
	})

	opt := &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}

	if t.Artwork != "" && shouldShowThumb(chatID) {
		opt.Media = utils.CleanURL(t.Artwork)
	}

	var newStatusMsg *telegram.NewMessage
	if statusMsg != nil {
		newStatusMsg, _ = utils.EOR(statusMsg, msg, opt)
	} else {
		newStatusMsg, _ = core.Bot.SendMessage(chatID, msg, opt)
	}

	if newStatusMsg != nil {
		r.SetStatusMsg(newStatusMsg)
	}

	return telegram.ErrEndGroup
}

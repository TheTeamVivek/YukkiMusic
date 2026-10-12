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

	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
	"yukkimusic/ntgcalls"
)

func streamEndHandler(
	chatID int64,
	streamType ntgcalls.StreamType,
	_ ntgcalls.StreamDevice,
) {
	if streamType == ntgcalls.VideoStream {
		utils.Log.Debug("[onStreamEndHandler] Video stream ended, returning")
		return
	}

	utils.Log.Debugf("[onStreamEndHandler] Stream ended in chat %d", chatID)
	r, ok := core.Rooms.Get(chatID)
	if !ok {
		return
	}
	scheduleOldPlayingMessage(r)

	c := core.Bot
	cid := r.ControlChatID
	r.Refresh()

	var t *state.Track
	var wasLooping bool
	if r.QueueLen() == 0 && r.Loop() == 0 {
		if autoplay := pickAutoplayTrack(r, r.Track()); autoplay != nil {
			r.AddTracks(autoplay)
		}

		if r.QueueLen() == 0 {
			core.Rooms.Delete(chatID)
			if _, err := c.SendTextMessage(cid, F(cid, "stream_queue_finished"), nil); err != nil {
				utils.Log.Error(err.Error())
			}
			return
		}
	}

	wasLooping = r.Loop() > 0
	t = r.NextTrack()

	statusText := F(cid, "stream_downloading_next")
	if wasLooping && t != nil && r.FilePath() != "" {
		statusText = F(cid, "cb_replaying")
	} else if t != nil && t.Requester == F(cid, "autoplay_requester") {
		statusText = F(cid, "stream_autoplay_next", locales.Arg{
			"title": utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
		})
	}

	statusMsg, err := c.SendTextMessage(cid, statusText, nil)
	if err != nil {
		utils.Log.Errorf("[call.go] Failed to send msg: %v", err)
	}

	var filePath string
	if wasLooping && t != nil && r.FilePath() != "" {
		filePath = r.FilePath()
	} else {
		filePath, err = platforms.Download(context.Background(), t, statusMsg)
	}

	if err != nil {
		utils.Log.Errorf(
			"[onStreamEndHandler] Download failed for %s: %v",
			t.URL,
			err,
		)
		utils.EOR(c, statusMsg, F(cid, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		core.Rooms.Delete(chatID)

		return
	}

	if err := r.Play(t, filePath, true); err != nil {
		utils.Log.Errorf(
			"[onStreamEndHandler] Play failed for %s: %v",
			t.URL,
			err,
		)
		utils.EOR(c, statusMsg, F(cid, "stream_play_fail"), nil)
		core.Rooms.Delete(chatID)

		return
	}

	statusMsg = sendNowPlaying(c, statusMsg, cid, r, t)
	r.SetStatusMsg(statusMsg)
}

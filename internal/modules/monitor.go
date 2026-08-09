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
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
)

func MonitorRooms() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	sem := make(chan struct{}, 20)

	for range ticker.C {
		for chatID, room := range core.GetAllRooms() {

			sem <- struct{}{}

			go func(chatID int64, r *core.RoomState) {
				defer func() { <-sem }()

				if !r.IsActiveChat() {
					/*
						// TODO: TEST IT AND INCREASE SLEEP TIME
						time.Sleep(5 * time.Second)

						if !r.IsActiveChat() {
							core.DeleteRoom(chatID)
							return
						}
					*/
					return
				}

				if r.IsPaused() {
					return
				}

				r.Parse()

				track := r.Track()
				statusMsg := r.StatusMsg()
				if track == nil || statusMsg == nil {
					return
				}

				text := nowPlayingText(r.ChatID, track)
				markup := core.GetPlayMarkup(r.ChatID, r, false)

				switch statusMsg.Content.(type) {
				case *td.MessagePhoto:
					_, _ = statusMsg.EditCaption(core.Bot, text, &td.EditCaptionOpts{
						ParseMode:   td.ParseModeHTML,
						ReplyMarkup: markup,
					})
				default:
					_, _ = statusMsg.EditText(core.Bot, text, &td.EditTextMessageOpts{
						ParseMode:   td.ParseModeHTML,
						ReplyMarkup: markup,
					})
				}
			}(chatID, room)
		}
	}
}

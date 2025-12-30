/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package modules

import (
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
)

var logger = gologging.GetLogger("monitor")

func MonitorRooms() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	sem := make(chan struct{}, 20)

	for range ticker.C {
		for _, chatID := range core.GetAllRoomIDs() {
			sem <- struct{}{}
			go func(id int64) {
				defer func() { <-sem }()
				ass, err := core.Assistants.ForChat(id)
				if err != nil {
					gologging.ErrorF("Failed to get Assistant for %d: %v", id, err)
					return
				}

				r, ok := core.GetRoom(id, ass)
				if !ok {
					gologging.DebugF("Room not exists for %d returning..", chatID)
					return
				}
				if !r.IsActiveChat() {
					// recheck after delay before deleting
					time.Sleep(7 * time.Second)
					if r2, ok2 := core.GetRoom(id, ass); ok2 && !r2.IsActiveChat() {
						core.DeleteRoom(id)
					}
					return
				}

				if r.IsPaused() {
					gologging.DebugF("Room paused for %d returning..", chatID)

					return
				}

				r.Parse()
				mystic := r.GetMystic()
				if mystic == nil {
					gologging.DebugF("mystic is nil for %d returning..", chatID)

					return
				}
				chatID := id
				if r.IsCPlay() {
					cid, err := database.GetChatIDFromCPlayID(id)
					if err == nil {
						chatID = cid
					}
				}
				markup := core.GetPlayMarkup(chatID, r, false)
				opts := &telegram.SendOptions{
					ReplyMarkup: markup,
					Entities:    mystic.Message.Entities,
				}
				mystic.Edit(mystic.Text(), opts)
			}(chatID)
		}
	}
}

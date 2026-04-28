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

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
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
				statusMsg := r.StatusMsg()
				if statusMsg == nil {
					return
				}

				markup := core.GetPlayMarkup(r.ChatID(), r, false)
				opts := &telegram.SendOptions{
					ReplyMarkup: markup,
					Entities:    statusMsg.Message.Entities,
				}
				statusMsg.Edit(statusMsg.Text(), opts)
			}(chatID, room)
		}
	}
}

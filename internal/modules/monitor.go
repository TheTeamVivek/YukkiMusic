/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
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
				mystic := r.GetMystic()
				if mystic == nil {
					return
				}

				markup := core.GetPlayMarkup(r.EffectiveChatID(), r, false)
				opts := &telegram.SendOptions{
					ReplyMarkup: markup,
					Entities:    mystic.Message.Entities,
				}
				mystic.Edit(mystic.Text(), opts)
			}(chatID, room)
		}
	}
}

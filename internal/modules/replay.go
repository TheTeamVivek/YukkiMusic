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
	"fmt"
	"html"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func replayHandler(m *telegram.NewMessage) error {
	return handleReplay(m, false)
}

func creplayHandler(m *telegram.NewMessage) error {
	return handleReplay(m, true)
}

func handleReplay(m *telegram.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.EndGroup
	}

	if err := r.Replay(); err != nil {
		m.Reply(F(chatID, "replay_failed", locales.Arg{
			"error": err,
		}))
	} else {
		trackTitle := html.EscapeString(utils.ShortTitle(r.Track.Title, 25))
		totalDuration := formatDuration(r.Track.Duration)
		m.Reply(F(chatID, "replay_success", locales.Arg{
			"title":    trackTitle,
			"duration": totalDuration,
			"speed":    fmt.Sprintf("%.2f", r.Speed),
		}))
	}

	return telegram.EndGroup
}

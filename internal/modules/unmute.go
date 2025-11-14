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

func unmuteHandler(m *telegram.NewMessage) error {
	return handleUnmute(m, false)
}

func cunmuteHandler(m *telegram.NewMessage) error {
	return handleUnmute(m, true)
}

func handleUnmute(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	chatID := m.ChannelID()

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.EndGroup
	}

	if !r.IsMuted() {
		m.Reply(F(chatID, "unmute_already"))
		return telegram.EndGroup
	}

	title := html.EscapeString(utils.ShortTitle(r.Track.Title, 25))
	mention := utils.MentionHTML(m.Sender)

	if _, err := r.Unmute(); err != nil {
		m.Reply(F(chatID, "unmute_failed", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	// optional speed line
	var speedOpt string
	if sp := r.GetSpeed(); sp != 1.0 {
		speedOpt = F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		})
	}

	msg := F(chatID, "unmute_success", locales.Arg{
		"title":      title,
		"user":       mention,
		"speed_line": speedOpt,
	})

	m.Reply(msg)
	return telegram.EndGroup
}

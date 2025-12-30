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

func init() {
	helpTexts["/resume"] = `<i>Resume the paused playback.</i>

<u>Usage:</u>
<b>/resume</b> — Resume playback from pause

<b>⚙️ Behavior:</b>
• Continues from last paused position
• Cancels auto-resume timer if active

<b>⚠️ Notes:</b>
• Can only resume if currently paused
• Position is preserved during pause
• Speed settings remain active after resume`
}

func resumeHandler(m *telegram.NewMessage) error {
	return handleResume(m, false)
}

func cresumeHandler(m *telegram.NewMessage) error {
	return handleResume(m, true)
}

func handleResume(m *telegram.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	if !r.IsPaused() {
		m.Reply(F(chatID, "resume_already_playing"))
		return telegram.ErrEndGroup
	}

	t := r.Track()
	if _, err := r.Resume(); err != nil {
		m.Reply(F(chatID, "resume_failed", locales.Arg{
			"error": err,
		}))
	} else {
		title := html.EscapeString(utils.ShortTitle(t.Title, 25))
		pos := formatDuration(r.Position())
		total := formatDuration(t.Duration)
		mention := utils.MentionHTML(m.Sender)

		speedLine := ""
		if sp := r.GetSpeed(); sp != 1.0 {
			speedLine = F(chatID, "speed_line", locales.Arg{
				"speed": fmt.Sprintf("%.2f", r.GetSpeed()),
			})
		}

		m.Reply(F(chatID, "resume_success", locales.Arg{
			"title":      title,
			"position":   pos,
			"duration":   total,
			"user":       mention,
			"speed_line": speedLine,
		}))
	}

	return telegram.ErrEndGroup
}

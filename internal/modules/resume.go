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
	"fmt"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
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

func resumeHandler(c *td.Client, m *td.Message) error {
	return handleResume(c, m, false)
}

func cresumeHandler(c *td.Client, m *td.Message) error {
	return handleResume(c, m, true)
}

func handleResume(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	if !r.IsActiveChat() {
		m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	if !r.IsPaused() {
		m.ReplyText(c, F(chatID, "resume_already_playing"), nil)
		return nil
	}

	t := r.Track()
	if _, err := r.Resume(); err != nil {
		m.ReplyText(c, F(chatID, "resume_failed", locales.Arg{
			"error": err,
		}), nil)
	} else {
		title := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
		pos := utils.FormatDuration(r.Position())
		total := utils.FormatDuration(t.Duration)
		sender, _ := m.GetUser(c)
		mention := mentionOf(sender, m.SenderID())

		speedLine := ""
		if sp := r.Speed(); sp != 1.0 {
			speedLine = F(chatID, "speed_line", locales.Arg{
				"speed": fmt.Sprintf("%.2f", r.Speed()),
			})
		}

		m.ReplyText(c, F(chatID, "resume_success", locales.Arg{
			"title":      title,
			"position":   pos,
			"duration":   total,
			"user":       mention,
			"speed_line": speedLine,
		}), nil)
	}

	return nil
}

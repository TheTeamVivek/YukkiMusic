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

	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/resume"] = `<i>Resume the paused playback.</i>

<u>Usage:</u>
<b>/resume</b> — Resume playback from pause

<b>⚙️ Behavior:</b>
• Continues from last paused position
• Now playing card updates instantly

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
	r, err := activeRoomFor(c, m, cplay)
	if err != nil || r == nil {
		return err
	}

	chatID := m.ChatID()

	if !r.Paused() {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "resume_already_playing"), &td.SendRichTextOpts{
			ReplyMarkup: core.GetPauseToggle(chatID, r),
		})
		return err
	}

	t := r.Track()
	if err := r.Resume(); err != nil {
		_, rerr := m.ReplyRichMessageHTML(c, F(chatID, "resume_failed", locales.Arg{
			"error": utils.EscapeHTML(err.Error()),
		}), nil)
		return rerr
	}

	title := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
	pos := utils.FormatDuration(r.Position())
	total := utils.FormatDuration(t.Duration)
	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())

	speedLine := ""
	if sp := r.Speed(); sp != 1.0 {
		speedLine = `<tr><td colspan="2">` + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		}) + `</td></tr>`
	}

	if _, rerr := m.ReplyRichMessageHTML(c, F(chatID, "resume_success", locales.Arg{
		"title":      title,
		"url":        utils.EscapeHTML(t.URL),
		"position":   pos,
		"duration":   total,
		"user":       mention,
		"speed_line": speedLine,
	}), &td.SendRichTextOpts{
		ReplyMarkup: core.GetPauseToggle(chatID, r),
	}); rerr != nil {
		return rerr
	}

	updatePlaybackMessageFor(c, r.StatusMsg(), m.SenderID(), r, "playing")
	return nil
}

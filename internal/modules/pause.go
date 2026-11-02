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
	helpTexts["/pause"] = `<i>Pause the current playback.</i>

<u>Usage:</u>
<b>/pause</b> — Pause playback

<b>⚙️ Features:</b>
• Manual pause/resume control

<b>💡 Examples:</b>
<code>/pause</code> — Pause indefinitely
`
}

func pauseHandler(c *td.Client, m *td.Message) error {
	return handlePause(c, m, false)
}

func cpauseHandler(c *td.Client, m *td.Message) error {
	return handlePause(c, m, true)
}

func handlePause(c *td.Client, m *td.Message, cplay bool) error {
	r, err := activeRoomFor(c, m, cplay)
	if err != nil || r == nil {
		return err
	}

	chatID := m.ChatID()

	if r.Paused() {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "pause_already"), &td.SendRichTextOpts{
			ReplyMarkup: core.GetPauseToggle(chatID, r),
		})
		return err
	}

	if pauseErr := r.Pause(); pauseErr != nil {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "room_pause_failed", locales.Arg{
			"error": utils.EscapeHTML(pauseErr.Error()),
		}), nil)
		return err
	}

	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())
	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))

	speedLine := ""
	if sp := r.Speed(); sp != 1.0 {
		speedLine = `<tr><td colspan="2">` + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		}) + `</td></tr>`
	}

	msg := F(chatID, "pause_success", locales.Arg{
		"title":      title,
		"url":        utils.EscapeHTML(r.Track().URL),
		"position":   utils.FormatDuration(r.Position()),
		"duration":   utils.FormatDuration(r.Track().Duration),
		"user":       mention,
		"speed_line": speedLine,
	})

	if _, rerr := m.ReplyRichMessageHTML(c, msg, &td.SendRichTextOpts{
		ReplyMarkup: core.GetPauseToggle(chatID, r),
	}); rerr != nil {
		return rerr
	}

	updatePlaybackMessageFor(c, r.StatusMsg(), m.SenderID(), r, "paused")
	return nil
}

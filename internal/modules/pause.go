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
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
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

	if r.IsPaused() {
		m.ReplyText(c, F(chatID, "pause_already"), nil)
		return nil
	}

	var pauseErr error
	_, pauseErr = r.Pause()
	if pauseErr != nil {
		m.ReplyText(c, F(chatID, "room_pause_failed", locales.Arg{
			"error": pauseErr.Error(),
		}), nil)
		return nil
	}

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())
	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))

	msg := F(chatID, "pause_success", locales.Arg{
		"title":    title,
		"position": utils.FormatDuration(r.Position()),
		"duration": utils.FormatDuration(r.Track().Duration),
		"user":     mention,
	})

	if sp := r.Speed(); sp != 1.0 {
		msg += "\n" + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		})
	}

	m.ReplyText(c, msg, nil)
	return nil
}

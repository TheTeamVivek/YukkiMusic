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
	helpTexts["/unmute"] = `<i>Unmute the audio output in voice chat.</i>

<u>Usage:</u>
<b>/unmute</b> — Restore audio

<b>⚙️ Behavior:</b>
• Restores audio immediately
• Cancels auto-unmute timer if active
• Shows current playback info`
}

func unmuteHandler(c *td.Client, m *td.Message) error {
	return handleUnmute(c, m, false)
}

func cunmuteHandler(c *td.Client, m *td.Message) error {
	return handleUnmute(c, m, true)
}

func handleUnmute(c *td.Client, m *td.Message, cplay bool) error {
	if m.Args() != "" {
		return nil
	}

	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
		return nil
	}

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	chatID := m.ChatID()

	if !r.IsActiveChat() {
		m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	if !r.IsMuted() {
		m.ReplyText(c, F(chatID, "unmute_already"), nil)
		return nil
	}

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))
	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	if _, err := r.Unmute(); err != nil {
		m.ReplyText(c, F(chatID, "unmute_failed", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	// optional speed line
	var speedOpt string
	if sp := r.Speed(); sp != 1.0 {
		speedOpt = F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		})
	}

	msg := F(chatID, "unmute_success", locales.Arg{
		"title":      title,
		"user":       mention,
		"speed_line": speedOpt,
	})

	m.ReplyText(c, msg, nil)
	return nil
}

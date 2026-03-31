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

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
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

func unmuteHandler(m *tg.NewMessage) error {
	return handleUnmute(m, false)
}

func cunmuteHandler(m *tg.NewMessage) error {
	return handleUnmute(m, true)
}

func handleUnmute(m *tg.NewMessage, cplay bool) error {
	if m.Args() != "" {
		return tg.ErrEndGroup
	}
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	chatID := m.ChannelID()

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if !r.IsMuted() {
		m.Reply(F(chatID, "unmute_already"))
		return tg.ErrEndGroup
	}

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))
	mention := utils.MentionHTML(m.Sender)

	if _, err := r.Unmute(); err != nil {
		m.Reply(F(chatID, "unmute_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
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

	m.Reply(msg)
	return tg.ErrEndGroup
}

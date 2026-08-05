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

func pauseHandler(m *tg.NewMessage) error {
	return handlePause(m, false)
}

func cpauseHandler(m *tg.NewMessage) error {
	return handlePause(m, true)
}

func handlePause(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if r.IsPaused() {
		m.Reply(F(chatID, "pause_already"))
		return tg.ErrEndGroup
	}

	var pauseErr error
	_, pauseErr = r.Pause()
	if pauseErr != nil {
		m.Reply(F(chatID, "room_pause_failed", locales.Arg{
			"error": pauseErr.Error(),
		}))
		return tg.ErrEndGroup
	}

	mention := utils.MentionHTML(m.Sender)
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

	m.Reply(msg)
	return tg.ErrEndGroup
}

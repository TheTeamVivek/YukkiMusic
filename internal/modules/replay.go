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
	helpTexts["/replay"] = `<i>Restart the current track from the beginning.</i>

<u>Usage:</u>
<b>/replay</b> — Restart current track

<b>⚙️ Behavior:</b>
• Resets position to 0:00
• Maintains speed setting
• Continues playback immediately

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this
`
}

func replayHandler(c *td.Client, m *td.Message) error {
	return handleReplay(c, m, false)
}

func creplayHandler(c *td.Client, m *td.Message) error {
	return handleReplay(c, m, true)
}

func handleReplay(c *td.Client, m *td.Message, cplay bool) error {
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
	t := r.Track()

	if err := r.Replay(); err != nil {
		m.ReplyText(c, F(chatID, "replay_failed", locales.Arg{
			"error": err,
		}), nil)
	} else {
		trackTitle := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
		totalDuration := utils.FormatDuration(t.Duration)
		m.ReplyText(c, F(chatID, "replay_success", locales.Arg{
			"title":    trackTitle,
			"duration": totalDuration,
			"speed":    fmt.Sprintf("%.2f", r.Speed()),
		}), nil)
	}

	return nil
}

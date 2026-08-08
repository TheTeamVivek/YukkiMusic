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
	helpTexts["/position"] = `<i>Show current playback position and track info.</i>

<u>Usage:</u>
<b>/position</b> — Show position

<b>📊 Information Displayed:</b>
• Current track title
• Current position (MM:SS)
• Total duration (MM:SS)
• Playback speed (if not 1.0x)

<b>💡 Use Case:</b>
Quick position check without full queue display.`
}

func positionHandler(c *td.Client, m *td.Message) error {
	return handlePosition(c, m, false)
}

func cpositionHandler(c *td.Client, m *td.Message) error {
	return handlePosition(c, m, true)
}

func handlePosition(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if cplay && !filterAuthUsersTd(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	if !r.IsActiveChat() || r.Track().ID == "" {
		m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	r.Parse()

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))

	m.ReplyText(c, F(chatID, "position_now", locales.Arg{
		"title":    title,
		"position": utils.FormatDuration(r.Position()),
		"duration": utils.FormatDuration(r.Track().Duration),
		"speed":    fmt.Sprintf("%.2f", r.Speed()),
	}), nil)

	return nil
}

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
	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/mute"] = `<i>Mute the audio output in voice chat.</i>

<u>Usage:</u>
<b>/mute</b> — Mute indefinitely

<b>⚙️ Features:</b>
• Audio continues playing (progress tracked)

<b>💡 Examples:</b>
<code>/mute</code> — Mute until manual unmute

<b>⚠️ Notes:</b>
• Track continues playing in background
• Use <code>/unmute</code> to restore audio`
}

func muteHandler(c *td.Client, m *td.Message) error {
	return handleMute(c, m, false)
}

func cmuteHandler(c *td.Client, m *td.Message) error {
	return handleMute(c, m, true)
}

func handleMute(c *td.Client, m *td.Message, cplay bool) error {
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

	if r.IsMuted() {
		m.ReplyText(c, F(chatID, "mute_already_muted"), nil)
		return nil
	}

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	_, muteErr := r.Mute()

	if muteErr != nil {
		m.ReplyText(c, F(chatID, "mute_failed", locales.Arg{
			"error": muteErr.Error(),
		}), nil)
		return nil
	}

	m.ReplyText(c, F(chatID, "mute_success", locales.Arg{
		"title": utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25)),
		"user":  mention,
	}), nil)

	return nil
}

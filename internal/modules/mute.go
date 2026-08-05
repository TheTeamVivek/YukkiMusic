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
	tg "github.com/amarnathcjd/gogram/telegram"

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

func muteHandler(m *tg.NewMessage) error {
	return handleMute(m, false)
}

func cmuteHandler(m *tg.NewMessage) error {
	return handleMute(m, true)
}

func handleMute(m *tg.NewMessage, cplay bool) error {
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

	if r.IsMuted() {
		m.Reply(F(chatID, "mute_already_muted"))
		return tg.ErrEndGroup
	}

	mention := utils.MentionHTML(m.Sender)

	_, muteErr := r.Mute()

	if muteErr != nil {
		m.Reply(F(chatID, "mute_failed", locales.Arg{
			"error": muteErr.Error(),
		}))
		return tg.ErrEndGroup
	}

	m.Reply(F(chatID, "mute_success", locales.Arg{
		"title": utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25)),
		"user":  mention,
	}))

	return tg.ErrEndGroup
}

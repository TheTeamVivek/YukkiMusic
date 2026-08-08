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
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/shuffle"] = `<i>Toggle shuffle mode for the queue.</i>

<u>Usage:</u>
<b>/shuffle</b> — Show current shuffle state
<b>/shuffle on</b> — Enable shuffle
<b>/shuffle off</b> — Disable shuffle

<b>⚙️ Behavior:</b>
• Randomly reorders queue when enabled
• Affects track selection order
• Can be toggled at any time

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/shuffle on</code> — Enable shuffle mode
<code>/shuffle off</code> — Disable shuffle mode

<b>⚠️ Note:</b>
Shuffle only affects queue order, not currently playing track.`
}

func shuffleHandler(c *td.Client, m *td.Message) error {
	return handleShuffle(c, m, false)
}

func cshuffleHandler(c *td.Client, m *td.Message) error {
	return handleShuffle(c, m, true)
}

func handleShuffle(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if !filterAuthUsers(c, m) {
		return nil
	}

	arg := strings.ToLower(m.Args())

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

	r.Parse()

	if arg == "" {
		state := F(chatID, "disabled")
		cmd := getCommand(m) + " on"
		if r.Shuffle() {
			state = F(chatID, "enabled")
			cmd = getCommand(m) + " off"
		}

		m.ReplyText(c, F(chatID, "shuffle_current_state", locales.Arg{
			"state": state,
			"cmd":   cmd,
		}), nil)
		return nil
	}

	var newState bool
	if arg == "on" || arg == "enable" || arg == "true" || arg == "1" {
		newState = true
	} else if arg == "off" || arg == "disable" || arg == "false" || arg == "0" {
		newState = false
	}

	r.SetShuffle(newState)

	state := F(chatID, "disabled")
	if newState {
		state = F(chatID, "enabled")
	}

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	m.ReplyText(c, F(chatID, "shuffle_updated", locales.Arg{
		"state": state,
		"user":  mention,
	}), nil)

	return nil
}

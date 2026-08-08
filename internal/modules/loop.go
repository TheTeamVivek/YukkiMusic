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
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/loop"] = `<i>Set loop count for the current track.</i>

<u>Usage:</u>
<b>/loop</b> — Show current loop count
<b>/loop [count]</b> — Set loop count (0-10)

<b>⚙️ Behavior:</b>
• 0 = No loop (play once)
• 1-10 = Repeat track that many times
• Loop counter decrements after each playback

<b>💡 Examples:</b>
<code>/loop 0</code> — Disable loop
<code>/loop 3</code> — Loop current track 3 times
<code>/loop 10</code> — Loop current track 10 times

<b>⚠️ Notes:</b>
• Maximum loop count: 10
• Loop affects only current track
• After loops complete, plays next in queue
• If the track is forcefully skipped using <code>/skip</code>, the loop will stop and reset automatically`
}

func loopHandler(c *td.Client, m *td.Message) error {
	return handleLoop(c, m, false)
}

func cloopHandler(c *td.Client, m *td.Message) error {
	return handleLoop(c, m, true)
}

func handleLoop(c *td.Client, m *td.Message, cplay bool) error {
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
	args := strings.Fields(m.Text())
	currentLoop := r.Loop()

	if !r.IsActiveChat() {
		m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	if len(args) < 2 {
		countLine := ""
		if currentLoop > 0 {
			countLine = "\n" + F(chatID, "loop_current", locales.Arg{
				"count": currentLoop,
			})
		}

		msg := F(chatID, "loop_usage", locales.Arg{
			"cmd":        getCommandTd(m),
			"count_line": countLine,
		})

		m.ReplyText(c, msg, nil)
		return nil
	}

	newLoop, err := strconv.Atoi(args[1])
	if err != nil || newLoop < 0 || newLoop > 10 {
		m.ReplyText(c, F(chatID, "loop_invalid"), nil)
		return nil
	}

	if newLoop == currentLoop {
		m.ReplyText(c, F(chatID, "loop_already_set", locales.Arg{
			"count": currentLoop,
		}), nil)
		return nil
	}

	r.SetLoop(newLoop)

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())
	var msg string
	if newLoop == 0 {
		msg = F(chatID, "loop_disabled", locales.Arg{
			"user": mention,
		})
	} else {
		msg = F(chatID, "loop_set", locales.Arg{
			"count": newLoop,
			"user":  mention,
		})
	}

	m.ReplyText(c, msg, nil)
	return nil
}

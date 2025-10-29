/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package modules

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

func shuffleHandler(m *telegram.NewMessage) error {
	return handleShuffle(m, false)
}

func cshuffleHandler(m *telegram.NewMessage) error {
	return handleShuffle(m, true)
}

func handleShuffle(m *telegram.NewMessage, cplay bool) error {
	arg := strings.ToLower(m.Args())

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	if r.Track == nil {
		m.Reply("⚠️ No active playback found.")
		return telegram.EndGroup
	}

	r.Parse()

	if arg == "" {
		state := "disabled ❌"
		cmd := getCommand(m) + " on"
		if r.Shuffle {
			state = "enabled ✅"
			cmd = getCommand(m) + " off"
		}

		m.Reply(fmt.Sprintf(
			"🔀 Currently shuffle is <b>%s</b> for this chat.\n\nUse <code>%s</code> to toggle it.",
			state, cmd,
		))
		return telegram.EndGroup
	}

	var newState bool
	if arg == "on" || arg == "enable" || arg == "true" || arg == "1" {
		newState = true
	} else if arg == "off" || arg == "disable" || arg == "false" || arg == "0" {
		newState = false
	}

	r.SetShuffle(newState)

	state := "disabled ❌"
	if newState {
		state = "enabled ✅"
	}

	m.Reply(fmt.Sprintf("🔀 Shuffle <b>%s</b> by %s.", state, utils.MentionHTML(m.Sender)))
	return telegram.EndGroup
}

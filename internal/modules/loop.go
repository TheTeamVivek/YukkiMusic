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
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/utils"
)

func loopHandler(m *tg.NewMessage) error {
	return handleLoop(m, false)
}

func cloopHandler(m *tg.NewMessage) error {
	return handleLoop(m, true)
}

func handleLoop(m *tg.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.EndGroup
	}

	args := strings.Fields(m.Text())
	currentLoop := r.Loop

	if !r.IsActiveChat() {
		m.Reply(F(m.ChatID(), "room_no_active"))
		return tg.EndGroup
	}

	if len(args) < 2 {
		countLine := ""
		if currentLoop > 0 {
			countLine = "\n" + F(m.ChatID(), "loop_current", arg{
				"count": currentLoop,
			})
		}

		msg := F(m.ChatID(), "loop_usage", arg{
			"cmd":        getCommand(m),
			"count_line": countLine,
		})

		m.Reply(msg)
		return tg.EndGroup
	}

	newLoop, err := strconv.Atoi(args[1])
	if err != nil || newLoop < 0 || newLoop > 10 {
		m.Reply(F(m.ChatID(), "loop_invalid"))
		return tg.EndGroup
	}

	if newLoop == currentLoop {
		m.Reply(F(m.ChatID(), "loop_already_set", arg{
			"count": currentLoop,
		}))
		return tg.EndGroup
	}

	r.Lock()
	r.Loop = newLoop
	r.Unlock()

	mention := utils.MentionHTML(m.Sender)
	var msg string
	if newLoop == 0 {
		msg = F(m.ChatID(), "loop_disabled", arg{
			"user": mention,
		})
	} else {
		msg = F(m.ChatID(), "loop_set", arg{
			"count": newLoop,
			"user":  mention,
		})
	}

	m.Reply(msg)
	return tg.EndGroup
}

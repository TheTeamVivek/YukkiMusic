/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 * ________________________________________________________________________________________
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
 * ________________________________________________________________________________________
 */

package modules

import (
	"html"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var helpTexts = map[string]string{}

func checkForHelpFlag(m *tg.NewMessage) bool {
	text := strings.Fields(strings.ToLower(strings.TrimSpace(m.Text())))
	for _, t := range text {
		switch t {
		case "-h", "--h", "-help", "--help", "help":
			return true
		}
	}
	return false
}

func showHelpFor(m *tg.NewMessage, cmd string) error {
	help, ok := helpTexts[cmd]
	if !ok {
		alt := strings.TrimPrefix(cmd, "/")
		if h, ok := helpTexts[alt]; ok {
			help = h
		}
	}
	if help == "" {
		_, err := m.Reply(
			"⚠️ <i>No help found for command <code>" + html.EscapeString(
				cmd,
			) + "</code></i>",
		)
		if err != nil {
			return err
		}
		return tg.ErrEndGroup
	}
	_, err := m.Reply("📘 <b>Help for</b> <code>" + cmd + "</code>:\n\n" + help)
	if err != nil {
		return err
	}
	return tg.ErrEndGroup
}

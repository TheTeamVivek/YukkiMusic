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
"html"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/config"
)

var helpTexts = map[string]string{
	"/help": fmt.Sprintf(`ℹ️ <b>Help Command</b>
<i>Displays general bot help or detailed information about a specific command.</i>

<u>Usage:</u>
<code>/help</code> — Show the main help menu.
<code>/help &lt;command&gt;</code> — Show help for a specific command.

<b>💡 Tip:</b> You can view help for any command directly by adding a <code>-h</code> or <code>--help</code> flag, e.g. <code>/play -h</code>

For more info, visit our <a href="%s">Support Chat</a>.`, config.SupportChat),
}

func checkForHelpFlag(m *tg.NewMessage) bool {
	text := strings.ToLower(strings.TrimSpace(m.Text()))
	return strings.Contains(text, " --help") || strings.Contains(text, " -h") || strings.Contains(text, " help")
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
_, err := m.Reply("⚠️ <i>No help found for command <code>" + html.EscapeString(cmd) + "</code></i>")
		return eoe(err)
	}
	_, err := m.Reply("📘 <b>Help for</b> <code>" + cmd + "</code>:\n\n" + help)
	return eoe(err)
}

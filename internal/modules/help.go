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
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/internal/core"
)

const helpMessage = "📚 <b>Help Menu</b>\n\nSelect a category below to explore detailed commands and their usage for managing, controlling, and customizing the bot."

var (
	ownerDesc = "👑 <b>Owner Commands</b>\n<i>Exclusive tools for managing sudoers, executing system code, and performing bot administration securely.</i>"

	sudoDesc = "⚡ <b>Sudoer Commands</b>\n<i>Advanced tools for managing bot behavior, monitoring performance, and controlling logging and automation.</i>"

	adminDesc = "🛠 <b>Admin Commands</b>\n<i>Music and playback control tools for admins for handling music playback</i>"

	publicDesc = "🌍 <b>Public Commands</b>\n<i>Features for playing songs, viewing queues, checking latency, and reporting bugs.</i>"

	ownerCommands = `
<b>Commands:</b>
<b>/addsudo</b> - Add a new user to bot's sudolist
<b>/delsudo</b> - Remove a user from bot's sudolist
<b>/eval</b> - Execute Go code snippets
<b>/maintenance</b> — Manage the bot’s maintenance mode.
<b>/sh</b> - Run the shell commands.
`

	sudoCommands = `
<b>Commands:</b>
<b>/active</b> - Shows total active chats.
<b>/autoleave</b> - Toggle automatic chat leaving
<b>/logger</b> - Enable or disable logger
<b>/stats</b> - Display the bot & system Stats.
`

	adminCommands = `
<b>Commands:</b>
<b>/speed</b> - Change playback speed
<b>/skip</b> - Skip the current song
<b>/pause</b> - Pause the current playback
<b>/resume</b> - Resume paused playback
<b>/replay</b> - Replay the current song
<b>/mute</b> - Mute the playback
<b>/unmute</b> - Unmute the playback
<b>/seek</b> - Seek forward by few seconds
<b>/seekback</b> - Seek backward by few seconds
<b>/jump</b> - Jump to a given time in track
<b>/move</b> - Move a queued track to another position
<b>/clear</b> - Clear all songs from queue
<b>/remove</b> - Remove a specific track from queue
<b>/shuffle</b> - Shuffle all queued tracks
<b>/loop</b> - Enable or disable looping
<b>/stop</b> - Stop playback and leave VC
`

	publicCommands = `
<b>Commands:</b>
<b>/play</b> - Play a song
<b>/queue</b> - View all tracks currently queued
<b>/ping</b> - Check bot’s network latency
<b>/start</b> - Start the bot
<b>/help</b> - Show help menu
<b>/bug</b> - Report an issue or problem
<b>/position</b> - Show current track’s timestamp
<b>/reload</b> - Reload admin or cache data
<b>/json</b> - Show message JSON structure
<b>/sudolist</b> - View sudo user list
`
)

func helpHandler(m *tg.NewMessage) error {
	if m.ChatType() != tg.EntityUser {

		m.Reply("🤖 Hi! For bot help and commands, please DM me directly - I'm more responsive in private chats!", tg.SendOptions{ReplyMarkup: core.GetGroupHelpKeyboard()})
		return tg.EndGroup
	}

	m.Reply(helpMessage, tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard()})
	return tg.EndGroup
}

func helpCB(c *tg.CallbackQuery) error {
	c.Edit(helpMessage, &tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard()})
	c.Answer("")
	return tg.EndGroup
}

func helpCallbackHandler(c *tg.CallbackQuery) error {
	data := c.DataString()
	c.Answer("")
	if data == "" {
		return tg.EndGroup
	}

	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return tg.EndGroup
	}

	var text string
	btn := core.GetBackKeyboard()

	switch parts[1] {
	case "admins":
		text = adminDesc + "\n" + adminCommands
	case "sudoers":
		text = sudoDesc + "\n" + sudoCommands
	case "owner":
		text = ownerDesc + "\n" + ownerCommands
	case "public":
		text = publicDesc + "\n" + publicCommands
	case "main":
		text = helpMessage

		btn = core.GetHelpKeyboard()
	}

	c.Edit(text, &tg.SendOptions{ReplyMarkup: btn})
	return tg.EndGroup
}

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
	"html"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
)

var helpTexts = map[string]string{}

func init() {
	helpTexts["/help"] = fmt.Sprintf(`ℹ️ <b>Help Command</b>
<i>Displays general bot help or detailed information about a specific command.</i>

<u>Usage:</u>
<code>/help</code> — Show the main help menu.  
<code>/help &lt;command&gt;</code> — Show help for a specific command.

For more info, visit our <a href="%s">Support Chat</a>.`, config.SupportChat)
}

func helpHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	if len(args) > 1 {
		cmd := args[1]
		if cmd != "pm_help" {
			if !strings.HasPrefix(cmd, "/") {
				cmd = "/" + cmd
			}
			return showHelpFor(m, cmd)
		}
	}

	if m.ChatType() != tg.EntityUser {
		m.Reply(
			F(m.ChannelID(), "help_private_only"),
			&tg.SendOptions{
				ReplyMarkup: core.GetGroupHelpKeyboard(m.ChannelID()),
			},
		)
		return tg.ErrEndGroup
	}

	m.Reply(
		F(m.ChannelID(), "help_main"),
		&tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard(m.ChannelID())},
	)
	return tg.ErrEndGroup
}

func helpCB(c *tg.CallbackQuery) error {
	c.Edit(
		F(c.ChannelID(), "help_main"),
		&tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard(c.ChannelID())},
	)
	c.Answer("")
	return tg.ErrEndGroup
}

func helpCallbackHandler(c *tg.CallbackQuery) error {
	data := c.DataString()
	c.Answer("")
	if data == "" {
		return tg.ErrEndGroup
	}
	chatID := c.ChannelID()
	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return tg.ErrEndGroup
	}

	var text string
	btn := core.GetBackKeyboard(chatID)

	switch parts[1] {
	case "admins":
		text = F(chatID, "help_admin")
	case "sudoers":
		text = F(chatID, "help_sudo")
	case "owner":
		text = F(chatID, "help_owner")
	case "public":
		text = F(chatID, "help_public")
	case "main":
		text = F(chatID, "help_main")
		btn = core.GetHelpKeyboard(chatID)
	}

	c.Edit(text, &tg.SendOptions{ReplyMarkup: btn})
	return tg.ErrEndGroup
}

func showHelpFor(m *tg.NewMessage, cmd string) error {
	helpText, ok := helpTexts[cmd]
	if !ok {
		trimmed := strings.TrimPrefix(cmd, "/")
		if value, exists := helpTexts[trimmed]; exists {
			helpText = value
		}
	}

	if helpText == "" {
		_, err := m.Reply(
			"⚠️ <i>No help found for command <code>" +
				html.EscapeString(cmd) +
				"</code></i>",
		)
		if err != nil {
			return err
		}
		return tg.ErrEndGroup
	}

	_, err := m.Reply(
		"📘 <b>Help for</b> <code>" + cmd + "</code>:\n\n" + helpText,
	)
	if err != nil {
		return err
	}
	return tg.ErrEndGroup
}

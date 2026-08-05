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

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
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

func helpHandler(c *td.Client, m *td.Message) error {
	if isChannel(c, m) {
		return nil
	}

	args := strings.Fields(m.Text())
	if len(args) > 1 {
		cmd := args[1]
		if cmd != "pm_help" {
			if !strings.HasPrefix(cmd, "/") {
				cmd = "/" + cmd
			}
			return showHelpFor(c, m, cmd)
		}
	}

	if !m.IsPrivate() {
		m.ReplyText(c, F(m.ChatID(), "help_private_only"), &td.SendTextMessageOpts{
			ReplyMarkup: core.GetGroupHelpKeyboard(m.ChatID()),
		})
		return nil
	}

	_, err := m.ReplyText(c, F(m.ChatID(), "help_main"), &td.SendTextMessageOpts{
		ReplyMarkup: core.GetHelpKeyboard(m.ChatID()),
	})
	return err
}

func helpCB(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	cb.Answer(c, 0, false, "", "")
	_, err := cb.EditMessageText(c, F(cb.ChatId, "help_main"), &td.EditTextMessageOpts{
		ReplyMarkup: core.GetHelpKeyboard(cb.ChatId),
	})
	return err
}

func helpCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	data := cb.DataString()
	cb.Answer(c, 0, false, "", "")
	if data == "" {
		return nil
	}
	chatID := cb.ChatId
	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return nil
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

	_, err := cb.EditMessageText(c, text, &td.EditTextMessageOpts{ReplyMarkup: btn})
	return err
}

func showHelpFor(c *td.Client, m *td.Message, cmd string) error {
	helpText, ok := helpTexts[cmd]
	if !ok {
		trimmed := strings.TrimPrefix(cmd, "/")
		if value, exists := helpTexts[trimmed]; exists {
			helpText = value
		}
	}

	if helpText == "" {
		_, err := m.ReplyText(
			c,
			"⚠️ <i>No help found for command <code>"+
				html.EscapeString(cmd)+
				"</code></i>",
			nil,
		)
		return err
	}

	_, err := m.ReplyText(
		c,
		"📘 <b>Help for</b> <code>"+cmd+"</code>:\n\n"+helpText,
		nil,
	)
	return err
}

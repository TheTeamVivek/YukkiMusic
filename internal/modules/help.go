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
	"yukkimusic/internal/locales"
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
		_, err := m.ReplyText(c, F(m.ChatID(), "help_private_only"), &td.SendMessageOpts{
			ReplyMarkup: core.GetGroupHelpKeyboard(m.ChatID()),
		})
		return err
	}

	_, err := m.ReplyRichMessage(c, helpRich(m.ChatID(), "help_main"), &td.SendMessageOpts{
		ReplyMarkup: core.GetHelpKeyboard(m.ChatID()),
	})
	return err
}

// helpRich builds a rich help message for the given locale key. On the main
// help menu it links the support chat.
func helpRich(chatID int64, key string) *td.InputRichMessage {
	var args locales.Arg

	if key == "help_main" {
		support := config.SupportChat
		if support == "" {
			support = config.SupportChannel
		}
		if support == "" {
			support = "https://t.me/TheTeamVk"
		}
		args = locales.Arg{"support": html.EscapeString(support)}
	}

	text := F(chatID, key, args)

	return &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: text},
	}
}

func helpCB(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	cb.Answer(c, "", nil)

	msg, err := cb.GetMessage(c)
	if err != nil {
		return err
	}

	rich := &td.InputMessageRichMessage{Message: helpRich(cb.ChatId, "help_main")}
	_, err = c.EditContent(cb.ChatId, msg.Id, rich, &td.EditMessageMediaOpts{ReplyMarkup: core.GetHelpKeyboard(cb.ChatId)})
	return err
}

func helpCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	data := cb.DataString()
	cb.Answer(c, "", nil)
	if data == "" {
		return nil
	}
	chatID := cb.ChatId
	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return nil
	}

	key := ""
	btn := core.GetBackKeyboard(chatID)

	switch parts[1] {
	case "admins":
		key = "help_admin"
	case "sudoers":
		key = "help_sudo"
	case "owner":
		key = "help_owner"
	case "public":
		key = "help_public"
	case "main":
		key = "help_main"
		btn = core.GetHelpKeyboard(chatID)
	}

	msg, err := cb.GetMessage(c)
	if err != nil {
		return err
	}

	rich := &td.InputMessageRichMessage{Message: helpRich(chatID, key)}
	_, err = c.EditContent(cb.ChatId, msg.Id, rich, &td.EditMessageMediaOpts{ReplyMarkup: btn})
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

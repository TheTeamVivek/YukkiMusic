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

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
)

func helpHandler(m *tg.NewMessage) error {
	if m.ChatType() != tg.EntityUser {
		m.Reply(F(m.ChannelID(), "help_private_only"))
		return tg.EndGroup
	}
	m.Reply(F(m.ChannelID(), "help_main"), tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard()})
	return tg.EndGroup
}

func helpCB(c *tg.CallbackQuery) error {
	chatID, err := getCbChatID(c)
	if err != nil {
		gologging.ErrorF("getCbChatID error %v", err)
		c.Answer("⚠️ Chat not recognized.", &tg.CallbackOptions{Alert: true})
		return tg.EndGroup
	}
	c.Edit(F(chatID, "help_main"), &tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard()})
	c.Answer("")
	return tg.EndGroup
}

func helpCallbackHandler(c *tg.CallbackQuery) error {
	data := c.DataString()
	c.Answer("")
	if data == "" {
		return tg.EndGroup
	}
	chatID, err := getCbChatID(c)
	if err != nil {
		gologging.ErrorF("getCbChatID error %v", err)
		c.Answer(FWithLang(config.DefaultLang, "chat_not_recognized"), &tg.CallbackOptions{Alert: true})
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
		text = F(chatID, "help_admin")
	case "sudoers":
		text = F(chatID, "help_sudo")
	case "owner":
		text = F(chatID, "help_owner")
	case "public":
		text = F(chatID, "help_public")
	case "main":
		text = F(chatID, "help_main")
		btn = core.GetHelpKeyboard()
	}

	c.Edit(text, &tg.SendOptions{ReplyMarkup: btn})
	return tg.EndGroup
}

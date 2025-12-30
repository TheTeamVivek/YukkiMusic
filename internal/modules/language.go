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

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func langHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	lang, err := database.GetChatLanguage(chatID)
	if err != nil {
		lang = config.DefaultLang
	}

	kb := telegram.NewKeyboard()
	var btns []telegram.KeyboardButton
	for _, l := range locales.GetAvailableLanguages() {
		name := locales.Get(l, "name", nil)
		if l == lang {
			name = "✔️ " + name
		}
		btns = append(btns, telegram.Button.Data(name, "lang:"+l))
	}
	kb.NewColumn(2, btns...)

	_, err = m.Reply(F(chatID, "lang_select"), &telegram.SendOptions{ReplyMarkup: kb.Build()})
	if err != nil {
		return err
	}
	return telegram.ErrEndGroup
}

func langCallbackHandler(cb *telegram.CallbackQuery) error {
	data := cb.DataString()
	opt := &telegram.CallbackOptions{Alert: true}
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		cb.Answer("⚠️ Invalid data.", opt)
		return telegram.ErrEndGroup
	}
	lang := parts[1]

	chatID := cb.ChannelID()
	if isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); err != nil || !isAdmin {
		cb.Answer(F(chatID, "only_admin_or_auth_cb"), opt)
		return telegram.ErrEndGroup
	}

	currentLang, _ := database.GetChatLanguage(chatID)

	if lang == currentLang {
		cb.Answer(F(chatID, "lang_same"), opt)
		return telegram.ErrEndGroup
	}

	langName := locales.Get(lang, "name", nil)

	if err := database.SetChatLanguage(chatID, lang); err != nil {
		logger.ErrorF("SetChatLanguage error: %v", err)
		cb.Answer(F(chatID, "lang_fail"), opt)
		return telegram.ErrEndGroup
	}

	msg := F(chatID, "lang_success", locales.Arg{"lang_name": langName})
	cb.Answer(msg, opt)
	cb.Edit(msg)
	return telegram.ErrEndGroup
}

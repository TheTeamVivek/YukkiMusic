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
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func langHandler(m *telegram.NewMessage) error {
	return showLangMenu(m, false)
}

func showLangMenu(m interface{}, isCallback bool) error {
	var chatID int64
	if isCallback {
		cb := m.(*telegram.CallbackQuery)
		chatID = cb.ChannelID()
	} else {
		msg := m.(*telegram.NewMessage)
		chatID = msg.ChannelID()
	}

	lang, err := database.Language(chatID)
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

	if isCallback {
		kb.AddRow(telegram.Button.Data(F(chatID, "BACK_BTN"), "set:main"))
	}

	text := F(chatID, "lang_select")
	if isCallback {
		cb := m.(*telegram.CallbackQuery)
		cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: kb.Build()})
	} else {
		msg := m.(*telegram.NewMessage)
		msg.Reply(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: kb.Build()})
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
	if lang == "select" {
		return showLangMenu(cb, true)
	}

	if isAdmin, _ := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); !isAdmin {
		cb.Answer(F(chatID, "only_admin_cb"), opt)
		return telegram.ErrEndGroup
	}

	currentLang, _ := database.Language(chatID)

	if lang == currentLang {
		cb.Answer(F(chatID, "lang_same"), opt)
		return telegram.ErrEndGroup
	}

	langName := locales.Get(lang, "name", nil)

	if err := database.SetLanguage(chatID, lang); err != nil {
		gologging.ErrorF("SetChatLanguage error: %v", err)
		cb.Answer(F(chatID, "lang_fail"), opt)
		return telegram.ErrEndGroup
	}

	msg := F(chatID, "lang_success", locales.Arg{"lang_name": langName})
	cb.Answer(msg, opt)
	cb.Edit(msg)
	return telegram.ErrEndGroup
}

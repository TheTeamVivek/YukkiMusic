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

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func langHandler(c *td.Client, m *td.Message) error {
	return showLangMenu(c, m)
}

func showLangMenu(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	lang, err := database.Language(chatID)
	if err != nil {
		lang = config.DefaultLang
	}

	markup := buildLangMarkup(chatID, lang, false)

	_, err = m.ReplyText(c, F(chatID, "lang_select"), &td.SendTextMessageOpts{
		ParseMode:   td.ParseModeHTML,
		ReplyMarkup: markup,
	})
	return err
}

func buildLangMarkup(chatID int64, lang string, isCallback bool) *td.ReplyMarkupInlineKeyboard {
	var rows [][]td.InlineKeyboardButton
	var row []td.InlineKeyboardButton

	for _, l := range locales.GetAvailableLanguages() {
		name := locales.Get(l, "name", nil)
		if l == lang {
			name = "✔️ " + name
		}
		row = append(row, td.InlineKeyboardButton{
			Text: name,
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("lang:" + l)},
		})
		if len(row) == 2 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	if isCallback {
		rows = append(rows, []td.InlineKeyboardButton{
			{
				Text: F(chatID, "BACK_BTN"),
				Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:main")},
			},
		})
	}

	return &td.ReplyMarkupInlineKeyboard{Rows: rows}
}

func langCallbackHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	data := u.DataString()
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		_ = u.Answer(c, 0, true, "⚠️ Invalid data.", "")
		return nil
	}
	lang := parts[1]

	chatID := u.ChatId
	if lang == "select" {
		cur, err := database.Language(chatID)
		if err != nil {
			cur = config.DefaultLang
		}
		markup := buildLangMarkup(chatID, cur, true)
		_, _ = u.EditMessageText(c, F(chatID, "lang_select"), &td.EditTextMessageOpts{
			ParseMode:   td.ParseModeHTML,
			ReplyMarkup: markup,
		})
		return nil
	}

	if isAdmin, _ := utils.IsChatAdmin(c, chatID, u.SenderUserId); !isAdmin {
		_ = u.Answer(c, 0, true, F(chatID, "only_admin_cb"), "")
		return nil
	}

	currentLang, _ := database.Language(chatID)

	if lang == currentLang {
		_ = u.Answer(c, 0, true, F(chatID, "lang_same"), "")
		return nil
	}

	langName := locales.Get(lang, "name", nil)

	if err := database.SetLanguage(chatID, lang); err != nil {
		logger.Errorf("SetChatLanguage error: %v", err)
		_ = u.Answer(c, 0, true, F(chatID, "lang_fail"), "")
		return nil
	}

	msg := F(chatID, "lang_success", locales.Arg{"lang_name": langName})
	_ = u.Answer(c, 0, true, msg, "")
	_, _ = u.EditMessageText(c, msg, &td.EditTextMessageOpts{ParseMode: td.ParseModeHTML})
	return nil
}

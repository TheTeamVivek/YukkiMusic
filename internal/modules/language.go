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

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func langHandler(c *td.Client, m *td.Message) error {
	if isChannel(c, m) {
		return nil
	}
	if m.IsPrivate() {
		return showLangMenu(c, m)
	}
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	return showLangMenu(c, m)
}

func showLangMenu(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	lang, err := database.Language(chatID)
	if err != nil {
		lang = config.DefaultLang
	}

	back := ""
	if m.IsPrivate() {
		back = "start"
	}

	markup := core.GetLangKeyboard(lang, back)

	_, err = m.ReplyText(c, F(chatID, "lang_select"), &td.SendMessageOpts{
		ParseMode:   td.ParseModeHTML,
		ReplyMarkup: markup,
	})
	return err
}

func langCallbackHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	data := u.DataString()
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		_ = u.Answer(c, "⚠️ Invalid data.", &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}
	lang := parts[1]

	chatID := u.ChatId
	if chatID > 0 {
		return privateLangCallback(c, u, lang)
	}

	if isAdmin, _ := utils.IsChatAdmin(c, chatID, u.SenderUserId); !isAdmin {
		_ = u.Answer(c, F(chatID, "only_admin_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if lang == "select" {
		cur, err := database.Language(chatID)
		if err != nil {
			cur = config.DefaultLang
		}
		markup := core.GetLangKeyboard(cur, "set:main")
		_, _ = c.EditRichMessageHTML(chatID, u.MessageId, F(chatID, "lang_select"), &td.EditRichTextOpts{
			ReplyMarkup: markup,
		})
		return nil
	}

	currentLang, _ := database.Language(chatID)

	if lang == currentLang {
		_ = u.Answer(c, F(chatID, "lang_same"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	langName := locales.Get(lang, "name", nil)

	if err := database.SetLanguage(chatID, lang); err != nil {
		utils.Log.Errorf("SetChatLanguage error: %v", err)
		_ = u.Answer(c, F(chatID, "lang_fail"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	msg := F(chatID, "lang_success", locales.Arg{"lang_name": langName})
	_ = u.Answer(c, msg, &td.CallbackAnswerOpts{ShowAlert: true})
	_, _ = c.EditRichMessageHTML(chatID, u.MessageId, msg, nil)
	return nil
}

func privateLangCallback(c *td.Client, u *td.UpdateNewCallbackQuery, lang string) error {
	chatID := u.ChatId

	if lang == "select" {
		cur, err := database.Language(chatID)
		if err != nil {
			cur = config.DefaultLang
		}
		markup := core.GetLangKeyboard(cur, "start")
		_, _ = c.EditRichMessageHTML(chatID, u.MessageId, F(chatID, "lang_select"), &td.EditRichTextOpts{
			ReplyMarkup: markup,
		})
		return nil
	}

	currentLang, _ := database.Language(chatID)

	if lang == currentLang {
		_ = u.Answer(c, F(chatID, "lang_same"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	langName := locales.Get(lang, "name", nil)

	if err := database.SetLanguage(chatID, lang); err != nil {
		utils.Log.Errorf("SetChatLanguage error: %v", err)
		_ = u.Answer(c, F(chatID, "lang_fail"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	_ = u.Answer(c, F(chatID, "lang_success", locales.Arg{"lang_name": langName}), &td.CallbackAnswerOpts{ShowAlert: true})

	user, _ := c.GetUser(u.SenderUserId)
	if msg, err := u.GetMessage(c); err == nil {
		rich := &td.InputMessageRichMessage{Message: startRich(c, chatID, user, u.SenderUserId)}
		_, _ = c.EditContent(chatID, msg.Id, rich, &td.EditMessageMediaOpts{ReplyMarkup: core.GetStartMarkup(chatID)})
	}
	return nil
}

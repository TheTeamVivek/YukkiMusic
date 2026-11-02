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
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

// dashboardDef describes a toggle-able setting rendered by the shared dashboard.
type dashboardDef struct {
	title       func(chatID int64) string
	description func(chatID int64) string
	canUse      func(c *td.Client, u *td.UpdateNewCallbackQuery) bool
	get         func(chatID int64) (bool, error)
	set         func(chatID int64, on bool) error
}

// dashboards maps a callback name to the setting it controls. Each toggle
// command registers its own entry from its init (see autoleave.go, autoplay.go).
var dashboards = map[string]dashboardDef{}

// dashboardRich renders the shared dashboard card: the given title and
// description followed by the current on/off state.
func dashboardRich(
	chatID int64,
	title, description string,
	enabled bool,
) *td.InputRichMessage {
	var b strings.Builder

	fmt.Fprintf(&b, "<h3>%s</h3>\n", title)
	fmt.Fprintf(&b, "<aside><i>%s</i></aside>\n", description)
	fmt.Fprintf(&b, "<p>%s</p>\n", F(chatID, "dashboard_status", locales.Arg{
		"status": F(chatID, utils.IfElse(enabled, "dashboard_enabled", "dashboard_disabled")),
	}))

	return &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: b.String()},
	}
}

// dashboardMarkup builds the toggle button and the close button for a dashboard.
func dashboardMarkup(chatID int64, name string, enabled bool) td.ReplyMarkup {
	lang, err := database.Language(chatID)
	if err != nil {
		lang = config.DefaultLang
	}
	label, iconID := locales.GetButton(lang, utils.IfElse(enabled, "dashboard_enabled", "dashboard_disabled"), nil)

	toggle := td.InlineKeyboardButton{
		Text: label,
		Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("dashboard:" + name)},
	}
	toggle.IconCustomEmojiId = iconID

	closeBtn := td.InlineKeyboardButton{
		Text: F(chatID, "CLOSE_BTN"),
		Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("close")},
	}

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{closeBtn, toggle}},
	}
}

// sendDashboard replies with the dashboard card for the given setting.
func sendDashboard(c *td.Client, m *td.Message, name string) error {
	def, ok := dashboards[name]
	if !ok {
		return nil
	}

	chatID := m.ChatID()

	enabled, err := def.get(chatID)
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "dashboard_fail"), nil)
		return err
	}

	_, err = m.ReplyRichMessage(
		c,
		dashboardRich(chatID, def.title(chatID), def.description(chatID), enabled),
		&td.SendMessageOpts{ReplyMarkup: dashboardMarkup(chatID, name, enabled)},
	)
	return err
}

// dashboardCallback toggles the setting behind a dashboard button and refreshes
// the card in place.
func dashboardCallback(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId
	name := strings.TrimPrefix(u.DataString(), "dashboard:")

	def, ok := dashboards[name]
	if !ok {
		utils.Log.Warnf("Unknown dashboard callback: %s", u.DataString())
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if !def.canUse(c, u) {
		return nil
	}

	enabled, err := def.get(chatID)
	if err != nil {
		u.Answer(c, F(chatID, "dashboard_fail_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	enabled = !enabled
	if err := def.set(chatID, enabled); err != nil {
		u.Answer(c, F(chatID, "dashboard_fail_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if _, err := c.EditContent(u.ChatId, u.MessageId, &td.InputMessageRichMessage{
		Message: dashboardRich(
			chatID,
			def.title(chatID),
			def.description(chatID),
			enabled,
		),
	}, &td.EditMessageMediaOpts{ReplyMarkup: dashboardMarkup(chatID, name, enabled)}); err != nil {
		utils.Log.Errorf("Failed to refresh dashboard %q: %v", name, err)
	}

	u.Answer(c, F(chatID, "dashboard_toggled_cb", locales.Arg{
		"status": F(chatID, utils.IfElse(enabled, "enabled", "disabled")),
	}), &td.CallbackAnswerOpts{ShowAlert: true})
	return nil
}

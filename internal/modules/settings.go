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

	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/playmode"] = `<i>Control who can use the /play command in this chat.</i>

<u>Usage:</u>
<b>/playmode [enable|disable]</b> — Set play mode restriction

<b>⚙️ Options:</b>
• <b>enable</b> — Only admins and authorized users can play
• <b>disable</b> — Everyone can play (default)`

	cmdDeleteHelp := `<i>Toggle automatic deletion of bot commands in this chat.</i>

<u>Usage:</u>
<b>/cmddelete [enable|disable]</b> — Set command deletion status

<b>⚙️ Options:</b>
• <b>enable</b> — Commands will be deleted after being handled
• <b>disable</b> — Commands will remain in the chat (default)`

	helpTexts["/cmddelete"] = cmdDeleteHelp
	helpTexts["/commanddelete"] = cmdDeleteHelp

	cleanModeHelp := `<i>Enable timed cleanup for command/service messages in this chat.</i>

<u>Usage:</u>
<b>/cleanmode [enable|disable]</b> — Set clean mode status

<b>⚙️ Behavior:</b>
• <b>enable</b> — Bot replies and command messages are auto-deleted after a short delay
• <b>disable</b> — Keep messages in chat (default)`

	helpTexts["/cleanmode"] = cleanModeHelp

	helpTexts["/adminmode"] = `<i>Control who can use admin-level music commands in this chat.</i>

<u>Usage:</u>
<b>/adminmode [admin|adminauth|everyone]</b> — Set admin command access

<b>⚙️ Options:</b>
• <b>admin</b> — Only chat admins can use admin commands
• <b>adminauth</b> — Chat admins + authorized users can use admin commands (default)
• <b>everyone</b> — Everyone can use admin commands`
}

func playmodeHandler(c *td.Client, m *td.Message) error {
	args := strings.Fields(m.Text())
	chatID := m.ChatID()

	current, err := database.PlayModeAdminsOnly(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		statusKey := "playmode_status_everyone"
		if current {
			statusKey = "playmode_status_admins"
		}

		_, _ = m.ReplyText(c, F(chatID, "playmode_help", locales.Arg{
			"status": F(chatID, statusKey),
		}), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
		return nil
	}

	adminsOnly, err := utils.ParseBool(args[1])
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return nil
	}

	if err := database.SetPlayModeAdminsOnly(chatID, adminsOnly); err != nil {
		return err
	}

	successKey := "playmode_success_everyone"
	if adminsOnly {
		successKey = "playmode_success_admins"
	}

	_, _ = m.ReplyText(c, F(chatID, successKey), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
	return nil
}

func cmdDeleteHandler(c *td.Client, m *td.Message) error {
	args := strings.Fields(m.Text())
	chatID := m.ChatID()
	cmd := getCommand(m)

	current, err := database.CommandDelete(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		actionKey := "disabled"
		if current {
			actionKey = "enabled"
		}

		_, _ = m.ReplyText(c, F(chatID, "cmddelete_status", locales.Arg{
			"cmd":    cmd,
			"action": F(chatID, actionKey),
		}), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
		return nil
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return nil
	}

	if err := database.SetCommandDelete(chatID, enabled); err != nil {
		return err
	}

	actionKey := "disabled"
	if enabled {
		actionKey = "enabled"
	}

	_, _ = m.ReplyText(c, F(chatID, "cmddelete_updated", locales.Arg{
		"action": F(chatID, actionKey),
	}), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
	return nil
}

func cleanModeHandler(c *td.Client, m *td.Message) error {
	args := strings.Fields(m.Text())
	chatID := m.ChatID()

	current, err := database.CleanMode(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		_, _ = m.ReplyText(
			c,
			cleanModeStatusText(chatID, current)+"\n\n"+F(chatID, "cleanmode_hint"),
			&td.SendTextMessageOpts{ParseMode: td.ParseModeHTML},
		)
		return nil
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return nil
	}

	if err := database.SetCleanMode(chatID, enabled); err != nil {
		return err
	}
	if !enabled {
		cleanScheduler.cancel(chatID)
	}

	_, _ = m.ReplyText(c, cleanModeStatusText(chatID, enabled)+"\n\n"+F(chatID, "cleanmode_hint"), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
	return nil
}

func adminModeHandler(c *td.Client, m *td.Message) error {
	args := strings.Fields(m.Text())
	chatID := m.ChatID()

	current, err := database.GetAdminMode(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		_, _ = m.ReplyText(c, F(chatID, "adminmode_help", locales.Arg{
			"status": F(chatID, adminModeStatusKey(current)),
		}), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
		return nil
	}

	mode, ok := parseAdminMode(args[1])
	if !ok {
		_, _ = m.ReplyText(c, F(chatID, "adminmode_invalid"), nil)
		return nil
	}

	if err := database.SetAdminMode(chatID, mode); err != nil {
		return err
	}

	_, _ = m.ReplyText(c, F(chatID, "adminmode_updated", locales.Arg{
		"status": F(chatID, adminModeStatusKey(mode)),
	}), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML})
	return nil
}

func adminModeStatusKey(mode database.AdminMode) string {
	switch mode {
	case database.AdminModeAdminsOnly:
		return "adminmode_status_admin"
	case database.AdminModeEveryone:
		return "adminmode_status_everyone"
	default:
		return "adminmode_status_adminauth"
	}
}

func parseAdminMode(input string) (database.AdminMode, bool) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "admin", "admins", "adminonly", "adminsonly", "admins_only":
		return database.AdminModeAdminsOnly, true
	case "adminauth", "auth", "admin+auth", "dj", "admin_auth":
		return database.AdminModeAdminAuth, true
	case "everyone", "all":
		return database.AdminModeEveryone, true
	default:
		return "", false
	}
}

func settingsHandler(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()
	settings, err := database.GetChatSettings(chatID)
	if err != nil {
		return err
	}

	title := "Chat"
	if chat, err := c.GetChat(chatID); err == nil && chat != nil && chat.Title != "" {
		title = chat.Title
	}

	kb := buildSettingsMarkup(chatID, settings)
	_, err = m.ReplyText(c, F(chatID, "settings_main", locales.Arg{
		"title": title,
		"id":    chatID,
	}), &td.SendTextMessageOpts{ParseMode: td.ParseModeHTML, ReplyMarkup: kb})
	return err
}

func settingsCallbackHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId
	data := u.DataString()
	parts := strings.Split(data, ":")
	title := "Chat"
	if chat, err := c.GetChat(chatID); err == nil && chat != nil && chat.Title != "" {
		title = chat.Title
	}

	if len(parts) < 2 {
		return nil
	}

	// Check permissions
	if isAdmin, err := utils.IsChatAdmin(c, chatID, u.SenderUserId); err != nil ||
		!isAdmin {
		_ = u.Answer(c, 0, true, F(chatID, "only_admin_cb"), "")
		return nil
	}

	settings, err := database.GetChatSettings(chatID)
	if err != nil {
		return nil
	}

	action := parts[1]
	if strings.HasPrefix(data, "info:") {
		_ = u.Answer(c, 0, true, F(chatID, "settings_info_"+action), "")
		return nil
	}
	if action == "main" {
		kb := buildSettingsMarkup(chatID, settings)
		_, _ = u.EditMessageText(c, F(chatID, "settings_main", locales.Arg{
			"title": title,
			"id":    chatID,
		}), &td.EditTextMessageOpts{ParseMode: td.ParseModeHTML, ReplyMarkup: kb})
		return nil
	}
	switch action {
	case "playmode":
		settings.PlayModeAdminsOnly = !settings.PlayModeAdminsOnly
	case "adminmode":
		switch settings.AdminMode {
		case database.AdminModeAdminsOnly:
			settings.AdminMode = database.AdminModeAdminAuth
		case database.AdminModeAdminAuth:
			settings.AdminMode = database.AdminModeEveryone
		default:
			settings.AdminMode = database.AdminModeAdminsOnly
		}
	case "cmddelete":
		settings.CommandDelete = !settings.CommandDelete
	case "cleanmode":
		settings.CleanMode = !settings.CleanMode
		if !settings.CleanMode {
			cleanScheduler.cancel(chatID)
		}
	case "cleanduration":
		next := cleanModeDurationOptions[0]
		for i, v := range cleanModeDurationOptions {
			if v == settings.CleanModeDurationMins {
				next = cleanModeDurationOptions[(i+1)%len(cleanModeDurationOptions)]
				break
			}
		}
		settings.CleanModeDurationMins = next
	case "nothumb":
		settings.ThumbnailsDisabled = !settings.ThumbnailsDisabled
	}

	if err := database.UpdateChatSettings(settings); err != nil {
		return nil
	}

	_ = u.Answer(c, 0, false, F(chatID, "settings_updated"), "")
	kb := buildSettingsMarkup(chatID, settings)

	_, _ = u.EditMessageText(c, F(chatID, "settings_main", locales.Arg{
		"title": title,
		"id":    chatID,
	}), &td.EditTextMessageOpts{ParseMode: td.ParseModeHTML, ReplyMarkup: kb})
	return nil
}

func buildSettingsMarkup(chatID int64, s *database.ChatSettings) *td.ReplyMarkupInlineKeyboard {
	rows := make([][]td.InlineKeyboardButton, 0, 6)

	// Admin Mode
	adminModeStatus := F(chatID, adminModeStatusKey(s.AdminMode))
	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_adminmode"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:adminmode")},
		},
		{
			Text: adminModeStatus,
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:adminmode")},
		},
	})

	// Play Mode
	playModeStatus := F(
		chatID,
		utils.IfElse(
			s.PlayModeAdminsOnly,
			"playmode_status_admins",
			"playmode_status_everyone",
		),
	)

	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_playmode"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:playmode")},
		},
		{
			Text: playModeStatus,
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:playmode")},
		},
	})

	// Cmd Delete
	cmdDeleteStatus := utils.IfElse(s.CommandDelete, "enabled", "disabled")
	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_cmddelete"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:cmddelete")},
		},
		{
			Text: F(chatID, cmdDeleteStatus),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:cmddelete")},
		},
	})

	// Clean Mode
	cleanModeStatus := utils.IfElse(s.CleanMode, "enabled", "disabled")
	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_cleanmode"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:cleanmode")},
		},
		{
			Text: F(chatID, cleanModeStatus),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:cleanmode")},
		},
	})

	cleanDuration := s.CleanModeDurationMins
	if cleanDuration <= 0 {
		cleanDuration = 15
	}
	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_cleanduration"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:cleanduration")},
		},
		{
			Text: utils.IntToStr(cleanDuration) + "m",
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:cleanduration")},
		},
	})

	// Thumbnails
	thumbStatus := utils.IfElse(!s.ThumbnailsDisabled, "enabled", "disabled")

	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_nothumb"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:nothumb")},
		},
		{
			Text: F(chatID, thumbStatus),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("set:nothumb")},
		},
	})

	// Language
	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "settings_btn_lang"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("info:lang")},
		},
		{
			Text: F(chatID, "name"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("lang:select")},
		},
	})

	rows = append(rows, []td.InlineKeyboardButton{
		{
			Text: F(chatID, "CLOSE_BTN"),
			Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("close")},
		},
	})

	return &td.ReplyMarkupInlineKeyboard{Rows: rows}
}

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

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
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

	helpTexts["/adminmode"] = `<i>Control who can use admin-level music commands in this chat.</i>

<u>Usage:</u>
<b>/adminmode [admin|adminauth|everyone]</b> — Set admin command access

<b>⚙️ Options:</b>
• <b>admin</b> — Only chat admins can use admin commands
• <b>adminauth</b> — Chat admins + authorized users can use admin commands (default)
• <b>everyone</b> — Everyone can use admin commands`
}

func playmodeHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	current, err := database.PlayModeAdminsOnly(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		statusKey := "playmode_status_everyone"
		if current {
			statusKey = "playmode_status_admins"
		}

		m.Reply(F(chatID, "playmode_help", locales.Arg{
			"status": F(chatID, statusKey),
		}), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	adminsOnly, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if err := database.SetPlayModeAdminsOnly(chatID, adminsOnly); err != nil {
		return err
	}

	successKey := "playmode_success_everyone"
	if adminsOnly {
		successKey = "playmode_success_admins"
	}

	m.Reply(F(chatID, successKey), &tg.SendOptions{ParseMode: "HTML"})
	return tg.ErrEndGroup
}

func cmdDeleteHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()
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

		m.Reply(F(chatID, "cmddelete_status", locales.Arg{
			"cmd":    cmd,
			"action": F(chatID, actionKey),
		}), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if err := database.SetCommandDelete(chatID, enabled); err != nil {
		return err
	}

	actionKey := "disabled"
	if enabled {
		actionKey = "enabled"
	}

	m.Reply(F(chatID, "cmddelete_updated", locales.Arg{
		"action": F(chatID, actionKey),
	}), &tg.SendOptions{ParseMode: "HTML"})
	return tg.ErrEndGroup
}

func adminModeHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	current, err := database.GetAdminMode(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		m.Reply(F(chatID, "adminmode_help", locales.Arg{
			"status": F(chatID, adminModeStatusKey(current)),
		}), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	mode, ok := parseAdminMode(args[1])
	if !ok {
		m.Reply(F(chatID, "adminmode_invalid"))
		return tg.ErrEndGroup
	}

	if err := database.SetAdminMode(chatID, mode); err != nil {
		return err
	}

	m.Reply(F(chatID, "adminmode_updated", locales.Arg{
		"status": F(chatID, adminModeStatusKey(mode)),
	}), &tg.SendOptions{ParseMode: "HTML"})
	return tg.ErrEndGroup
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

func settingsHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()
	settings, err := database.GetChatSettings(chatID)
	if err != nil {
		return err
	}

	title := "Chat"
	if m.Channel != nil {
		title = m.Channel.Title
	}

	kb := buildSettingsMarkup(chatID, settings)
	_, err = m.Reply(F(chatID, "settings_main", locales.Arg{
		"title": title,
		"id":    chatID,
	}), &tg.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func settingsCallbackHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	data := cb.DataString()
	parts := strings.Split(data, ":")
	title := "Chat"
	if cb.Channel != nil {
		title = cb.Channel.Title
	}

	if len(parts) < 2 {
		return nil
	}

	// Check permissions
	if isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); err != nil ||
		!isAdmin {
		cb.Answer(F(chatID, "only_admin_cb"), &tg.CallbackOptions{Alert: true})
		return nil
	}

	settings, err := database.GetChatSettings(chatID)
	if err != nil {
		return err
	}

	action := parts[1]
	if strings.HasPrefix(data, "info:") {
		cb.Answer(F(chatID, "settings_info_"+action), &tg.CallbackOptions{Alert: true})
		return nil
	}
	if action == "main" {
		kb := buildSettingsMarkup(chatID, settings)
		cb.Edit(F(chatID, "settings_main", locales.Arg{
			"title": title,
			"id":    chatID,
		}), &tg.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
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
	case "nothumb":
		settings.ThumbnailsDisabled = !settings.ThumbnailsDisabled
	}

	if err := database.UpdateChatSettings(settings); err != nil {
		return err
	}

	cb.Answer(F(chatID, "settings_updated"))
	kb := buildSettingsMarkup(chatID, settings)

	cb.Edit(F(chatID, "settings_main", locales.Arg{
		"title": title,
		"id":    chatID,
	}), &tg.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
	return nil
}

func buildSettingsMarkup(chatID int64, s *database.ChatSettings) *tg.ReplyInlineMarkup {
	kb := tg.NewKeyboard()


	// Admin Mode
	adminModeStatus := F(chatID, adminModeStatusKey(s.AdminMode))
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_adminmode"), "info:adminmode"),
		tg.Button.Data(adminModeStatus, "set:adminmode"),
	)

	// Cmd Delete
	cmdDeleteStatus := utils.IfElse(s.CommandDelete, "enabled", "disabled")
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_cmddelete"), "info:cmddelete"),
		tg.Button.Data(F(chatID, cmdDeleteStatus), "set:cmddelete"),
	)


	// Play Mode
playModeStatus := F(chatID, utils.IfElse(s.PlayModeAdminsOnly,"playmode_status_admins","playmode_status_everyone" ))

	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_playmode"), "info:playmode"),
		tg.Button.Data(playModeStatus, "set:playmode"),
	)

	// Thumbnails
	thumbStatus := utils.IfElse(!s.ThumbnailsDisabled, "enabled", "disabled")

	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_nothumb"), "info:nothumb"),
		tg.Button.Data(F(chatID, thumbStatus), "set:nothumb"),
	)

	// Language
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_lang"), "info:lang"),
		tg.Button.Data(F(chatID, "name"), "lang:select"),
	)

	kb.AddRow(tg.Button.Data(F(chatID, "CLOSE_BTN"), "close"))

	return kb.Build()
}

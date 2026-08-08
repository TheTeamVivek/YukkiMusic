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
	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/database"
)

func isSuperGroupTd(c *td.Client, m *td.Message) bool {
	if m.IsPrivate() {
		m.ReplyText(c, F(m.ChatID(), "only_supergroup"), nil)
		database.AddServedUser(m.ChatID())
		return false
	}

	chat, err := m.GetChat(c)
	if err != nil {
		return false
	}

	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	if !ok || sg.IsChannel {
		return false
	}

	database.AddServedChat(m.ChatID())
	return true
}

func filterAuthUsersTd(c *td.Client, m *td.Message) bool {
	if canUseAdminCommandTd(c, m.ChatID(), m.SenderID()) {
		return true
	}

	mode, err := database.GetAdminMode(m.ChatID())
	if err == nil && mode == database.AdminModeAdminsOnly {
		m.ReplyText(c, F(m.ChatID(), "only_admin"), nil)
	} else {
		m.ReplyText(c, F(m.ChatID(), "only_admin_or_auth"), nil)
	}
	return false
}

func canUseAdminCommandTd(c *td.Client, chatID, userID int64) bool {
	if isOwnerOrSudo(userID) {
		return true
	}

	mode, err := database.GetAdminMode(chatID)
	if err != nil {
		mode = database.AdminModeAdminAuth
	}

	if mode == database.AdminModeEveryone {
		return true
	}

	isAdmin, err := tdIsChatAdmin(c, chatID, userID)
	if err == nil && isAdmin {
		return true
	}

	if mode == database.AdminModeAdminsOnly {
		return false
	}

	isAuth, err := database.IsAuthorized(chatID, userID)
	return err == nil && isAuth
}

func tdIsChatAdmin(c *td.Client, chatID, userID int64) (bool, error) {
	if chatID == userID {
		return true, nil
	}

	admins, err := c.GetChatAdministrators(chatID)
	if err != nil {
		return false, err
	}

	for _, a := range admins.Administrators {
		if a.UserId == userID {
			return true, nil
		}
	}
	return false, nil
}

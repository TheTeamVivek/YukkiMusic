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

	"main/internal/config"
	"main/internal/database"
	"main/internal/utils"
)

var (
	superGroupFilter    = tg.CustomFilter(filterSuperGroup)
	adminFilter         = tg.CustomFilter(filterChatAdmins)
	authFilter          = tg.CustomFilter(filterAuthUsers)
	ignoreChannelFilter = tg.CustomFilter(filterChannel)
	sudoOnlyFilter      = tg.CustomFilter(filterSudo)
	ownerFilter         = tg.CustomFilter(filterOwner)
)

func filterSuperGroup(m *tg.NewMessage) bool {
	if !filterChannel(m) {
		return false
	}

	switch m.ChatType() {
	case tg.EntityChat:
		// EntityChat can be basic group or supergroup — allow only supergroup
		if m.Channel != nil && !m.Channel.Broadcast {
			database.AddServedChat(m.ChannelID())
			return true // Supergroup
		}
		warnAndLeave(m.Client, m.ChannelID()) // Basic group → leave
		database.RemoveServedChat(m.ChannelID())
		return false

	case tg.EntityChannel:
		return false // Pure channel chat → ignore

	case tg.EntityUser:
		m.Reply(F(m.ChannelID(), "only_supergroup"))
		database.AddServedUser(m.ChannelID())
		return false // Private chat → warn
	}

	return false
}

func filterChatAdmins(m *tg.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err != nil || !isAdmin {
		m.Reply(F(m.ChannelID(), "only_admin"))
		return false
	}
	return true
}

func filterAuthUsers(m *tg.NewMessage) bool {
	if canUseAdminCommand(m.Client, m.ChannelID(), m.SenderID()) {
		return true
	}

	mode, err := database.GetAdminMode(m.ChannelID())
	if err == nil && mode == database.AdminModeAdminsOnly {
		m.Reply(F(m.ChannelID(), "only_admin"))
	} else {
		m.Reply(F(m.ChannelID(), "only_admin_or_auth"))
	}
	return false
}

func filterSudo(m *tg.NewMessage) bool {
	is, _ := database.IsSudo(m.SenderID())

	if config.OwnerID == 0 || (m.SenderID() != config.OwnerID && !is) {
		if m.IsPrivate() ||
			strings.HasSuffix(m.GetCommand(), m.Client.Me().Username) {
			m.Reply(F(m.ChannelID(), "only_sudo"))
		}
		return false
	}

	return true
}

func filterChannel(m *tg.NewMessage) bool {
	if _, ok := m.Message.FromID.(*tg.PeerChannel); ok {
		return false
	}
	return true
}

func canUseAdminCommand(c *tg.Client, chatID, userID int64) bool {
	mode, err := database.GetAdminMode(chatID)
	if err != nil {
		mode = database.AdminModeAdminAuth
	}

	if mode == database.AdminModeEveryone {
		return true
	}

	isAdmin, err := utils.IsChatAdmin(c, chatID, userID)
	if err == nil && isAdmin {
		return true
	}

	if mode == database.AdminModeAdminsOnly {
		return false
	}

	isAuth, err := database.IsAuthorized(chatID, userID)
	return err == nil && isAuth
}

func filterOwner(m *tg.NewMessage) bool {
	if config.OwnerID == 0 || m.SenderID() != config.OwnerID {
		if m.IsPrivate() ||
			strings.HasSuffix(m.GetCommand(), m.Client.Me().Username) {
			m.Reply(F(m.ChannelID(), "only_owner"))
		}
		return false
	}
	return true
}

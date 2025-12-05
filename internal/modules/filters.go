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

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/utils"
)

var (
	superGroupFilter    = tg.FilterFunc(filterSuperGroup)
	adminFilter         = tg.FilterFunc(filterChatAdmins)
	authFilter          = tg.FilterFunc(filterAuthUsers)
	ignoreChannelFilter = tg.FilterFunc(filterChannel)
	sudoOnlyFilter      = tg.FilterFunc(filterSudo)
	ownerFilter         = tg.FilterFunc(filterOwner)
)

func filterSuperGroup(m *tg.NewMessage) bool {
	if !filterChannel(m) {
		return false
	}

	switch m.ChatType() {
	case tg.EntityChat:
		// EntityChat can be basic group or supergroup — allow only supergroup
		if m.Channel != nil && !m.Channel.Broadcast {
			database.AddServed(m.ChannelID())
			return true // Supergroup
		}
		warnAndLeave(m.Client, m.ChatID()) // Basic group → leave
		database.DeleteServed(m.ChannelID())
		return false

	case tg.EntityChannel:
		return false // Pure channel chat → ignore

	case tg.EntityUser:
		m.Reply(F(m.ChannelID(), "only_supergroup"))
		database.AddServed(m.ChannelID(), true)
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
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err == nil && isAdmin {
		return true
	}

	isAuth, err := database.IsAuthUser(m.ChannelID(), m.SenderID())
	if err == nil && isAuth {
		return true
	}

	m.Reply(F(m.ChannelID(), "only_admin_or_auth"))
	return false
}

func filterSudo(m *tg.NewMessage) bool {
	is, _ := database.IsSudo(m.SenderID())

	if config.OwnerID == 0 || (m.SenderID() != config.OwnerID && !is) {
		if m.IsPrivate() || strings.HasSuffix(m.GetCommand(), core.BUser.Username) {
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

func filterOwner(m *tg.NewMessage) bool {
	if config.OwnerID == 0 || m.SenderID() != config.OwnerID {
		if m.IsPrivate() || strings.HasSuffix(m.GetCommand(), core.BUser.Username) {
			m.Reply(F(m.ChannelID(), "only_owner"))
		}
		return false
	}
	return true
}

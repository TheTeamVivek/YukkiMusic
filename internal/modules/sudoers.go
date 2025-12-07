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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func handleAddSudo(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	// No args + no reply -> ask for user
	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.EndGroup
	}

	// Extract target user
	targetID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	// Owner trying to self
	if targetID == config.OwnerID {
		m.Reply(F(chatID, "sudo_owner_self"))
		return telegram.EndGroup
	}

	// Trying to add the bot itself
	if targetID == core.BUser.ID {
		m.Reply(F(chatID, "sudo_bot_self"))
		return telegram.EndGroup
	}

	// Trying to add the assitant
	if ass, err := core.Assistants.ForChat(chatID); err == nil {
		if targetID == ass.User.ID {
			m.Reply(F(chatID, "sudo_assistant_self"))
			return telegram.EndGroup
		}
	}

	// Fetch user info
	user, err := m.Client.GetUser(targetID)
	if err != nil {
		m.Reply(F(chatID, "sudo_fetch_user_fail", locales.Arg{
			"error": err.Error(),
		}))
		gologging.Error("Failed to get user: " + err.Error())
		return telegram.EndGroup
	}

	// Bots cannot be sudoers
	if user.Bot {
		m.Reply(F(chatID, "sudo_bot_user"))
		return telegram.EndGroup
	}

	// Username / mention
	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname = "@" + user.Username
	}
	idStr := strconv.FormatInt(targetID, 10)

	// Check if already sudo
	exists, err := database.IsSudo(targetID)
	if err != nil {
		m.Reply(F(chatID, "sudo_check_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	if exists {
		m.Reply(F(chatID, "sudo_already", locales.Arg{
			"user": uname,
			"id":   idStr,
		}))
		return telegram.EndGroup
	}

	// Add to sudo
	if err := database.AddSudo(targetID); err != nil {
		m.Reply(F(chatID, "sudo_add_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	m.Reply(F(chatID, "sudo_add_success", locales.Arg{
		"user": uname,
		"id":   idStr,
	}))

	if config.SetCmds {
		// Update commands for this sudo user
		sudoCommands := append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...)

		if _, err := m.Client.BotsSetBotCommands(
			&telegram.BotCommandScopePeer{
				Peer: &telegram.InputPeerUser{UserID: targetID, AccessHash: 0},
			},
			"",
			sudoCommands,
		); err != nil {
			gologging.Error("Failed to set PrivateSudoCommands " + err.Error())
		}
	}

	return telegram.EndGroup
}

func handleDelSudo(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	// No args + no reply -> ask for user
	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.EndGroup
	}

	// Extract target user
	targetID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	// Cannot remove owner
	if targetID == config.OwnerID {
		m.Reply(F(chatID, "sudo_owner_remove_block"))
		return telegram.EndGroup
	}

	// Cannot remove assistant (UbUser)
	if ass, err := core.Assistants.ForChat(chatID); err == nil {
		if targetID == ass.User.ID {
			m.Reply(F(chatID, "sudo_assistant_cannot_remove"))
			return telegram.EndGroup
		}
	}
	// Fetch user info
	user, err := m.Client.GetUser(targetID)
	if err != nil {
		m.Reply(F(chatID, "sudo_fetch_user_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname = "@" + user.Username
	}
	idStr := strconv.FormatInt(targetID, 10)

	// Check if sudo
	exists, err := database.IsSudo(targetID)
	if err != nil {
		m.Reply(F(chatID, "sudo_check_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	if !exists {
		m.Reply(F(chatID, "sudo_not_exists", locales.Arg{
			"user": uname,
			"id":   idStr,
		}))
		return telegram.EndGroup
	}

	// Reset that user's bot commands
	if _, err := m.Client.BotsResetBotCommands(
		&telegram.BotCommandScopePeer{
			Peer: &telegram.InputPeerUser{UserID: targetID, AccessHash: 0},
		},
		"",
	); err != nil {
		gologging.Error("Failed to reset sudo commands: " + err.Error())
	}

	// Delete from DB
	if err := database.DeleteSudo(targetID); err != nil {
		m.Reply(F(chatID, "sudo_remove_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	// Success
	m.Reply(F(chatID, "sudo_remove_success", locales.Arg{
		"user": uname,
		"id":   idStr,
	}))

	return telegram.EndGroup
}

func handleGetSudoers(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	floodKey := fmt.Sprintf("sudoers:%d%d", chatID, m.SenderID())
	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		m.Reply(F(chatID, "flood_seconds", locales.Arg{
			"duration": int(remaining.Seconds()),
		}))
		return telegram.EndGroup
	}
	utils.SetFlood(floodKey, 30*time.Second)

	// "⏳ Fetching sudoers list..."
	mystic, _ := m.Reply(F(chatID, "sudo_list_fetching"))

	list, err := database.GetSudoers()
	if err != nil {
		utils.EOR(mystic, F(chatID, "sudo_list_fetch_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.EndGroup
	}

	var sb strings.Builder

	// Header
	sb.WriteString(F(chatID, "sudo_list_header"))
	sb.WriteString("\n\n")

	// First, show owner
	ownerID := config.OwnerID
	ownerIDStr := strconv.FormatInt(ownerID, 10)

	ownerStr := "<code>" + ownerIDStr + "</code>"
	if user, err := m.Client.GetUser(ownerID); err == nil {
		if user.Username != "" {
			ownerStr = "@" + user.Username + " (ID: <code>" + ownerIDStr + "</code>)"
		} else {
			ownerStr = utils.MentionHTML(user) + " (ID: <code>" + ownerIDStr + "</code>)"
		}
	}

	sb.WriteString(F(chatID, "sudo_list_owner", locales.Arg{
		"index": 1,
		"user":  ownerStr,
	}))
	sb.WriteString("\n")

	// Then list other sudoers
	idx := 2
	for _, id := range list {
		if id == ownerID {
			continue // skip owner since already listed
		}

		idStr := strconv.FormatInt(id, 10)
		userStr := "<code>" + idStr + "</code>" // fallback

		if user, err := m.Client.GetUser(id); err == nil {
			if user.Username != "" {
				userStr = "@" + user.Username + " (ID: <code>" + idStr + "</code>)"
			} else {
				userStr = utils.MentionHTML(user) + " (ID: <code>" + idStr + "</code>)"
			}
		}

		sb.WriteString(F(chatID, "sudo_list_entry", locales.Arg{
			"index": idx,
			"user":  userStr,
		}))
		sb.WriteString("\n")
		idx++
		time.Sleep(1 * time.Second)
	}

	if idx == 2 {
		// no extra sudoers beyond owner
		sb.WriteString(F(chatID, "sudo_list_no_extra"))
		sb.WriteString("\n")
	}

	utils.EOR(mystic, sb.String())
	return telegram.EndGroup
}

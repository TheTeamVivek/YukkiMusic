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
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func handleBlockUser(m *telegram.NewMessage) error {
	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(m.ChannelID(), "auth_no_user", locales.Arg{"cmd": getCommand(m)}))
		return telegram.ErrEndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(m.ChannelID(), "user_extract_fail", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	if err := database.AddBlacklistedUser(userID); err != nil {
		m.Reply("Failed to block user: " + err.Error())
		return telegram.ErrEndGroup
	}
	m.Reply(F(m.ChannelID(), "blacklist_block_user_success", locales.Arg{"id": userID}))
	return telegram.ErrEndGroup
}

func handleUnblockUser(m *telegram.NewMessage) error {
	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(m.ChannelID(), "auth_no_user", locales.Arg{"cmd": getCommand(m)}))
		return telegram.ErrEndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(m.ChannelID(), "user_extract_fail", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	if err := database.RemoveBlacklistedUser(userID); err != nil {
		m.Reply(F(m.ChannelID(), "blacklist_unblock_user_fail", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	m.Reply(F(m.ChannelID(), "blacklist_unblock_user_success", locales.Arg{"id": userID}))
	return telegram.ErrEndGroup
}

func handleBlockChat(m *telegram.NewMessage) error {
	if m.Args() == "" {
		m.Reply(F(m.ChannelID(), "blacklist_usage_blockchat"))
		return telegram.ErrEndGroup
	}
	chatID, err := utils.ExtractChat(m)
	if err != nil {
		m.Reply(F(m.ChannelID(), "blacklist_invalid_chat_identifier", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	if err := database.AddBlacklistedChat(chatID); err != nil {
		m.Reply("Failed to block chat: " + err.Error())
		return telegram.ErrEndGroup
	}
	m.Reply(F(m.ChannelID(), "blacklist_block_chat_success", locales.Arg{"id": chatID}))
	return telegram.ErrEndGroup
}

func handleUnblockChat(m *telegram.NewMessage) error {
	if m.Args() == "" {
		m.Reply(F(m.ChannelID(), "blacklist_usage_unblockchat"))
		return telegram.ErrEndGroup
	}
	chatID, err := utils.ExtractChat(m)
	if err != nil {
		m.Reply(F(m.ChannelID(), "blacklist_invalid_chat_identifier", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	if err := database.RemoveBlacklistedChat(chatID); err != nil {
		m.Reply(F(m.ChannelID(), "blacklist_unblock_chat_fail", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	m.Reply(F(m.ChannelID(), "blacklist_unblock_chat_success", locales.Arg{"id": chatID}))
	return telegram.ErrEndGroup
}

func handleBlacklisted(m *telegram.NewMessage) error {
	chats, err := database.BlacklistedChats()
	if err != nil {
		m.Reply(F(m.ChannelID(), "blacklist_fetch_chats_fail", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}
	users, err := database.BlacklistedUsers()
	if err != nil {
		m.Reply(F(m.ChannelID(), "blacklist_fetch_users_fail", locales.Arg{"error": err.Error()}))
		return telegram.ErrEndGroup
	}

	var b strings.Builder

	b.WriteString(F(m.ChannelID(), "blacklist_list_title"))
	b.WriteString("\n\n")

	b.WriteString(F(m.ChannelID(), "blacklist_list_chats"))
	b.WriteString("\n")

	if len(chats) == 0 {
		b.WriteString("• None\n")
	} else {
		for i, id := range chats {
			b.WriteString(strconv.Itoa(i+1) + ". <code>" + strconv.FormatInt(id, 10) + "</code>\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(F(m.ChannelID(), "blacklist_list_users"))
	b.WriteString("\n")

	if len(users) == 0 {
		b.WriteString("• None")
	} else {
		for i, id := range users {
			b.WriteString(strconv.Itoa(i+1) + ". <code>" + strconv.FormatInt(id, 10) + "</code>\n")
		}
	}

	m.Reply(b.String())
	return telegram.ErrEndGroup
}

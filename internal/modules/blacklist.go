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

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func handleBlockUser(c *td.Client, m *td.Message) error {
	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, _ = m.ReplyText(c, F(m.ChatID(), "auth_no_user", locales.Arg{"cmd": getCommand(m)}), nil)
		return nil
	}

	userID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, F(m.ChatID(), "user_extract_fail", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	if err := database.AddBlacklistedUser(userID); err != nil {
		_, _ = m.ReplyText(c, "Failed to block user: "+err.Error(), nil)
		return nil
	}
	_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_block_user_success", locales.Arg{"id": userID}), nil)
	return nil
}

func handleUnblockUser(c *td.Client, m *td.Message) error {
	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, _ = m.ReplyText(c, F(m.ChatID(), "auth_no_user", locales.Arg{"cmd": getCommand(m)}), nil)
		return nil
	}

	userID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, F(m.ChatID(), "user_extract_fail", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	if err := database.RemoveBlacklistedUser(userID); err != nil {
		_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_user_fail", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_user_success", locales.Arg{"id": userID}), nil)
	return nil
}

func handleBlockChat(c *td.Client, m *td.Message) error {
	if m.Args() == "" {
		_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_usage_blockchat"), nil)
		return nil
	}
	chatID, err := utils.ExtractChat(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_invalid_chat_identifier", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	if err := database.AddBlacklistedChat(chatID); err != nil {
		_, _ = m.ReplyText(c, "Failed to block chat: "+err.Error(), nil)
		return nil
	}
	_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_block_chat_success", locales.Arg{"id": chatID}), nil)
	return nil
}

func handleUnblockChat(c *td.Client, m *td.Message) error {
	if m.Args() == "" {
		_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_usage_unblockchat"), nil)
		return nil
	}
	chatID, err := utils.ExtractChat(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_invalid_chat_identifier", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	if err := database.RemoveBlacklistedChat(chatID); err != nil {
		_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_chat_fail", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	_, _ = m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_chat_success", locales.Arg{"id": chatID}), nil)
	return nil
}

func handleBlacklisted(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()
	chats, err := database.BlacklistedChats()
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "blacklist_fetch_chats_fail", locales.Arg{"error": err.Error()}), nil)
		return nil
	}
	users, err := database.BlacklistedUsers()
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "blacklist_fetch_users_fail", locales.Arg{"error": err.Error()}), nil)
		return nil
	}

	var b strings.Builder

	b.WriteString(F(chatID, "blacklist_list_title"))
	b.WriteString("\n\n")

	b.WriteString(F(chatID, "blacklist_list_chats"))
	b.WriteString("\n")

	if len(chats) == 0 {
		b.WriteString("• None\n")
	} else {
		for i, id := range chats {
			b.WriteString(strconv.Itoa(i+1) + ". <code>" + strconv.FormatInt(id, 10) + "</code>\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(F(chatID, "blacklist_list_users"))
	b.WriteString("\n")

	if len(users) == 0 {
		b.WriteString("• None")
	} else {
		for i, id := range users {
			b.WriteString(strconv.Itoa(i+1) + ". <code>" + strconv.FormatInt(id, 10) + "</code>\n")
		}
	}

	_, _ = m.ReplyText(c, b.String(), nil)
	return nil
}

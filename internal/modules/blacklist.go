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
	if !checkOwner(c, m) {
		return nil
	}
	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, err := m.ReplyText(c, F(m.ChatID(), "auth_no_user", locales.Arg{"cmd": getCommand(m)}), nil)
		return err
	}

	userID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, err := m.ReplyText(c, F(m.ChatID(), "user_extract_fail", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	if err := database.AddBlacklistedUser(userID); err != nil {
		_, err := m.ReplyText(c, "Failed to block user: "+err.Error(), nil)
		return err
	}
	_, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_block_user_success", locales.Arg{"id": userID}), nil)
	return rerr
}

func handleUnblockUser(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, err := m.ReplyText(c, F(m.ChatID(), "auth_no_user", locales.Arg{"cmd": getCommand(m)}), nil)
		return err
	}

	userID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, err := m.ReplyText(c, F(m.ChatID(), "user_extract_fail", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	if err := database.RemoveBlacklistedUser(userID); err != nil {
		_, err := m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_user_fail", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	_, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_user_success", locales.Arg{"id": userID}), nil)
	return rerr
}

func handleBlockChat(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	if m.Args() == "" {
		_, err := m.ReplyText(c, F(m.ChatID(), "blacklist_usage_blockchat"), nil)
		return err
	}
	chatID, err := utils.ExtractChat(c, m)
	if err != nil {
		_, err := m.ReplyText(c, F(m.ChatID(), "blacklist_invalid_chat_identifier", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	if err := database.AddBlacklistedChat(chatID); err != nil {
		_, err := m.ReplyText(c, "Failed to block chat: "+err.Error(), nil)
		return err
	}
	_, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_block_chat_success", locales.Arg{"id": chatID}), nil)
	return rerr
}

func handleUnblockChat(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	if m.Args() == "" {
		_, err := m.ReplyText(c, F(m.ChatID(), "blacklist_usage_unblockchat"), nil)
		return err
	}
	chatID, err := utils.ExtractChat(c, m)
	if err != nil {
		_, err := m.ReplyText(c, F(m.ChatID(), "blacklist_invalid_chat_identifier", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	if err := database.RemoveBlacklistedChat(chatID); err != nil {
		_, err := m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_chat_fail", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	_, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_unblock_chat_success", locales.Arg{"id": chatID}), nil)
	return rerr
}

func handleBlacklisted(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	chatID := m.ChatID()
	chats, err := database.BlacklistedChats()
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "blacklist_fetch_chats_fail", locales.Arg{"error": err.Error()}), nil)
		return err
	}
	users, err := database.BlacklistedUsers()
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "blacklist_fetch_users_fail", locales.Arg{"error": err.Error()}), nil)
		return err
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
	_, rerr := m.ReplyText(c, b.String(), nil)
	return rerr
}

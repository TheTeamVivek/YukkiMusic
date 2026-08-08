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
	"strconv"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"
	"github.com/amarnathcjd/gogram/telegram"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func handleAddSudo(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	// No args + no reply -> ask for user
	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, _ = m.ReplyText(c, F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return nil
	}

	// Extract target user
	targetID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	// Owner trying to self
	if targetID == config.OwnerID {
		_, _ = m.ReplyText(c, F(chatID, "sudo_owner_self"), nil)
		return nil
	}

	// Trying to add the bot itself
	if targetID == c.Me.Id {
		_, _ = m.ReplyText(c, F(chatID, "sudo_bot_self"), nil)
		return nil
	}

	// Fetch user info
	user, err := c.GetUser(targetID)
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "sudo_fetch_user_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		logger.Error("Failed to get user: " + err.Error())
		return nil
	}

	// Bots cannot be sudoers
	if user.Type != nil {
		if _, isBot := user.Type.(*td.UserTypeBot); isBot {
			_, _ = m.ReplyText(c, F(chatID, "sudo_bot_user"), nil)
			return nil
		}
	}

	// Username / mention
	uname := mentionOf(user, targetID)
	if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
		uname = "@" + user.Usernames.ActiveUsernames[0]
	}
	idStr := strconv.FormatInt(targetID, 10)

	// Check if already sudo
	exists, err := database.IsSudo(targetID)
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "sudo_check_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	if exists {
		_, _ = m.ReplyText(c, F(chatID, "sudo_already", locales.Arg{
			"user": uname,
			"id":   idStr,
		}), nil)
		return nil
	}

	// Add to sudo
	if err := database.AddSudo(targetID); err != nil {
		_, _ = m.ReplyText(c, F(chatID, "sudo_add_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	_, _ = m.ReplyText(c, F(chatID, "sudo_add_success", locales.Arg{
		"user": uname,
		"id":   idStr,
	}), nil)

	if config.SetCmds {
		// Update commands for this sudo user
		sudoCommands := append(
			AllCommands.PrivateUserCommands,
			AllCommands.PrivateSudoCommands...,
		)

		if _, err := core.Bot.BotsSetBotCommands(
			&telegram.BotCommandScopePeer{
				Peer: &telegram.InputPeerUser{UserID: targetID, AccessHash: 0},
			},
			"",
			sudoCommands,
		); err != nil {
			logger.Error("Failed to set PrivateSudoCommands " + err.Error())
		}
	}

	return nil
}

func handleDelSudo(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	// No args + no reply -> ask for user
	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, _ = m.ReplyText(c, F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return nil
	}

	// Extract target user
	targetID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	// Cannot remove owner
	if targetID == config.OwnerID {
		_, _ = m.ReplyText(c, F(chatID, "sudo_owner_remove_block"), nil)
		return nil
	}

	// Fetch user info
	user, _ := c.GetUser(targetID)

	var uname string
	if user != nil {
		uname = mentionOf(user, user.Id)
		if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
			uname = "@" + user.Usernames.ActiveUsernames[0]
		}
	} else {
		uname = "User"
	}
	idStr := strconv.FormatInt(targetID, 10)

	// Check if sudo
	exists, err := database.IsSudo(targetID)
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "sudo_check_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	if !exists {
		_, _ = m.ReplyText(c, F(chatID, "sudo_not_exists", locales.Arg{
			"user": uname,
			"id":   idStr,
		}), nil)
		return nil
	}

	// Reset that user's bot commands
	if _, err := core.Bot.BotsResetBotCommands(
		&telegram.BotCommandScopePeer{
			Peer: &telegram.InputPeerUser{UserID: targetID, AccessHash: 0},
		},
		"",
	); err != nil {
		logger.Error("Failed to reset sudo commands: " + err.Error())
	}

	// Delete from DB
	if err := database.RemoveSudo(targetID); err != nil {
		_, _ = m.ReplyText(c, F(chatID, "sudo_remove_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	// Success
	_, _ = m.ReplyText(c, F(chatID, "sudo_remove_success", locales.Arg{
		"user": uname,
		"id":   idStr,
	}), nil)

	return nil
}

func handleGetSudoers(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	floodKey := fmt.Sprintf("sudoers:%d%d", chatID, m.SenderID())
	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		_, _ = m.ReplyText(c, F(chatID, "flood_seconds", locales.Arg{
			"duration": int(remaining.Seconds()),
		}), nil)
		return nil
	}
	utils.SetFlood(floodKey, 30*time.Second)

	// "⏳ Fetching sudoers list..."
	statusMsg, _ := m.ReplyText(c, F(chatID, "sudo_list_fetching"), nil)

	list, err := database.Sudoers()
	if err != nil {
		_, _ = utils.EOR(c, statusMsg, F(chatID, "sudo_list_fetch_fail", locales.Arg{
			"error": err.Error(),
		}), &td.EditTextMessageOpts{ParseMode: td.ParseModeHTML})
		return nil
	}

	var sb strings.Builder

	// Header
	sb.WriteString(F(chatID, "sudo_list_header"))
	sb.WriteString("\n\n")

	// First, show owner
	ownerID := config.OwnerID
	ownerIDStr := strconv.FormatInt(ownerID, 10)

	ownerStr := "<code>" + ownerIDStr + "</code>"
	if user, err := c.GetUser(ownerID); err == nil {
		if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
			ownerStr = "@" + user.Usernames.ActiveUsernames[0] + " (ID: <code>" + ownerIDStr + "</code>)"
		} else {
			ownerStr = mentionOf(user, ownerID) + " (ID: <code>" + ownerIDStr + "</code>)"
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

		if user, err := c.GetUser(id); err == nil {
			if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
				userStr = "@" + user.Usernames.ActiveUsernames[0] + " (ID: <code>" + idStr + "</code>)"
			} else {
				userStr = mentionOf(user, id) + " (ID: <code>" + idStr + "</code>)"
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

	_, _ = utils.EOR(c, statusMsg, sb.String(), &td.EditTextMessageOpts{
		ParseMode: td.ParseModeHTML,
	})
	return nil
}

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
	"time"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func actionFilter(m *td.Message) bool {
	switch m.Content.(type) {
	case *td.MessageChatAddMembers,
		*td.MessageChatDeleteMember,
		*td.MessageVideoChatStarted,
		*td.MessageVideoChatEnded:
		return true
	}
	return false
}

func handleActions(c *td.Client, m *td.Message) error {
	if !isValidChatType(c, m) {
		warnAndLeave(c, m.ChatID())
		return nil
	}

	switch m.Content.(type) {
	case *td.MessageChatAddMembers:
		return handleChatMemberAdd(c, m)
	case *td.MessageChatDeleteMember:
		return handleChatMemberDelete(c, m)
	case *td.MessageVideoChatStarted, *td.MessageVideoChatEnded:
		return handleVoiceChatAction(c, m)
	}

	return nil
}

func handleChatMemberAdd(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	if c.Me == nil {
		return nil
	}
	content := m.Content.(*td.MessageChatAddMembers)

	for _, uid := range content.MemberUserIds {
		if uid != c.Me.Id {
			continue
		}

		if blockedChat, _ := database.IsBlacklistedChat(chatID); blockedChat && !isOwnerOrSudo(m.SenderID()) {
			m.ReplyText(c, F(chatID, "blacklist_chat_blocked"), nil)
			leaveChat(c, chatID)
			return nil
		}

		ownerID, err := utils.GetChatOwner(c, chatID)
		if err == nil {
			if blockedOwner, _ := database.IsBlacklistedUser(ownerID); blockedOwner && !isOwnerOrSudo(m.SenderID()) {
				m.ReplyText(c, F(chatID, "blacklist_owner_blocked_leave"), nil)
				leaveChat(c, chatID)
				return nil
			}
		}

		logger.Debug("Bot added to " + utils.IntToStr(chatID))
		m.ReplyText(c, F(chatID, "bot_added_normal"), nil)
		database.AddServedChat(chatID)

		if config.LoggerID != 0 {
			c.SendTextMessage(config.LoggerID, F(config.LoggerID, "logger_bot_added", buildLogArgs(c, m, chatID, "added")), nil)
		}

		return nil
	}

	return nil
}

func handleChatMemberDelete(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	if c.Me == nil {
		return nil
	}
	content := m.Content.(*td.MessageChatDeleteMember)

	if content.UserId == c.Me.Id {
		logger.Debug("Bot removed from " + utils.IntToStr(chatID))

		cleanScheduler.cancel(chatID)
		core.DeleteRoom(chatID)
		core.DeleteChatState(chatID)
		database.RemoveServedChat(chatID)

		if config.LoggerID != 0 {
			c.SendTextMessage(config.LoggerID, F(config.LoggerID, "logger_bot_removed", buildLogArgs(c, m, chatID, "removed")), nil)
		}
	}

	return nil
}

func handleVoiceChatAction(c *td.Client, m *td.Message) error {
	if isMaint, _ := database.IsMaintenanceEnabled(); isMaint {
		return nil
	}

	chatID := m.ChatID()
	isActive := true
	var duration int32

	if ended, ok := m.Content.(*td.MessageVideoChatEnded); ok {
		isActive = false
		duration = ended.Duration
	}

	s, err := core.GetChatState(chatID)
	if err != nil {
		logger.Errorf("Failed to get chat state for %d: %v", chatID, err)
		return nil
	}

	s.SetVoiceChatActive(isActive)

	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")
	c.SendTextMessage(chatID, F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(duration))}), nil)
	logger.Debugf("Voice chat %s in %d", msgKey, chatID)

	if !isActive {
		room, ok := core.GetRoom(chatID, nil, false)
		go func() {
			time.Sleep(500 * time.Millisecond)
			if ok {
				scheduleOldPlayingMessage(room)
			}
			core.DeleteRoom(chatID)
			logger.Debugf("Room destroyed for ended voice chat in %d", chatID)
		}()
	}

	return nil
}

func isValidChatType(c *td.Client, m *td.Message) bool {
	chat, err := m.GetChat(c)
	if err != nil {
		return false
	}

	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	return ok && !sg.IsChannel
}

func buildLogArgs(c *td.Client, m *td.Message, chatID int64, action string) locales.Arg {
	groupName := "N/A"
	if chat, err := m.GetChat(c); err == nil && chat != nil {
		groupName = chat.Title
	}

	sender, _ := m.GetUser(c)
	actorID := m.SenderID()
	actorName := "N/A"
	if sender != nil {
		if full := strings.TrimSpace(sender.FirstName + " " + sender.LastName); full != "" {
			actorName = full
		}
	}

	return locales.Arg{
		"group_name":            groupName,
		"group_id":              chatID,
		"group_username":        "N/A",
		action + "_by_name":     actorName,
		action + "_by_id":       actorID,
		action + "_by_username": mentionOf(sender, actorID),
		"date_time":             time.Now().Format("02 Jan 2006 • 15:04"),
	}
}

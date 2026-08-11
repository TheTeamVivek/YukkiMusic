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

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

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
	chat, err := m.GetChat(c)
	if err != nil {
		return nil
	}
	switch t := chat.Type.(type) {
	case *td.ChatTypeSupergroup:
		if t.IsChannel {
			return nil
		}
	case *td.ChatTypePrivate, *td.ChatTypeSecret:
		return nil
	default:
		text := F(m.ChatID(), "supergroup_needed", locales.Arg{
			"chat_id":       m.ChatID(),
			"support_group": config.SupportChat,
		})
		if _, err := c.SendTextMessage(m.ChatID(), text, &td.SendTextMessageOpts{DisableWebPagePreview: true}); err != nil {
			logger.Errorf("failed to send supergroup conversion message to chat %d: %v", m.ChatID(), err)
			return nil
		}
		go func() {
			leaveChat(c, m.ChatID())
		}()
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
			if _, rerr := m.ReplyText(c, F(chatID, "blacklist_chat_blocked"), nil); rerr != nil {
				logger.Error(rerr)
			}
			leaveChat(c, chatID)
			return nil
		}

		ownerID, err := utils.GetChatOwner(c, chatID)
		if err == nil {
			if blockedOwner, _ := database.IsBlacklistedUser(ownerID); blockedOwner && !isOwnerOrSudo(m.SenderID()) {
				if _, rerr := m.ReplyText(c, F(chatID, "blacklist_owner_blocked_leave"), nil); rerr != nil {
					logger.Error(rerr)
				}
				leaveChat(c, chatID)
				return nil
			}
		}

		logger.Debug("Bot added to " + utils.IntToStr(chatID))
		if _, rerr := m.ReplyText(c, F(chatID, "bot_added_normal"), nil); rerr != nil {
			return rerr
		}
		database.AddServedChat(chatID)

		if config.LoggerID != 0 {
			if _, rerr := c.SendTextMessage(config.LoggerID, F(config.LoggerID, "logger_bot_added", buildLogArgs(c, m, chatID, "added")), nil); rerr != nil {
				return rerr
			}
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
			if _, rerr := c.SendTextMessage(config.LoggerID, F(config.LoggerID, "logger_bot_removed", buildLogArgs(c, m, chatID, "removed")), nil); rerr != nil {
				return rerr
			}
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
	if _, rerr := c.SendTextMessage(chatID, F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(duration))}), nil); rerr != nil {
		logger.Error(rerr)
	}
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

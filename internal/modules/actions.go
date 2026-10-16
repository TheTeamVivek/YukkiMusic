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

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func actionFilter(m *td.Message) bool {
	switch m.Content.(type) {
	case *td.MessageChatAddMembers,
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
		if _, err := c.SendMessage(m.ChatID(), text, &td.SendMessageOpts{DisableWebPagePreview: true}); err != nil {
			utils.Log.Errorf("failed to send supergroup conversion message to chat %d: %v", m.ChatID(), err)
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
				utils.Log.Error(rerr.Error())
			}
			leaveChat(c, chatID)
			return nil
		}

		ownerID, err := utils.GetChatOwner(c, chatID)
		if err == nil {
			if blockedOwner, _ := database.IsBlacklistedUser(ownerID); blockedOwner && !isOwnerOrSudo(m.SenderID()) {
				if _, rerr := m.ReplyText(c, F(chatID, "blacklist_owner_blocked_leave"), nil); rerr != nil {
					utils.Log.Error(rerr.Error())
				}
				leaveChat(c, chatID)
				return nil
			}
		}

		utils.Log.Debug("Bot added to " + utils.IntToStr(chatID))
		if _, rerr := m.ReplyText(c, F(chatID, "bot_added_normal"), nil); rerr != nil {
			return rerr
		}
		database.AddServedChat(chatID)

		if config.LoggerID != 0 {
			if _, rerr := c.SendMessage(config.LoggerID, F(config.LoggerID, "logger_bot_added", buildLogArgs(c, m, chatID, "added")), nil); rerr != nil {
				return rerr
			}
		}

		return nil
	}

	return nil
}

func handleBotRemoved(c *td.Client, u *td.UpdateChatMember, chatID int64) {
	utils.Log.Debugf("Bot removed from chat %d", chatID)

	cleanScheduler.cancel(chatID)
	core.Rooms.Delete(chatID)
	core.DropChatState(chatID)
	database.RemoveServedChat(chatID)

	if config.LoggerID == 0 || config.LoggerID == chatID {
		return
	}

	groupName := "N/A"
	if chat, err := c.GetChat(chatID); err == nil && chat != nil {
		groupName = chat.Title
	}

	actorID := u.ActorUserId
	actorName := "N/A"
	actorMention := "N/A"
	if actor, err := c.GetUser(actorID); err == nil && actor != nil {
		actorName = strings.TrimSpace(actor.FirstName + " " + actor.LastName)
		if actorName == "" {
			actorName = "N/A"
		}
		actorMention = mention(actor, actorID)
	}

	msg := F(config.LoggerID, "logger_bot_removed", locales.Arg{
		"group_name":          groupName,
		"group_id":            chatID,
		"group_username":      "N/A",
		"removed_by_name":     actorName,
		"removed_by_id":       actorID,
		"removed_by_username": actorMention,
		"date_time":           time.Now().Format("02 Jan 2006 • 15:04"),
	})

	if _, err := c.SendMessage(config.LoggerID, msg, nil); err != nil {
		utils.Log.Errorf("failed to send bot-removed log: %v", err)
	}
}

func handleVoiceChatAction(c *td.Client, m *td.Message) error {
	if isMaint, _ := database.GetMaintenance(); isMaint {
		return nil
	}

	chatID := m.ChatID()
	isActive := true
	var duration int32

	if ended, ok := m.Content.(*td.MessageVideoChatEnded); ok {
		isActive = false
		duration = ended.Duration
	}

	s, err := core.ChatStateFor(chatID)
	if err != nil {
		utils.Log.Errorf("Failed to get chat state for %d: %v", chatID, err)
		return nil
	}

	s.SetVoiceChat(isActive)

	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")
	if _, rerr := c.SendMessage(chatID, F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(duration))}), nil); rerr != nil {
		utils.Log.Error(rerr.Error())
	}
	utils.Log.Debugf("Voice chat %s in %d", msgKey, chatID)

	if !isActive {
		room, ok := core.Rooms.Get(chatID)
		go func() {
			time.Sleep(500 * time.Millisecond)
			if ok {
				scheduleOldPlayingMessage(room)
			}
			core.Rooms.Delete(chatID)
			utils.Log.Debugf("Room destroyed for ended voice chat in %d", chatID)
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
		action + "_by_username": mention(sender, actorID),
		"date_time":             time.Now().Format("02 Jan 2006 • 15:04"),
	}
}

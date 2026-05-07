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
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func handleActions(m *telegram.NewMessage) error {
	gologging.Info(m.Marshal())

	if !isValidChatType(m) {
		warnAndLeave(m.Client, m.ChannelID())
		return telegram.ErrEndGroup
	}

	if action, ok := m.Action.(*telegram.MessageActionGroupCall); ok {
		return handleVoiceChatAction(m, action)
	}

	return handleChatMemberAction(m)
}

func handleChatMemberAction(m *telegram.NewMessage) error {
	chatID := m.ChannelID()
	botID := m.Client.Me().ID

	switch action := m.Action.(type) {
	case *telegram.MessageActionChatAddUser:
		for _, uid := range action.Users {
			if uid == botID {
				if blockedChat, _ := database.IsBlacklistedChat(chatID); blockedChat && !isOwnerOrSudo(m.SenderID()) {
					m.Reply(F(chatID, "blacklist_chat_blocked"))
					leaveChat(m.Client, chatID)
					return telegram.ErrEndGroup
				}
				ownerID, err := utils.GetChatOwner(m.Client, chatID)
				if err == nil {
					if blockedOwner, _ := database.IsBlacklistedUser(ownerID); blockedOwner && !isOwnerOrSudo(m.SenderID()) {
						m.Reply(F(chatID, "blacklist_owner_blocked_leave"))
						leaveChat(m.Client, chatID)
						return telegram.ErrEndGroup
					}
				}

				gologging.Debug("Bot added to " + utils.IntToStr(chatID))
				m.Reply(F(chatID, "bot_added_normal"))
				database.AddServedChat(chatID)

				if config.LoggerID != 0 {
					m.Client.SendMessage(config.LoggerID, F(config.LoggerID, "logger_bot_added", buildLogArgs(m, chatID, "added")))
				}
				return nil
			}
		}

	case *telegram.MessageActionChatDeleteUser:
		if action.UserID == botID {
			gologging.Debug("Bot removed from " + utils.IntToStr(chatID))

			cleanScheduler.cancel(chatID)
			core.DeleteRoom(chatID)
			core.DeleteChatState(chatID)
			database.RemoveServedChat(chatID)

			if config.LoggerID != 0 {
				m.Client.SendMessage(config.LoggerID, F(config.LoggerID, "logger_bot_removed", buildLogArgs(m, chatID, "removed")))
			}
			return nil
		}
	}

	return nil
}

func handleVoiceChatAction(m *telegram.NewMessage, action *telegram.MessageActionGroupCall) error {
	if isMaint, _ := database.IsMaintenanceEnabled(); isMaint {
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	isActive := action.Duration == 0

	go clearRTMPState(chatID)
	s, err := core.GetChatState(chatID)
	if err != nil {
		gologging.ErrorF("Failed to get chat state for %d: %v", chatID, err)
		return telegram.ErrEndGroup
	}

	s.SetVoiceChatActive(isActive)

	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")
	m.Respond(F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(action.Duration))}))
	gologging.DebugF("Voice chat %s in %d", msgKey, chatID)

	if !isActive {
		room, ok := core.GetRoom(chatID, nil, false)
		go func() {
			time.Sleep(500 * time.Millisecond)
			if ok {
				scheduleOldPlayingMessage(room)
			}
			core.DeleteRoom(chatID)
			gologging.DebugF("Room destroyed for ended voice chat in %d", chatID)
		}()
	}

	return telegram.ErrEndGroup
}

func isValidChatType(m *telegram.NewMessage) bool {
	return m.ChatType() != telegram.EntityChat ||
		(m.Channel != nil && m.Channel.Megagroup)
}

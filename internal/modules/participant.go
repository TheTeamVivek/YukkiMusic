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
	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func memberID(m *td.ChatMember) int64 {
	if m == nil || m.MemberId == nil {
		return 0
	}
	if user, ok := m.MemberId.(*td.MessageSenderUser); ok {
		return user.UserId
	}
	return 0
}

func memberStatus(m *td.ChatMember) string {
	if m == nil || m.Status == nil {
		return "left"
	}

	switch status := m.Status.(type) {
	case *td.ChatMemberStatusCreator:
		return "creator"
	case *td.ChatMemberStatusAdministrator:
		return "administrator"
	case *td.ChatMemberStatusMember:
		return "member"
	case *td.ChatMemberStatusLeft:
		return "left"
	case *td.ChatMemberStatusBanned:
		return "kicked"
	case *td.ChatMemberStatusRestricted:
		if status.IsMember {
			return "member"
		}
		return "kicked"
	default:
		return "unknown"
	}
}

func handleParticipantUpdate(c *td.Client, u *td.UpdateChatMember) error {
	if !canBypassMaintenence(u.ActorUserId) {
		return nil
	}

	chatID := u.ChatId
	if chatID == 0 {
		return nil
	}

	userID := memberID(u.NewChatMember)
	if userID == 0 {
		userID = memberID(u.OldChatMember)
	}
	if userID == 0 {
		return nil
	}

	state, err := core.GetChatState(chatID)
	if err != nil {
		logger.Error("Failed to get chat state: " + err.Error())
		state = nil
	}

	oldStatus := memberStatus(u.OldChatMember)
	newStatus := memberStatus(u.NewChatMember)

	switch {
	case (newStatus == "administrator" || newStatus == "creator") &&
		(oldStatus != "administrator" && oldStatus != "creator"):

		utils.AddChatAdmin(c, chatID, userID)

	case oldStatus == "administrator" &&
		newStatus != "administrator" &&
		newStatus != "creator":

		if c.Me != nil && userID == c.Me.ID && config.LeaveOnDemoted {
			cleanScheduler.cancel(chatID)
			core.DeleteRoom(chatID)
			core.DeleteChatState(chatID)

			c.SendTextMessage(chatID, F(chatID, "bot_demotion_goodbye"), nil)
			c.LeaveChat(chatID)

			if state != nil && state.Assistant != nil {
				state.Assistant.Client.LeaveChannel(chatID)
			}

			return nil
		}

		utils.RemoveChatAdmin(c, chatID, userID)

	case oldStatus == "left" &&
		(newStatus == "member" || newStatus == "administrator" || newStatus == "creator"):

		handleSudoJoin(c, chatID, userID)
	}

	if state != nil && state.Assistant != nil && userID == state.Assistant.Self.ID {
		switch newStatus {
		case "member", "administrator", "creator":
			state.SetAssistantPresent(true)
			state.SetAssistantBanned(false)
			return nil

		case "left":
			state.SetAssistantPresent(false)
			state.SetAssistantBanned(false)
			return nil

		case "kicked":
			handleAssistantRestriction(c, u, state, chatID)
			return nil
		}

		if !state.AssistantFetched() {
			state.Snapshot(true)
		}
	}

	return nil
}

func handleSudoJoin(c *td.Client, chatID, userID int64) {
	var msgKey string

	if userID == config.OwnerID {
		msgKey = "sudo_join_owner"
	} else {
		isSudo, err := database.IsSudo(userID)
		if err != nil {
			logger.Errorf("IsSudo failed for user %d: %v", userID, err)
			return
		}
		if !isSudo {
			return
		}
		msgKey = "sudo_join_sudo"
	}

	user, _ := c.GetUser(userID)
	userMention := mentionOf(user, userID)

	botMention := "bot"
	if c.Me != nil {
		botMention = mentionOf(c.Me, c.Me.ID)
	}

	text := F(chatID, msgKey, locales.Arg{
		"user": userMention,
		"bot":  botMention,
	})

	c.SendTextMessage(chatID, text, nil)
}

func handleAssistantRestriction(
	c *td.Client,
	u *td.UpdateChatMember,
	s *core.ChatState,
	chatID int64,
) {
	if !isTrueBan(u) {
		s.SetAssistantPresent(true)
		s.SetAssistantBanned(false)

		logger.Debug("Assistant muted in " + utils.IntToStr(chatID))

		return
	}

	logger.Debug("Assistant banned in " + utils.IntToStr(chatID))

	s.SetAssistantPresent(false)
	if room, ok := core.GetRoom(chatID, nil, false); ok {
		scheduleOldPlayingMessage(room)
	}
	core.DeleteRoom(chatID)

	err := c.SetChatMemberStatus(
		chatID,
		&td.MessageSenderUser{UserId: s.Assistant.Self.ID},
		&td.ChatMemberStatusMember{},
	)
	if err == nil {
		s.SetAssistantBanned(false)
	} else {
		s.SetAssistantBanned(true)

		msg := F(chatID, "assistant_restricted_warning", locales.Arg{
			"assistant": mentionOfTg(s.Assistant.Self),
			"id":        s.Assistant.Self.ID,
		})

		c.SendTextMessage(chatID, msg, nil)
	}
}

func isTrueBan(u *td.UpdateChatMember) bool {
	if u.NewChatMember == nil {
		return false
	}

	_, banned := u.NewChatMember.Status.(*td.ChatMemberStatusBanned)
	return banned
}

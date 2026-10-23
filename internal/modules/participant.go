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
	"cmp"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func memberID(m *td.ChatMember) int64 {
	if m == nil {
		return 0
	}
	if user, ok := m.MemberId.(*td.MessageSenderUser); ok {
		return user.UserId
	}
	return 0
}

// memberStatus names the state of a chat member: creator, administrator,
// member, left or kicked (banned, or restricted and no longer a member).
func memberStatus(m *td.ChatMember) string {
	if m == nil {
		return "left"
	}

	switch status := m.Status.(type) {
	case *td.ChatMemberStatusCreator:
		return "creator"
	case *td.ChatMemberStatusAdministrator:
		return "administrator"
	case *td.ChatMemberStatusMember:
		return "member"
	case nil, *td.ChatMemberStatusLeft:
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

func isAdminStatus(status string) bool {
	return status == "administrator" || status == "creator"
}

func handleParticipantUpdate(c *td.Client, u *td.UpdateChatMember) error {
	if !canBypassMaintenance(u.ActorUserId) {
		return nil
	}

	chatID := u.ChatId
	userID := cmp.Or(memberID(u.NewChatMember), memberID(u.OldChatMember))
	if chatID == 0 || userID == 0 {
		return nil
	}

	oldStatus, newStatus := memberStatus(u.OldChatMember), memberStatus(u.NewChatMember)
	isBot := userID == c.Me().Id

	if isBot && (newStatus == "left" || newStatus == "kicked") {
		handleBotRemoved(c, u, chatID)
		return nil
	}

	switch {
	case isAdminStatus(newStatus) && !isAdminStatus(oldStatus):
		utils.AddChatAdmin(c, chatID, userID)

	case isAdminStatus(oldStatus) && !isAdminStatus(newStatus):
		if isBot && config.LeaveOnDemoted {
			dropChat(chatID)
			if _, err := c.SendMessage(chatID, F(chatID, "bot_demotion_goodbye"), nil); err != nil {
				utils.Log.Error(err.Error())
			}
			leaveChat(c, chatID)
			return nil
		}
		utils.RemoveChatAdmin(c, chatID, userID)
	}

	cs, err := core.ChatStateFor(chatID)
	if err != nil {
		utils.Log.Error("Failed to get chat state: " + err.Error())
		return nil
	}
	if userID != cs.Assistant().Self.ID {
		return nil
	}

	if presence := core.PresenceOf(u.NewChatMember); presence == core.PresenceBanned {
		handleAssistantBanned(c, cs, chatID)
	} else {
		cs.SetPresence(presence)
	}
	return nil
}

// handleAssistantBanned reacts to the assistant being banned: it gets
// unbanned if the bot can, and otherwise the chat is warned.
func handleAssistantBanned(c *td.Client, cs *core.ChatState, chatID int64) {
	utils.Log.Debug("Assistant banned in " + utils.IntToStr(chatID))
	cs.SetPresence(core.PresenceBanned)
	core.Rooms.Delete(chatID)

	if cs.Unban() == nil {
		return
	}

	self := cs.Assistant().Self
	msg := F(chatID, "assistant_restricted_warning", locales.Arg{
		"assistant": mentionHTML(self.FirstName+" "+self.LastName, self.ID),
		"id":        self.ID,
	})
	if _, err := c.SendMessage(chatID, msg, nil); err != nil {
		utils.Log.Error(err.Error())
	}
}

/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package modules

import (
	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/utils"
)

func shouldIgnoreParticipant(p *telegram.ParticipantUpdate) bool {
	if is, _ := database.IsMaintenance(); is && p.ActorID() != config.OwnerID {
		if ok, _ := database.IsSudo(p.ActorID()); !ok {
			return true
		}
	}
	return false
}

func handleParticipantUpdate(p *telegram.ParticipantUpdate) error {
	// Skip if in maintenance mode (non-owner + non-sudo)
	if shouldIgnoreParticipant(p) {
		return nil
	}

	if p.ChannelID() == 0 {
		return nil
	}

	chatID, err := utils.GetPeerID(p.Client, p.ChannelID())
	if err != nil {
		gologging.Error("Failed to resolve peer for chatID=" + utils.IntToStr(p.ChannelID()) + ", Error=" + err.Error())
		return err
	}

	if p.UserID() == core.UbUser.ID {
		handleAssistantState(p, chatID)
	}

	// Handle demotion
	if p.IsDemoted() {
		handleDemotion(p, chatID)
	}

	// Handle promotion
	if p.IsPromoted() {
		utils.AddChatAdmin(p.Client, chatID, p.UserID())
	}

	return nil
}

func handleActions(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	// Only allow in super groups
	if m.ChatType() == telegram.EntityChat && (m.Channel == nil || !m.Channel.Megagroup) {
		warnAndLeave(m.Client, chatID)
		return telegram.EndGroup
	}

	isMaintenance, _ := database.IsMaintenance()

	switch action := m.Action.(type) {
	case *telegram.MessageActionGroupCall:
		return handleGroupCallAction(m, chatID, action, isMaintenance)

	case *telegram.MessageActionChatAddUser:
		return handleAddUserAction(m, chatID, action, isMaintenance)

	case *telegram.MessageActionChatDeleteUser:
		return handleDeleteUserAction(m, chatID, action)

	default:
		return telegram.EndGroup
	}
}

func handleAssistantState(p *telegram.ParticipantUpdate, chatID int64) {
	s, _ := core.GetChatState(chatID, true)

	// Joined / Left
	if p.IsLeft() {
		s.SetAssistantPresence(bool_(false))
		s.SetAssistantBanned(bool_(false))
	}
	if p.IsJoined() {
		s.SetAssistantPresence(bool_(true))
		s.SetAssistantBanned(bool_(false))
	}

	// Restricted / Banned / Kicked
	if p.IsBanned() || p.IsKicked() || (p.New != nil && isAssistantRestricted(p.New)) {
		handleAssistantRestriction(p, chatID, s)
	}

	// Fallback state (cache update)
	if s.GetAssistantPresence() == nil || s.GetAssistantBanned() == nil {
		handleAssistantFallback(p, chatID, s)
	}
}

func handleAssistantRestriction(p *telegram.ParticipantUpdate, chatID int64, s *core.ChatState) {
	gologging.Debug("Assistant restricted in chatID " + utils.IntToStr(chatID))

	s.SetAssistantPresence(bool_(false))
	core.DeleteRoom(chatID)

	ok, err := p.Unban()
	if err != nil || !ok {
		s.SetAssistantBanned(bool_(true))

		if !shouldIgnoreParticipant(p) {
			_, sendErr := p.Client.SendMessage(chatID,
				"⚠️ <b>Assistant Restricted</b>\n\n"+
					"The assistant "+utils.MentionHTML(core.UbUser)+
					" (ID: <code>"+utils.IntToStr(core.UbUser.ID)+"</code>) has been <b>banned</b> or restricted in this chat.\n"+
					"I cannot play music, manage the queue, or clean up tracks while this lasts.\n\n"+
					"<i>Unban the assistant to restore all music features 🎵</i>",
			)
			if sendErr != nil {
				gologging.Error("Failed to send assistant restricted warning in ChatID: " +
					utils.IntToStr(chatID) + " Error: " + sendErr.Error())
			}
		}
	}
}

func handleAssistantFallback(p *telegram.ParticipantUpdate, chatID int64, s *core.ChatState) {
	member, err := p.Client.GetChatMember(chatID, core.UbUser.ID)
	if err != nil {
		if telegram.MatchError(err, "USER_NOT_PARTICIPANT") {
			s.SetAssistantPresence(bool_(false))
			s.SetAssistantBanned(bool_(false))
		} else {
			gologging.Error("Error getting assistant membership; ChatId: " +
				utils.IntToStr(chatID) + ", Error: " + err.Error())
		}
		return
	}

	switch member.Status {
	case telegram.Kicked, telegram.Restricted:
		s.SetAssistantPresence(bool_(false))
		s.SetAssistantBanned(bool_(true))
	case telegram.Left:
		s.SetAssistantPresence(bool_(false))
		s.SetAssistantBanned(bool_(false))
	case telegram.Admin, telegram.Member:
		s.SetAssistantPresence(bool_(true))
		s.SetAssistantBanned(bool_(false))
	}
}

func handleDemotion(p *telegram.ParticipantUpdate, chatID int64) {
	if p.UserID() == core.BUser.ID {
		core.DeleteRoom(chatID)
		core.DeleteChatState(chatID)
		p.Client.SendMessage(chatID,
			"<b>⚠️ Permission Lost</b>\n\n"+
				"I’ve been <b>demoted from admin</b> and can’t perform my tasks anymore.\n"+
				"Without proper rights, I’m unable to function here, so I’ll be <i>leaving the chat now</i>.\n\n"+
				"<b>Goodbye 👋</b>",
		)
		p.Client.LeaveChannel(chatID)
		core.UBot.LeaveChannel(chatID)
		return
	}
	utils.RemoveChatAdmin(p.Client, chatID, p.UserID())
}

func handleGroupCallAction(m *telegram.NewMessage, chatID int64, action *telegram.MessageActionGroupCall, isMaintenance bool) error {
	if isMaintenance {
		return telegram.EndGroup
	}

	core.DeleteRoom(chatID)
	s, _ := core.GetChatState(chatID, true)

	if action.Duration == 0 {
		s.SetVoiceChatStatus(bool_(true))
		m.Respond("📢 Voice chat started!\nUse /play <song> to play music")
		gologging.Debug("Voice chat started in " + utils.IntToStr(chatID))
	} else {
		s.SetVoiceChatStatus(bool_(false))
		m.Respond("📴 Voice chat ended!\nAll queues cleared")
		gologging.Debug("Voice chat ended in " + utils.IntToStr(chatID))
	}

	return telegram.EndGroup
}

func handleDeleteUserAction(m *telegram.NewMessage, chatID int64, action *telegram.MessageActionChatDeleteUser) error {
	gologging.Debug(
		"User removed from chatID " + utils.IntToStr(chatID) +
			": " + utils.IntToStr(action.UserID),
	)

	if action.UserID == core.BUser.ID {
		core.UBot.LeaveChannel(chatID)
		core.DeleteRoom(chatID)
		core.DeleteChatState(chatID)
		database.DeleteServed(chatID)
		gologging.Debug("Bot left chat " + utils.IntToStr(chatID))
		return telegram.EndGroup
	}
	return telegram.EndGroup
}

func handleAddUserAction(m *telegram.NewMessage, chatID int64, action *telegram.MessageActionChatAddUser, isMaintenance bool) error {
	for _, uid := range action.Users {
		gologging.Debug("User added to chatID " + utils.IntToStr(chatID) + ": " + utils.IntToStr(uid))

		if uid != core.BUser.ID {
			continue
		}

		if isMaintenance && m.SenderID() != config.OwnerID {
			if ok, _ := database.IsSudo(m.SenderID()); !ok {
				msg := "⚠️ I'm currently under maintenance and will be back soon. Thanks for your patience! ⏰"
				if reason, err := database.GetMaintReason(); err == nil && reason != "" {
					msg += "\n\n<i>📝 Reason: " + reason + "</i>"
				}
				m.Reply(msg)
				m.Client.LeaveChannel(chatID)
				gologging.Debug("Bot left chatID " + utils.IntToStr(chatID) + " due to maintenance")
				return telegram.EndGroup
			}
		}

		m.Respond("🎵 Thanks for adding me! Use /play <song> to start playing music.")
		gologging.Debug("Bot added to chat: " + utils.IntToStr(chatID))
		database.AddServed(chatID)
		return telegram.EndGroup
	}
	return telegram.EndGroup
}

func isAssistantRestricted(newParticipant telegram.ChannelParticipant) bool {
	_, left := newParticipant.(*telegram.ChannelParticipantLeft)
	_, banned := newParticipant.(*telegram.ChannelParticipantBanned)
	return left || banned
}

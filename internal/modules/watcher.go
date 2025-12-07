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
	"main/internal/locales"
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

	assistant, err := core.Assistants.ForChat(chatID)
	if err == nil && p.UserID() == assistant.User.ID {
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

	handleSudoJoin(p, chatID)
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
	s := core.GetChatState(chatID)

	// Joined / Left
	if p.IsLeft() {
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)
	}
	if p.IsJoined() {
		s.SetAssistantPresent(true)
		s.SetAssistantBanned(false)
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

	s.SetAssistantPresent(false)
	core.DeleteRoom(chatID)

	ok, err := p.Unban()
	if err != nil || !ok {
		s.SetAssistantBanned(true)
		assistant, aErr := core.Assistants.ForChat(chatID)
		if aErr != nil {
			gologging.Error("Failed to get assitand: " + aErr.Error())
			return
		}

		if !shouldIgnoreParticipant(p) {
			_, sendErr := assistant.Client.SendMessage(
				chatID,
				F(chatID, "assistant_restricted_warning", locales.Arg{
					"assistant": utils.MentionHTML(assistant.User),
					"id":        assistant.User.ID,
				}),
			)

			if sendErr != nil {
				gologging.Error("Failed to send assistant restricted warning in ChatID: " +
					utils.IntToStr(chatID) + " Error: " + sendErr.Error())
			}
		}
	}
}

func handleAssistantFallback(p *telegram.ParticipantUpdate, chatID int64, s *core.ChatState) {
	assistant, aErr := core.Assistants.ForChat(chatID)
	if aErr != nil {
		gologging.Error("Failed to get assitand: " + aErr.Error())
		return
	}
	member, err := p.Client.GetChatMember(chatID, assistant.User.ID)
	if err != nil {
		if telegram.MatchError(err, "USER_NOT_PARTICIPANT") {
			s.SetAssistantPresent(false)
			s.SetAssistantBanned(false)
		} else {
			gologging.Error("Error getting assistant membership; ChatId: " +
				utils.IntToStr(chatID) + ", Error: " + err.Error())
		}
		return
	}

	switch member.Status {
	case telegram.Kicked, telegram.Restricted:
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(true)
	case telegram.Left:
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)
	case telegram.Admin, telegram.Member:
		s.SetAssistantPresent(true)
		s.SetAssistantBanned(false)
	}
}

func handleDemotion(p *telegram.ParticipantUpdate, chatID int64) {
	if p.UserID() == core.BUser.ID {

		core.DeleteRoom(chatID)
		core.DeleteChatState(chatID)

		p.Client.SendMessage(chatID, F(chatID, "bot_demotion_goodbye"))

		p.Client.LeaveChannel(chatID)
		assistant, aErr := core.Assistants.ForChat(chatID)
		if aErr != nil {
			gologging.Error("Failed to get assitand: " + aErr.Error())
			return
		}

		assistant.Client.LeaveChannel(chatID)
		return
	}

	utils.RemoveChatAdmin(p.Client, chatID, p.UserID())
}

func handleSudoJoin(p *telegram.ParticipantUpdate, chatID int64) {
	var text string

	if p.UserID() == config.OwnerID {
		text = F(chatID, "sudo_join_owner", locales.Arg{
			"user": utils.MentionHTML(p.User),
			"bot":  utils.MentionHTML(core.BUser),
		})
	} else if database.IsSudoWithoutError(p.UserID()) {
		text = F(chatID, "sudo_join_sudo", locales.Arg{
			"user": utils.MentionHTML(p.User),
			"bot":  utils.MentionHTML(core.BUser),
		})
	}

	if text != "" {
		p.Client.SendMessage(chatID, text)
	}
}

func handleGroupCallAction(m *telegram.NewMessage, chatID int64, action *telegram.MessageActionGroupCall, isMaintenance bool) error {
	if isMaintenance {
		return telegram.EndGroup
	}

	core.DeleteRoom(chatID)
	s := core.GetChatState(chatID)

	if action.Duration == 0 {
		// Voice chat started
		s.SetVoiceChatActive(true)
		m.Respond(F(chatID, "voicechat_started"))
		gologging.Debug("Voice chat started in " + utils.IntToStr(chatID))
	} else {
		// Voice chat ended
		s.SetVoiceChatActive(false)
		m.Respond(F(chatID, "voicechat_ended"))
		gologging.Debug("Voice chat ended in " + utils.IntToStr(chatID))
	}

	return telegram.EndGroup
}

func handleDeleteUserAction(m *telegram.NewMessage, chatID int64, action *telegram.MessageActionChatDeleteUser) error {
	gologging.Debug(
		"User removed from chatID " + utils.IntToStr(chatID) +
			": " + utils.IntToStr(action.UserID),
	)
	assistant, aErr := core.Assistants.ForChat(chatID)
	if aErr != nil {
		gologging.Error("Failed to get assitand: " + aErr.Error())
	}
	if aErr == nil && action.UserID == assistant.User.ID {
		assistant.Client.LeaveChannel(chatID)
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

		// Not the bot, skip
		if uid != core.BUser.ID {
			continue
		}

		// Bot added during maintenance
		// Only owner + sudo can add the bot during maintenance
		if isMaintenance && m.SenderID() != config.OwnerID {
			if ok, _ := database.IsSudo(m.SenderID()); !ok {

				msg := F(chatID, "bot_added_maintenance")
				if reason, err := database.GetMaintReason(); err == nil && reason != "" {
					msg += "\n\n" + F(chatID, "maint_reason_generic", locales.Arg{
						"reason": reason,
					})
				}

				m.Reply(msg)
				m.Client.LeaveChannel(chatID)

				gologging.Debug("Bot left chatID " + utils.IntToStr(chatID) + " due to maintenance")
				return telegram.EndGroup
			}
		}

		m.Respond(F(chatID, "bot_added_normal"))

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

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
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
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

	chatID := p.ChannelID()

	if chatID == 0 {
		return nil
	}

	s, err := core.GetChatState(chatID)
	if err != nil {
		gologging.Error("Failed to get chat state: " + err.Error())
	}

	if err == nil && p.UserID() == s.Assistant.User.ID {
		handleAssistantState(p, s, chatID)
	}

	if p.UserID() == core.BUser.ID {
		handleBotState(p, chatID)
	}

	// Handle demotion
	if p.IsDemoted() {
		handleDemotion(p, s, chatID)
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
		return telegram.ErrEndGroup
	}

	isMaintenance, _ := database.IsMaintenance()

	switch action := m.Action.(type) {
	case *telegram.MessageActionGroupCall:
		return handleGroupCallAction(m, chatID, action, isMaintenance)

	case *telegram.MessageActionChatAddUser:
		return handleAddUserAction(m, chatID, action, isMaintenance)

	default:
		return telegram.ErrEndGroup
	}
}

func handleBotState(p *telegram.ParticipantUpdate, chatID int64) {
	action := "removed"
	if p.IsLeft() {
		action = "left"
	}
	if !p.IsLeft() && !p.IsBanned() && !p.IsKicked() && !(p.New != nil && isRestricted(p.New)) {
		return
	}

	gologging.Debug(
		"Bot " + action + " from chatID " + utils.IntToStr(chatID) +
			": " + utils.IntToStr(p.ActorID()),
	)

	ass, aErr := core.Assistants.ForChat(chatID)
	if aErr != nil {
		gologging.ErrorF("Failed to get Assistant for %d: %v", chatID, aErr)
	}
	core.DeleteRoom(chatID)
	core.DeleteChatState(chatID)
	database.DeleteServed(chatID)
	if aErr == nil {
		ass.Client.LeaveChannel(chatID)
	}
	if config.LoggerID != 0 && isLogger() {
		group_username := "N/A"
		removed_by_username := utils.MentionHTML(p.Actor)

		if u := p.Channel.Username; u != "" {
			group_username = "@" + u
		}

		if u := p.Actor.Username; u != "" {
			removed_by_username = "@" + u
		}

		msg := F(config.LoggerID, "logger_bot_removed", locales.Arg{
			"group_name":     p.Channel.Title,
			"group_id":       chatID,
			"group_username": group_username,

			"removed_by_name":     strings.TrimSpace(p.Actor.FirstName + " " + p.Actor.LastName),
			"removed_by_id":       p.ActorID(),
			"removed_by_username": removed_by_username,

			"date_time": time.Now().Format("02 Jan 2006 • 15:04"),
		})

		_, err := p.Client.SendMessage(config.LoggerID, msg)
		if err != nil {
			gologging.Error("Failed to send logger_bot_removed msg, Error: " + err.Error())
		}
	}
	gologging.Debug("Bot left chat " + utils.IntToStr(chatID))

	return
}

func handleAssistantState(p *telegram.ParticipantUpdate, s *core.ChatState, chatID int64) {
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
	if p.IsBanned() || p.IsKicked() || (p.New != nil && isRestricted(p.New)) {
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

		if !shouldIgnoreParticipant(p) {
			_, sendErr := s.Assistant.Client.SendMessage(
				chatID,
				F(chatID, "assistant_restricted_warning", locales.Arg{
					"assistant": utils.MentionHTML(s.Assistant.User),
					"id":        s.Assistant.User.ID,
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
	member, err := p.Client.GetChatMember(chatID, s.Assistant.User.ID)
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

func handleDemotion(p *telegram.ParticipantUpdate, s *core.ChatState, chatID int64) {
	if p.UserID() == core.BUser.ID && config.LeaveOnDemoted {

		core.DeleteRoom(chatID)
		core.DeleteChatState(chatID)

		p.Client.SendMessage(chatID, F(chatID, "bot_demotion_goodbye"))

		p.Client.LeaveChannel(chatID)
		if s != nil && s.Assistant != nil {
			s.Assistant.Client.LeaveChannel(chatID)
		}
		return
	}

	utils.RemoveChatAdmin(p.Client, chatID, p.UserID())
}

func handleSudoJoin(p *telegram.ParticipantUpdate, chatID int64) {
	var text string

	if !p.IsJoined() { return }

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
		return telegram.ErrEndGroup
	}

	core.DeleteRoom(chatID)
	s, err := core.GetChatState(chatID)
	if err != nil {
		gologging.Error("Failed to get chat state: " + err.Error())
		return telegram.ErrEndGroup
	}

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

	return telegram.ErrEndGroup
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
				return telegram.ErrEndGroup
			}
		}

		m.Respond(F(chatID, "bot_added_normal"))

		if config.LoggerID != 0 && isLogger() {
			group_username := "N/A"
			removed_by_username := utils.MentionHTML(m.Sender)

			if u := m.Channel.Username; u != "" {
				group_username = "@" + u
			}

			if u := m.Sender.Username; u != "" {
				removed_by_username = "@" + u
			}

			msg := F(config.LoggerID, "logger_bot_added", locales.Arg{
				"group_name":     m.Channel.Title,
				"group_id":       chatID,
				"group_username": group_username,

				"added_by_name":     strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName),
				"added_by_id":       m.SenderID(),
				"added_by_username": removed_by_username,

				"date_time": time.Now().Format("02 Jan 2006 • 15:04"),
				// "members_count": memCount,
			})

			_, err := m.Client.SendMessage(config.LoggerID, msg)
			if err != nil {
				gologging.Error("Failed to send logger_bot_added msg, Error: " + err.Error())
			}
		}

		gologging.Debug("Bot added to chat: " + utils.IntToStr(chatID))
		database.AddServed(chatID)
		return telegram.ErrEndGroup
	}
	return telegram.ErrEndGroup
}

func isRestricted(newParticipant telegram.ChannelParticipant) bool {
	_, left := newParticipant.(*telegram.ChannelParticipantLeft)
	_, banned := newParticipant.(*telegram.ChannelParticipantBanned)
	return left || banned
}

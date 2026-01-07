/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
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

func handleParticipantUpdate(p *telegram.ParticipantUpdate) error {
	if isMaintenanceBlocked(p.ActorID()) {
		return nil
	}

	chatID := p.ChannelID()
	if chatID == 0 {
		return nil
	}

	state, err := core.GetChatState(chatID)
	if err != nil {
		gologging.Error("Failed to get chat state: " + err.Error())
		state = nil
	}

	userID := p.UserID()

	switch {
	case userID == core.BUser.ID:
		handleBotEvent(p, state, chatID)
	case state != nil && userID == state.Assistant.User.ID:
		handleAssistantEvent(p, state, chatID)
	default:
		handleOtherUserEvent(p, state, chatID)
	}

	return nil
}

// =============================================================================
// BOT EVENT HANDLERS
// =============================================================================

func handleBotEvent(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
	if p.IsJoined() {
		handleBotJoin(p, chatID)
		return
	}

	if isUserPresent(p) {
		return
	}

	handleBotLeave(p, s, chatID)
}

func handleBotJoin(p *telegram.ParticipantUpdate, chatID int64) {
	if isMaintenanceBlocked(p.ActorID()) {
		sendMaintenanceNotice(p, chatID)
		p.Client.LeaveChannel(chatID)
		return
	}

	p.Client.SendMessage(chatID, F(chatID, "bot_added_normal"))
	database.AddServed(chatID)
	logBotJoin(p, chatID)
}

func handleBotLeave(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
	action := getLeaveAction(p)
	gologging.Debug("Bot " + action + " from chat " + utils.IntToStr(chatID))

	if s != nil && s.Assistant != nil {
		s.Assistant.Client.LeaveChannel(chatID)
	}

	core.DeleteRoom(chatID)
	core.DeleteChatState(chatID)
	database.DeleteServed(chatID)

	logBotLeave(p, chatID)
}

// =============================================================================
// ASSISTANT EVENT HANDLERS
// =============================================================================

func handleAssistantEvent(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
	if p.IsJoined() {
		s.SetAssistantPresent(true)
		s.SetAssistantBanned(false)
		return
	}

	if p.IsLeft() {
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)
		return
	}

	if isUserRestricted(p) {
		handleAssistantRestriction(p, s, chatID)
		return
	}

	if s.IsStateUnknown() {
		s.RefreshAssistantState()
	}
}

func handleAssistantRestriction(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
	if !isTrueBan(p) {
		s.SetAssistantPresent(true)
		s.SetAssistantBanned(false)
		gologging.Debug("Assistant muted in " + utils.IntToStr(chatID))
		return
	}

	gologging.Debug("Assistant banned in " + utils.IntToStr(chatID))
	s.SetAssistantPresent(false)
	core.DeleteRoom(chatID)

	if ok, _ := p.Unban(); ok {
		s.SetAssistantBanned(false)
	} else {
		s.SetAssistantBanned(true)
		notifyAssistantRestricted(p, s, chatID)
	}
}

// =============================================================================
// OTHER USER EVENT HANDLERS
// =============================================================================

func handleOtherUserEvent(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
	if p.IsPromoted() {
		utils.AddChatAdmin(p.Client, chatID, p.UserID())
	}

	if p.IsDemoted() {
		handleDemotion(p, s, chatID)
	}

	if p.IsJoined() {
		handleSudoJoin(p, chatID)
	}
}

func handleDemotion(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
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
	var msgKey string

	if p.UserID() == config.OwnerID {
		msgKey = "sudo_join_owner"
	} else if database.IsSudoWithoutError(p.UserID()) {
		msgKey = "sudo_join_sudo"
	} else {
		return
	}

	text := F(chatID, msgKey, locales.Arg{
		"user": utils.MentionHTML(p.User),
		"bot":  utils.MentionHTML(core.BUser),
	})

	p.Client.SendMessage(chatID, text)
}

// =============================================================================
// DETECTION FUNCTIONS
// =============================================================================

func isTrueBan(p *telegram.ParticipantUpdate) bool {
	if p.New == nil {
		return false
	}

	banned, ok := p.New.(*telegram.ChannelParticipantBanned)
	if !ok {
		return false
	}

	// ViewMessages=true means actually banned (can't access chat)
	// ViewMessages=false means muted (can access but restricted)
	return banned.BannedRights.ViewMessages
}

func isUserPresent(p *telegram.ParticipantUpdate) bool {
	if p.IsLeft() || p.IsBanned() || p.IsKicked() {
		return false
	}
	return p.New == nil || !isUserRestricted(p)
}

func isUserRestricted(p *telegram.ParticipantUpdate) bool {
	if p.New == nil {
		return false
	}

	_, banned := p.New.(*telegram.ChannelParticipantBanned)
	_, left := p.New.(*telegram.ChannelParticipantLeft)
	return banned || left
}

// =============================================================================
// HELPERS
// =============================================================================

func getLeaveAction(p *telegram.ParticipantUpdate) string {
	if p.IsLeft() {
		return "left"
	}
	return "removed"
}

func notifyAssistantRestricted(
	p *telegram.ParticipantUpdate,
	s *core.ChatState,
	chatID int64,
) {
	if isMaintenanceBlocked(p.ActorID()) {
		return
	}

	msg := F(chatID, "assistant_restricted_warning", locales.Arg{
		"assistant": utils.MentionHTML(s.Assistant.User),
		"id":        s.Assistant.User.ID,
	})

	if _, err := p.Client.SendMessage(chatID, msg); err != nil {
		gologging.Error("Failed to send restriction warning: " + err.Error())
	}
}

func sendMaintenanceNotice(p *telegram.ParticipantUpdate, chatID int64) {
	msg := F(chatID, "bot_added_maintenance")
	if reason, err := database.GetMaintReason(); err == nil && reason != "" {
		msg += "\n\n" + F(chatID, "maint_reason_generic",
			locales.Arg{"reason": reason})
	}

	p.Client.SendMessage(chatID, msg)
}

// =============================================================================
// LOGGING
// =============================================================================

func logBotJoin(p *telegram.ParticipantUpdate, chatID int64) {
	if config.LoggerID == 0 || !isLogger() {
		return
	}

	msg := F(
		config.LoggerID,
		"logger_bot_added",
		buildLogArgs(p, chatID, "added"),
	)
	if _, err := p.Client.SendMessage(config.LoggerID, msg); err != nil {
		gologging.Error("Failed to send logger_bot_added: " + err.Error())
	}
}

func logBotLeave(p *telegram.ParticipantUpdate, chatID int64) {
	if config.LoggerID == 0 || !isLogger() {
		return
	}

	msg := F(
		config.LoggerID,
		"logger_bot_removed",
		buildLogArgs(p, chatID, "removed"),
	)
	if _, err := p.Client.SendMessage(config.LoggerID, msg); err != nil {
		gologging.Error("Failed to send logger_bot_removed: " + err.Error())
	}
}

func buildLogArgs(
	p *telegram.ParticipantUpdate,
	chatID int64,
	action string,
) locales.Arg {
	groupUsername := "N/A"
	if u := p.Channel.Username; u != "" {
		groupUsername = "@" + u
	}

	actorUsername := utils.MentionHTML(p.Actor)
	if u := p.Actor.Username; u != "" {
		actorUsername = "@" + u
	}

	actorName := strings.TrimSpace(p.Actor.FirstName + " " + p.Actor.LastName)

	return locales.Arg{
		"group_name":            p.Channel.Title,
		"group_id":              chatID,
		"group_username":        groupUsername,
		action + "_by_name":     actorName,
		action + "_by_id":       p.ActorID(),
		action + "_by_username": actorUsername,
		"date_time":             time.Now().Format("02 Jan 2006 • 15:04"),
	}
}

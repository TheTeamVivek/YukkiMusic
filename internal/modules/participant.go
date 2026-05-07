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

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func getParticipantStatus(p telegram.ChannelParticipant) string {
	if p == nil {
		return "left"
	}

	switch v := p.(type) {
	case *telegram.ChannelParticipantCreator:
		return "creator"
	case *telegram.ChannelParticipantAdmin:
		return "administrator"
	case *telegram.ChannelParticipantSelf, *telegram.ChannelParticipantObj:
		return "member"
	case *telegram.ChannelParticipantLeft:
		return "left"
	case *telegram.ChannelParticipantBanned:
		if v.Left {
			return "left"
		}
		return "kicked"
	default:
		return "unknown"
	}
}

func handleParticipantUpdate(p *telegram.ParticipantUpdate) error {
	if !canBypassMaintenence(p.ActorID()) {
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

	oldStatus := getParticipantStatus(p.Old)
	newStatus := getParticipantStatus(p.New)

	/*gologging.DebugF(
		"participant change %d: %s -> %s",
		userID,
		oldStatus,
		newStatus,
	)*/

	switch {

	case (newStatus == "administrator" || newStatus == "creator") &&
		(oldStatus != "administrator" && oldStatus != "creator"):

		utils.AddChatAdmin(p.Client, chatID, userID)

	case oldStatus == "administrator" &&
		newStatus != "administrator" &&
		newStatus != "creator":

		if userID == p.Client.Me().ID && config.LeaveOnDemoted {

			cleanScheduler.cancel(chatID)
			core.DeleteRoom(chatID)
			core.DeleteChatState(chatID)

			p.Client.SendMessage(chatID, F(chatID, "bot_demotion_goodbye"))
			p.Client.LeaveChannel(chatID)

			if state != nil && state.Assistant != nil {
				state.Assistant.Client.LeaveChannel(chatID)
			}

			return nil
		}

		utils.RemoveChatAdmin(p.Client, chatID, userID)

	case oldStatus == "left" &&
		(newStatus == "member" || newStatus == "administrator" || newStatus == "creator"):

		handleSudoJoin(p, chatID)
	}

	if state != nil && userID == state.Assistant.Self.ID {

		if p.IsJoined() {
			state.SetAssistantPresent(true)
			state.SetAssistantBanned(false)
			return nil
		}

		if p.IsLeft() {
			state.SetAssistantPresent(false)
			state.SetAssistantBanned(false)
			return nil
		}

		if isUserRestricted(p) {
			handleAssistantRestriction(p, state, chatID)
			return nil
		}

		if !state.AssistantFetched() {
			state.Snapshot(true)
		}
	}

	return nil
}

func handleSudoJoin(p *telegram.ParticipantUpdate, chatID int64) {
	userID := p.UserID()

	var msgKey string

	if userID == config.OwnerID {
		msgKey = "sudo_join_owner"
	} else {
		isSudo, err := database.IsSudo(userID)
		if err != nil {
			gologging.ErrorF("IsSudo failed for user %d: %v", userID, err)
			return
		}
		if !isSudo {
			return
		}
		msgKey = "sudo_join_sudo"
	}

	text := F(chatID, msgKey, locales.Arg{
		"user": utils.MentionHTML(p.User),
		"bot":  utils.MentionHTML(p.Client.Me()),
	})

	p.Client.SendMessage(chatID, text)
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
	if room, ok := core.GetRoom(chatID, nil, false); ok {
		scheduleOldPlayingMessage(room)
	}
	core.DeleteRoom(chatID)

	if ok, _ := p.Unban(); ok {
		s.SetAssistantBanned(false)
	} else {

		s.SetAssistantBanned(true)

		msg := F(chatID, "assistant_restricted_warning", locales.Arg{
			"assistant": utils.MentionHTML(s.Assistant.Self),
			"id":        s.Assistant.Self.ID,
		})

		p.Client.SendMessage(chatID, msg)
	}
}

func isTrueBan(p *telegram.ParticipantUpdate) bool {
	if p.New == nil {
		return false
	}

	banned, ok := p.New.(*telegram.ChannelParticipantBanned)
	if !ok {
		return false
	}

	return banned.BannedRights.ViewMessages
}

func isUserRestricted(p *telegram.ParticipantUpdate) bool {
	if p.New == nil {
		return false
	}

	_, banned := p.New.(*telegram.ChannelParticipantBanned)
	_, left := p.New.(*telegram.ChannelParticipantLeft)

	return banned || left
}

func buildLogArgs(
	m *telegram.NewMessage,
	chatID int64,
	action string,
) locales.Arg {
	groupUsername := "N/A"
	if u := m.Channel.Username; u != "" {
		groupUsername = "@" + u
	}

	actorUsername := utils.MentionHTML(m.Sender)
	if u := m.Sender.Username; u != "" {
		actorUsername = "@" + u
	}

	name := strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)

	return locales.Arg{
		"group_name":            m.Channel.Title,
		"group_id":              chatID,
		"group_username":        groupUsername,
		action + "_by_name":     name,
		action + "_by_id":       m.SenderID(),
		action + "_by_username": actorUsername,
		"date_time":             time.Now().Format("02 Jan 2006 • 15:04"),
	}
}

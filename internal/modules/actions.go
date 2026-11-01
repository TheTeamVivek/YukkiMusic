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
	"slices"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func handleActions(c *td.Client, m *td.Message) error {
	if m.IsPrivate() || isChannel(c, m) {
		return nil
	}

	switch content := m.Content.(type) {
	case *td.MessageChatAddMembers:
		return handleChatMemberAdd(c, m, content.MemberUserIds)
	case *td.MessageVideoChatStarted:
		return handleVoiceChatAction(c, m.ChatID(), true, 0)
	case *td.MessageVideoChatEnded:
		return handleVoiceChatAction(c, m.ChatID(), false, content.Duration)
	default:
		return nil
	}
}

func handleChatMemberAdd(c *td.Client, m *td.Message, added []int64) error {
	me := c.Me()
	if me == nil || me.Id == 0 || !slices.Contains(added, me.Id) {
		return nil
	}

	chatID := m.ChatID()
	if !isOwnerOrSudo(m.SenderID()) && isOwnerBlacklisted(c, chatID) {
		replyKey(c, m, "blacklist_owner_blocked_leave")
		leaveChat(c, chatID)
		return nil
	}

	utils.Log.Debugf("Bot added to chat %d", chatID)
	if _, err := m.ReplyText(c, F(chatID, "bot_added_normal"), nil); err != nil {
		return err
	}
	database.AddServedChat(chatID)
	logChatAction(c, chatID, m.SenderID(), "added")
	return nil
}

// isOwnerBlacklisted reports whether the owner of a chat is blacklisted.
func isOwnerBlacklisted(c *td.Client, chatID int64) bool {
	owner, err := utils.GetChatOwner(c, chatID)
	if err != nil {
		return false
	}
	blocked, _ := database.IsBlacklistedUser(owner)
	return blocked
}

func handleVoiceChatAction(c *td.Client, chatID int64, isActive bool, duration int32) error {
	if maintenance, _ := database.GetMaintenance(); maintenance {
		return nil
	}

	s, err := core.ChatStateFor(chatID)
	if err != nil {
		utils.Log.Errorf("Failed to get chat state for %d: %v", chatID, err)
		return nil
	}
	s.SetVoice(utils.IfElse(isActive, core.VoiceActive, core.VoiceInactive))

	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")
	text := F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(duration))})
	if _, err := c.SendMessage(chatID, text, nil); err != nil {
		utils.Log.Error(err.Error())
	}
	utils.Log.Debugf("Voice chat %s in %d", msgKey, chatID)

	if !isActive {
		time.AfterFunc(500*time.Millisecond, func() {
			core.Rooms.Delete(chatID)
			utils.Log.Debugf("Room destroyed for ended voice chat in %d", chatID)
		})
	}
	return nil
}

func memberID(m *td.ChatMember) int64 {
	if m == nil {
		return 0
	}
	if user, ok := m.MemberId.(*td.MessageSenderUser); ok {
		return user.UserId
	}
	return 0
}

type status string

const (
	statusCreator       status = "creator"
	statusAdministrator status = "administrator"
	statusMember        status = "member"
	statusLeft          status = "left"
	statusKicked        status = "kicked"
)

func (s status) isAdmin() bool {
	return s == statusCreator || s == statusAdministrator
}

func (s status) isGone() bool {
	return s == statusLeft || s == statusKicked
}

func memberStatus(m *td.ChatMember) status {
	if m == nil {
		return statusLeft
	}

	switch st := m.Status.(type) {
	case *td.ChatMemberStatusCreator:
		return statusCreator
	case *td.ChatMemberStatusAdministrator:
		return statusAdministrator
	case *td.ChatMemberStatusMember:
		return statusMember
	case *td.ChatMemberStatusBanned:
		return statusKicked
	case *td.ChatMemberStatusRestricted:
		if st.IsMember {
			return statusMember
		}
		return statusKicked
	}
	return statusLeft
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

	was, now := memberStatus(u.OldChatMember), memberStatus(u.NewChatMember)
	isBot := userID == c.Me().Id

	if isBot && now.isGone() {
		if chatID > 0 {
			utils.Log.Infof("Bot was blocked by user: %d", chatID)
			database.RemoveServedUser(chatID)
			return nil
		}
		handleBotRemoved(c, u, chatID)
		return nil
	}

	switch {
	case now.isAdmin() && !was.isAdmin():
		utils.AddChatAdmin(c, chatID, userID)

	case was.isAdmin() && !now.isAdmin():
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

func handleBotRemoved(c *td.Client, u *td.UpdateChatMember, chatID int64) {
	dropChat(chatID)
	database.RemoveServedChat(chatID)

	if u.ActorUserId == c.Me().Id {
		return
	}

	utils.Log.Debugf("Bot removed from chat %d", chatID)
	logChatAction(c, chatID, u.ActorUserId, "removed")
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

// dropChat forgets everything the bot holds for a chat it no longer serves.
func dropChat(chatID int64) {
	cleaner.Clear(chatID)
	core.Rooms.Delete(chatID)
	core.DropChatState(chatID)
}

func logChatAction(c *td.Client, chatID, actorID int64, action string) {
	if config.LoggerID == 0 || config.LoggerID == chatID {
		return
	}
	msg := F(config.LoggerID, "logger_bot_"+action, logArgs(c, chatID, actorID, action))
	if _, err := c.SendMessage(config.LoggerID, msg, nil); err != nil {
		utils.Log.Errorf("failed to send bot-%s log: %v", action, err)
	}
}

// logArgs fills the logger templates for something an actor did to a chat;
// action is "added" or "removed" and names the *_by_* keys.
func logArgs(c *td.Client, chatID, actorID int64, action string) locales.Arg {
	groupName := "N/A"
	if chat, err := c.GetChat(chatID); err == nil && chat != nil {
		groupName = chat.Title
	}

	actor, _ := c.GetUser(actorID)
	actorName := "N/A"
	if actor != nil {
		if full := strings.TrimSpace(actor.FirstName + " " + actor.LastName); full != "" {
			actorName = full
		}
	}

	return locales.Arg{
		"group_name":            groupName,
		"group_id":              chatID,
		"group_username":        "N/A",
		action + "_by_name":     actorName,
		action + "_by_id":       actorID,
		action + "_by_username": mention(actor, actorID),
		"date_time":             time.Now().Format("02 Jan 2006 • 15:04"),
	}
}

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
	if m.IsPrivate() || isChannel(c, m) {
		return nil
	}

	if _, added := m.Content.(*td.MessageChatAddMembers); added {
		return handleChatMemberAdd(c, m)
	}
	return handleVoiceChatAction(c, m)
}

func handleChatMemberAdd(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()
	added := m.Content.(*td.MessageChatAddMembers).MemberUserIds
	me := c.Me()
	if me == nil || me.Id == 0 || !slices.Contains(added, me.Id) {
		return nil
	}

	if !isOwnerOrSudo(m.SenderID()) {
		if owner, err := utils.GetChatOwner(c, chatID); err == nil {
			if blocked, _ := database.IsBlacklistedUser(owner); blocked {
				replyKey(c, m, "blacklist_owner_blocked_leave")
				leaveChat(c, chatID)
				return nil
			}
		}
	}

	utils.Log.Debug("Bot added to " + utils.IntToStr(chatID))
	if _, err := m.ReplyText(c, F(chatID, "bot_added_normal"), nil); err != nil {
		return err
	}
	database.AddServedChat(chatID)

	if config.LoggerID == 0 {
		return nil
	}
	_, err := c.SendMessage(
		config.LoggerID,
		F(config.LoggerID, "logger_bot_added", logArgs(c, chatID, m.SenderID(), "added")),
		nil,
	)
	return err
}

// dropChat forgets everything the bot holds for a chat it no longer serves.
func dropChat(chatID int64) {
	cleaner.Clear(chatID)
	core.Rooms.Delete(chatID)
	core.DropChatState(chatID)
}

func handleBotRemoved(c *td.Client, u *td.UpdateChatMember, chatID int64) {
	utils.Log.Debugf("Bot removed from chat %d", chatID)

	dropChat(chatID)
	database.RemoveServedChat(chatID)

	if config.LoggerID == 0 || config.LoggerID == chatID {
		return
	}
	msg := F(config.LoggerID, "logger_bot_removed", logArgs(c, chatID, u.ActorUserId, "removed"))
	if _, err := c.SendMessage(config.LoggerID, msg, nil); err != nil {
		utils.Log.Errorf("failed to send bot-removed log: %v", err)
	}
}

func handleVoiceChatAction(c *td.Client, m *td.Message) error {
	if maintenance, _ := database.GetMaintenance(); maintenance {
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
	s.SetVoice(utils.IfElse(isActive, core.VoiceActive, core.VoiceInactive))

	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")
	text := F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(duration))})
	if _, err := c.SendMessage(chatID, text, nil); err != nil {
		utils.Log.Error(err.Error())
	}
	utils.Log.Debugf("Voice chat %s in %d", msgKey, chatID)

	if !isActive {
		go func() {
			time.Sleep(500 * time.Millisecond)
			core.Rooms.Delete(chatID)
			utils.Log.Debugf("Room destroyed for ended voice chat in %d", chatID)
		}()
	}
	return nil
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

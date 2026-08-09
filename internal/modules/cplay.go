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
	"errors"
	"strings"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const linkedCPlayTarget = "linked"

func setCPlayHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	args := strings.Fields(m.Text())
	if len(args) <= 1 {
		m.ReplyText(c, F(m.ChatID(), "cplay_usage"), nil)
		return nil
	}

	chatID := m.ChatID()
	arg := strings.TrimSpace(args[1])

	enabled, boolErr := utils.ParseBool(arg)
	if boolErr == nil && !enabled {
		return disableCPlay(c, m, chatID)
	}

	var targetChannelID int64
	var err error

	if strings.EqualFold(arg, linkedCPlayTarget) {
		targetChannelID, err = getLinkedChannelID(c, chatID)
	} else {
		targetChannelID, err = resolveChannelPlay(c, chatID, arg)
	}
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	me := c.Me
	if me == nil {
		if fetched, err := c.GetMe(); err == nil && fetched != nil {
			me = fetched
			c.Me = me
		}
	}
	if me == nil {
		m.ReplyText(c, F(chatID, "cplay_channel_not_accessible"), nil)
		return nil
	}

	member, err := c.GetChatMember(targetChannelID, &td.MessageSenderUser{UserId: me.Id})
	if err != nil {
		logger.Errorf("Failed to fetch bot member state for cplay target %d: %v", targetChannelID, err)
		m.ReplyText(c, F(chatID, "cplay_channel_not_accessible"), nil)
		return nil
	}
	if member == nil {
		m.ReplyText(c, F(chatID, "cplay_channel_not_accessible"), nil)
		return nil
	}

	isAdmin := false
	canInvite := false
	switch st := member.Status.(type) {
	case *td.ChatMemberStatusCreator:
		isAdmin = true
		canInvite = true
	case *td.ChatMemberStatusAdministrator:
		isAdmin = true
		canInvite = st.Rights != nil && st.Rights.CanInviteUsers
	}
	if !isAdmin {
		m.ReplyText(c, F(chatID, "cplay_channel_not_accessible"), nil)
		return nil
	}
	if isAdmin && !canInvite {
		m.ReplyText(c, F(chatID, "cplay_bot_invite_permission_missing"), nil)
		return nil
	}

	return saveCPlayTarget(c, m, chatID, targetChannelID)
}

func disableCPlay(c *td.Client, m *td.Message, chatID int64) error {
	allowed, err := canSetCPlayTarget(c, m, chatID, chatID)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}
	if !allowed {
		m.ReplyText(c, F(chatID, "cplay_owner_required"), nil)
		return nil
	}

	if err := database.LinkChannel(chatID, 0); err != nil {
		logger.Errorf("Failed to disable cplay for chat %d: %v", chatID, err)
		m.ReplyText(c, F(chatID, "cplay_save_error"), nil)
		return nil
	}

	m.ReplyText(c, F(chatID, "cplay_disabled"), nil)
	return nil
}

func getLinkedChannelID(c *td.Client, chatID int64) (int64, error) {
	chat, err := c.GetChat(chatID)
	if err != nil {
		return 0, errors.New(F(chatID, "cplay_resolve_peer_fail"))
	}

	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	if !ok {
		return 0, errors.New(F(chatID, "supergroup_needed", locales.Arg{"chat_id": chatID, "support_group": config.SupportChat}))
	}

	full, err := c.GetSupergroupFullInfo(sg.SupergroupId)
	if err != nil || full == nil || full.LinkedChatId == 0 {
		return 0, errors.New(F(chatID, "cplay_channel_not_linked"))
	}
	return full.LinkedChatId, nil
}

func resolveChannelPlay(c *td.Client, chatID int64, target string) (int64, error) {
	chat, err := c.SearchPublicChat(target)
	if err != nil {
		logger.Errorf("Failed to resolve cplay target %v for chat %d: %v", target, chatID, err)
		return 0, errors.New(F(chatID, "cplay_channel_not_accessible"))
	}
	if chat == nil {
		return 0, errors.New(F(chatID, "cplay_channel_not_accessible"))
	}

	if _, ok := chat.Type.(*td.ChatTypeSupergroup); !ok {
		return 0, errors.New(F(chatID, "cplay_invalid_target"))
	}

	return chat.Id, nil
}

func saveCPlayTarget(c *td.Client, m *td.Message, chatID, channelID int64) error {
	allowed, err := canSetCPlayTarget(c, m, chatID, channelID)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}
	if !allowed {
		m.ReplyText(c, F(chatID, "cplay_owner_required"), nil)
		return nil
	}

	if err := database.LinkChannel(chatID, channelID); err != nil {
		logger.Errorf("Failed to set cplay ID for chat %d: %v", chatID, err)
		m.ReplyText(c, F(chatID, "cplay_save_error"), nil)
		return nil
	}

	m.ReplyText(c, F(chatID, "cplay_enabled", locales.Arg{"channel_id": channelID}), nil)
	return nil
}

func canSetCPlayTarget(c *td.Client, m *td.Message, sourceChatID, targetChatID int64) (bool, error) {
	userID := m.SenderID()
	if isOwnerOrSudo(userID) {
		return true, nil
	}

	sourceOwnerID, err := utils.GetChatOwner(c, sourceChatID)
	if err != nil {
		logger.Errorf("Failed to get source chat owner for %d: %v", sourceChatID, err)
		return false, errors.New(F(sourceChatID, "cplay_owner_check_failed"))
	}
	if sourceOwnerID != 0 && sourceOwnerID == userID {
		return true, nil
	}

	targetOwnerID, err := utils.GetChatOwner(c, targetChatID)
	if err != nil {
		logger.Errorf("Failed to get target chat owner for %d: %v", targetChatID, err)
		return false, errors.New(F(sourceChatID, "cplay_owner_check_failed"))
	}
	return targetOwnerID != 0 && targetOwnerID == userID, nil
}

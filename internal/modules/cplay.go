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

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

const linkedCPlayTarget = "linked"

func setCPlayHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	if len(args) <= 1 {
		m.Reply(F(m.ChannelID(), "cplay_usage"))
		return tg.ErrEndGroup
	}

	chatID := m.ChannelID()
	arg := strings.TrimSpace(args[1])

	enabled, boolErr := utils.ParseBool(arg)
	if boolErr == nil && !enabled {
		return disableCPlay(m, chatID)
	}

	var targetChannelID int64
	var err error

	if strings.EqualFold(arg, linkedCPlayTarget) {
		targetChannelID, err = getLinkedChannelID(m, chatID)
	} else {
		targetChannelID, err = resolveChannelPlay(m, chatID, arg)
	}
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	member, err := m.Client.GetChatMember(targetChannelID, m.Client.Me().ID)
	if err != nil {
		gologging.ErrorF("Failed to fetch bot member state for cplay target %d: %v", targetChannelID, err)
		m.Reply(F(chatID, "cplay_channel_not_accessible"))
		return tg.ErrEndGroup
	}
	if member == nil || (member.Status != tg.Admin && member.Status != tg.Creator) {
		m.Reply(F(chatID, "cplay_channel_not_accessible"))
		return tg.ErrEndGroup
	}
	if member.Status == tg.Admin && (member.Rights == nil || !member.Rights.InviteUsers) {
		m.Reply(F(chatID, "cplay_bot_invite_permission_missing"))
		return tg.ErrEndGroup
	}

	return saveCPlayTarget(m, chatID, targetChannelID)
}


func disableCPlay(m *tg.NewMessage, chatID int64) error {
	allowed, err := canSetCPlayTarget(m, chatID, chatID)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	if !allowed {
		m.Reply(F(chatID, "cplay_owner_required"))
		return tg.ErrEndGroup
	}

	if err := database.LinkChannel(chatID, 0); err != nil {
		gologging.ErrorF("Failed to disable cplay for chat %d: %v", chatID, err)
		m.Reply(F(chatID, "cplay_save_error"))
		return tg.ErrEndGroup
	}

	m.Reply(F(chatID, "cplay_disabled"))
	return tg.ErrEndGroup
}

func getLinkedChannelID(m *tg.NewMessage, chatID int64) (int64, error) {
	peer, err := m.Client.ResolvePeer(chatID)
	if err != nil {
		return 0, errors.New(F(chatID, "cplay_resolve_peer_fail"))
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		full, err := m.Client.ChannelsGetFullChannel(&tg.InputChannelObj{ChannelID: p.ChannelID, AccessHash: p.AccessHash})
		if err != nil || full == nil {
			return 0, errors.New(F(chatID, "cplay_resolve_peer_fail"))
		}
		cf, ok := full.FullChat.(*tg.ChannelFull)
		if !ok || cf.LinkedChatID == 0 {
			return 0, errors.New(F(chatID, "cplay_channel_not_linked"))
		}
		return -100_000_000_0000 - cf.LinkedChatID, nil
	case *tg.InputPeerChat:
		return 0, errors.New(F(chatID, "supergroup_needed", locales.Arg{"chat_id": p.ChatID, "support_group": config.SupportChat}))
	default:
		return 0, errors.New(F(chatID, "cplay_invalid_target"))
	}
}

func resolveChannelPlay(m *tg.NewMessage, chatID int64, target any) (int64, error) {
	peer, err := m.Client.ResolvePeer(target)
	if err != nil {
		gologging.ErrorF("Failed to resolve cplay target %v for chat %d: %v", target, chatID, err)
		return 0, errors.New(F(chatID, "cplay_channel_not_accessible"))
	}

	chPeer, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		return 0, errors.New(F(chatID, "cplay_invalid_target"))
	}

	fullChat, err := m.Client.ChannelsGetFullChannel(&tg.InputChannelObj{ChannelID: chPeer.ChannelID, AccessHash: chPeer.AccessHash})
	if err != nil || fullChat == nil {
		gologging.ErrorF("Failed to get full channel for cplay target %v: %v", target, err)
		return 0, errors.New(F(chatID, "cplay_channel_not_accessible"))
	}

	return -100_000_000_0000 - chPeer.ChannelID, nil
}

func saveCPlayTarget(m *tg.NewMessage, chatID, channelID int64) error {
	allowed, err := canSetCPlayTarget(m, chatID, channelID)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	if !allowed {
		m.Reply(F(chatID, "cplay_owner_required"))
		return tg.ErrEndGroup
	}

	if err := database.LinkChannel(chatID, channelID); err != nil {
		gologging.ErrorF("Failed to set cplay ID for chat %d: %v", chatID, err)
		m.Reply(F(chatID, "cplay_save_error"))
		return tg.ErrEndGroup
	}

	m.Reply(F(chatID, "cplay_enabled", locales.Arg{"channel_id": channelID}))
	return tg.ErrEndGroup
}

func canSetCPlayTarget(m *tg.NewMessage, sourceChatID, targetChatID int64) (bool, error) {
	userID := m.SenderID()
	if isOwnerOrSudo(userID) {
		return true, nil
	}

	sourceOwnerID, err := utils.GetChatOwner(m.Client, sourceChatID)
	if err != nil {
		gologging.ErrorF("Failed to get source chat owner for %d: %v", sourceChatID, err)
		return false, errors.New(F(sourceChatID, "cplay_owner_check_failed"))
	}
	if sourceOwnerID != 0 && sourceOwnerID == userID {
		return true, nil
	}

	targetOwnerID, err := utils.GetChatOwner(m.Client, targetChatID)
	if err != nil {
		gologging.ErrorF("Failed to get target chat owner for %d: %v", targetChatID, err)
		return false, errors.New(F(sourceChatID, "cplay_owner_check_failed"))
	}
	return targetOwnerID != 0 && targetOwnerID == userID, nil
}

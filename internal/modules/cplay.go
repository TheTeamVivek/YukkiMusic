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
	"fmt"
	"strconv"
	"strings"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
)

func setCPlayHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	if len(args) <= 1 {
		m.Reply(F(m.ChannelID(), "cplay_usage"), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	chatID := m.ChannelID()
	channelID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		m.Reply(F(chatID, "cplay_invalid_chat_id"), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	if err := assertChannelAccessible(m, channelID); err != nil {
		m.Reply(err.Error(), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	if err := database.LinkChannel(chatID, channelID); err != nil {
		gologging.ErrorF("Failed to set cplay ID for chat %d: %v", chatID, err)
		m.Reply(F(chatID, "cplay_save_error"), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	m.Reply(
		F(chatID, "cplay_enabled", locales.Arg{"channel_id": channelID}),
		&tg.SendOptions{ParseMode: "HTML"},
	)
	return tg.ErrEndGroup
}

func assertChannelAccessible(m *tg.NewMessage, channelID int64) error {
	chatID := m.ChannelID()

	peer, err := m.Client.ResolvePeer(channelID)
	if err != nil {
		return fmt.Errorf(F(chatID, "cplay_resolve_peer_fail"))
	}

	chPeer, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		return fmt.Errorf(F(chatID, "cplay_invalid_target"))
	}

	fullChat, err := m.Client.ChannelsGetFullChannel(&tg.InputChannelObj{
		ChannelID:  chPeer.ChannelID,
		AccessHash: chPeer.AccessHash,
	})
	if err != nil || fullChat == nil {
		gologging.ErrorF("Failed to get full channel for cplay ID %d: %v", channelID, err)
		return fmt.Errorf(F(chatID, "cplay_channel_not_accessible"))
	}

	return nil
}

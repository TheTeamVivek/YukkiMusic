/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 * ________________________________________________________________________________________
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
 * ________________________________________________________________________________________
 */

package utils

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetFullChannel(
	client *telegram.Client,
	chatID int64,
) (*telegram.ChannelFull, error) {
	peer, err := client.ResolvePeer(chatID)
	if err != nil {
		return nil, err
	}
	chPeer, ok := peer.(*telegram.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf(
			"chatID %d is not an InputPeerChannel, got %T",
			chatID,
			peer,
		)
	}

	fullChat, err := client.ChannelsGetFullChannel(&telegram.InputChannelObj{
		ChannelID:  chPeer.ChannelID,
		AccessHash: chPeer.AccessHash,
	})
	if err != nil {
		return nil, err
	}

	return fullChat.FullChat.(*telegram.ChannelFull), nil
}

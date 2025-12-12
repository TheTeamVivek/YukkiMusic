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
package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func ExtractUser(m *telegram.NewMessage) (int64, error) {
	if m == nil || m.Message == nil {
		return 0, fmt.Errorf("invalid message")
	}

	if m.IsReply() {
		return extractFromReply(m)
	}

	text := m.Text()
	if text == "" {
		return 0, fmt.Errorf("empty message text")
	}

	if id, err := extractFromEntities(m, text); err != nil {
		return 0, err
	} else if id != 0 {
	  return id, nil
	}

	return extractFromPlainText(m, text)
}

// --- Sub Functions ---

func extractFromReply(m *telegram.NewMessage) (int64, error) {
	r, err := m.GetReplyMessage()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch reply message: %w", err)
	}

	if r.Message.FromID == nil {
		return 0, fmt.Errorf("replied message's sender is not a user (may be anon admin)")
	}

	if _, ok := r.Message.FromID.(*telegram.PeerUser); !ok {
		return 0, fmt.Errorf("replied message's sender is not a user (maybe channel/group)")
	}

	return r.SenderID(), nil
}

func extractFromEntities(m *telegram.NewMessage, text string) (int64, error) {
	for _, ent := range m.Message.Entities {
		switch e := ent.(type) {

		// Inline mention (tg://user?id=xxxx)
		case *telegram.MessageEntityMentionName:
			return e.UserID, nil

		// @username mention → resolve peer
		case *telegram.MessageEntityMention:
			username := strings.TrimPrefix(text[e.Offset:e.Offset+e.Length], "@")
			peer, err := m.Client.ResolvePeer(username)
			if err != nil {
				return 0, fmt.Errorf("failed to resolve peer for @%s: %w", username, err)
			}

			userPeer, ok := peer.(*telegram.InputPeerUser)
			if !ok {
				return 0, fmt.Errorf("resolved peer is not a user (maybe channel/group)")
			}

			return userPeer.UserID, nil
		}
	}
	return 0,  nil
}

func extractFromPlainText(m *telegram.NewMessage, text string) (int64, error) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return 0, fmt.Errorf("no user identifier found")
	}

	idStr := parts[1]

	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return id, nil
	}

	peer, err := m.Client.ResolvePeer(idStr)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve peer: %w", err)
	}

	userPeer, ok := peer.(*telegram.InputPeerUser)
	if !ok {
		return 0, fmt.Errorf("resolved peer is not a user (maybe channel/group)")
	}

	return userPeer.UserID, nil
}

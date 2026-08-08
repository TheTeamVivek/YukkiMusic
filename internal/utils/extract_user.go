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

package utils

import (
	"fmt"
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"
)

// ExtractUser extracts a user ID from a message.
// It supports replies, inline mentions, @username mentions and plain text IDs/usernames.
func ExtractUser(c *td.Client, m *td.Message) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("invalid message")
	}

	if m.ReplyToMessageID() > 0 {
		return extractUserFromReply(c, m)
	}

	text := strings.TrimSpace(m.Text())
	if text == "" {
		return 0, fmt.Errorf("empty message text")
	}

	if id, err := extractUserFromEntities(c, m, text); err != nil {
		return 0, err
	} else if id != 0 {
		return id, nil
	}

	return extractUserFromPlainText(c, text)
}

func extractUserFromReply(c *td.Client, m *td.Message) (int64, error) {
	r, err := c.GetMessage(m.ChatID(), m.ReplyToMessageID())
	if err != nil {
		return 0, fmt.Errorf("failed to fetch reply message: %w", err)
	}
	if r == nil || r.SenderId == nil {
		return 0, fmt.Errorf(
			"replied message's sender is not a user (may be anon admin)",
		)
	}
	u, ok := r.SenderId.(*td.MessageSenderUser)
	if !ok {
		return 0, fmt.Errorf(
			"replied message's sender is not a user (maybe channel/group)",
		)
	}
	return u.UserId, nil
}

func extractUserFromEntities(c *td.Client, m *td.Message, text string) (int64, error) {
	for _, ent := range m.Entities() {
		switch e := ent.Type.(type) {
		case *td.TextEntityTypeMentionName:
			return e.UserId, nil
		case *td.TextEntityTypeMention:
			start := int(ent.Offset)
			end := start + int(ent.Length)
			if start < 0 || end > len(text) || start > end {
				continue
			}
			username := strings.TrimPrefix(text[start:end], "@")
			uid, err := resolveUserID(c, username)
			if err != nil {
				return 0, err
			}
			if uid == 0 {
				return 0, fmt.Errorf(
					"resolved peer is not a user (maybe channel/group)",
				)
			}
			return uid, nil
		}
	}
	return 0, nil
}

func extractUserFromPlainText(c *td.Client, text string) (int64, error) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return 0, fmt.Errorf("no user identifier found")
	}

	idStr := strings.TrimPrefix(strings.TrimSpace(parts[1]), "@")
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return id, nil
	}

	uid, err := resolveUserID(c, idStr)
	if err != nil {
		return 0, err
	}
	if uid == 0 {
		return 0, fmt.Errorf(
			"resolved peer is not a user (maybe channel/group)",
		)
	}
	return uid, nil
}

func resolveUserID(c *td.Client, username string) (int64, error) {
	chat, err := c.SearchPublicChat(username)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve peer: %w", err)
	}
	if chat == nil {
		return 0, fmt.Errorf("resolved peer is not a user")
	}
	priv, ok := chat.Type.(*td.ChatTypePrivate)
	if !ok {
		return 0, nil
	}
	return priv.UserId, nil
}

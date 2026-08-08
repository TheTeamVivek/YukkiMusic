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

// ExtractChat extracts a chat ID from a message.
// It supports plain numeric IDs and @username / username identifiers.
func ExtractChat(c *td.Client, m *td.Message) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("invalid message")
	}

	parts := strings.Fields(m.Text())
	if len(parts) < 2 {
		return 0, fmt.Errorf("no chat identifier found")
	}

	target := strings.TrimSpace(strings.TrimPrefix(parts[1], "@"))
	if target == "" {
		return 0, fmt.Errorf("empty chat identifier")
	}

	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}

	chat, err := c.SearchPublicChat(target)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve peer: %w", err)
	}
	if chat == nil {
		return 0, fmt.Errorf("resolved peer is not a chat")
	}
	return chat.Id, nil
}

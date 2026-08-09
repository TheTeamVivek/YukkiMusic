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

	td "github.com/AshokShau/gotdbot"
)

// FullChat bundles a chat with its supergroup full info (when the chat is a
// supergroup or channel).
type FullChat struct {
	Chat               *td.Chat
	SupergroupFullInfo *td.SupergroupFullInfo
}

// GetFullChat fetches the chat (and its supergroup full info) for chatID.
func GetFullChat(c *td.Client, chatID int64) (*FullChat, error) {
	chat, err := c.GetChat(chatID)
	if err != nil {
		return nil, err
	}

	fc := &FullChat{Chat: chat}
	if ct, ok := chat.Type.(*td.ChatTypeSupergroup); ok {
		full, err := c.GetSupergroupFullInfo(ct.SupergroupId)
		if err != nil {
			return nil, fmt.Errorf("get supergroup full info: %w", err)
		}
		fc.SupergroupFullInfo = full
	}
	return fc, nil
}

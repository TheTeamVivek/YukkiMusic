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

	"github.com/amarnathcjd/gogram/telegram"
)

func GetGroupCall(c *telegram.Client, chatID int64) (*telegram.PhoneGroupCall, error) {
	fullChat, err := GetFullChannel(c, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get full channel for chatID %d: %w", chatID, err)
	}

	if fullChat.Call == nil {
		return nil, fmt.Errorf("no active group call found in chatID %d", chatID)
	}

	var call interface{}
	call, err = c.PhoneGetGroupCall(fullChat.Call, 0) // reuse 'err' here
	if err != nil {
		return nil, fmt.Errorf("failed to get group call details for chatID %d: %w", chatID, err)
	}

	gc, ok := call.(*telegram.PhoneGroupCall)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned for group call in chatID %d: got %T", chatID, call)
	}

	return gc, nil
}

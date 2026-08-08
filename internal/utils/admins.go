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
	"slices"
	"time"

	td "github.com/AshokShau/gotdbot"
)

var adminCache = NewCache[int64, []int64](30 * time.Minute)

// IsChatAdmin checks if a user is an admin in a chat.
func IsChatAdmin(c *td.Client, chatID, userID int64) (bool, error) {
	if chatID == userID { // chat anon admin or pvt chat
		return true, nil
	}

	ids, ok := adminCache.Get(chatID)
	if !ok {
		var err error
		ids, err = fetchAdmins(c, chatID)
		if err != nil {
			return false, err
		}
	}

	return slices.Contains(ids, userID), nil
}

// RefreshChatAdmin reloads the chat admins from Telegram and updates the cache.
func RefreshChatAdmin(c *td.Client, chatID int64) ([]int64, error) {
	return fetchAdmins(c, chatID)
}

// GetChatOwner returns the ID of the chat owner (creator).
func GetChatOwner(c *td.Client, chatID int64) (int64, error) {
	admins, err := c.GetChatAdministrators(chatID)
	if err != nil {
		return 0, err
	}

	for _, a := range admins.Administrators {
		if a.IsOwner {
			return a.UserId, nil
		}
	}

	return 0, fmt.Errorf("chat owner not found")
}

// AddChatAdmin adds a user to the cached admin list of a chat without refetching.
func AddChatAdmin(chatID, userID int64) {
	ids, ok := adminCache.Get(chatID)
	if !ok {
		return
	}

	if !slices.Contains(ids, userID) {
		adminCache.Set(chatID, append(ids, userID))
	}
}

// RemoveChatAdmin removes a user from the cached admin list of a chat without refetching.
func RemoveChatAdmin(chatID, userID int64) {
	ids, ok := adminCache.Get(chatID)
	if !ok {
		return
	}

	adminCache.Set(chatID, slices.DeleteFunc(ids, func(id int64) bool {
		return id == userID
	}))
}

// fetchAdmins fetches chat admin IDs from Telegram and caches them.
func fetchAdmins(c *td.Client, chatID int64) ([]int64, error) {
	admins, err := c.GetChatAdministrators(chatID)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(admins.Administrators))
	for _, a := range admins.Administrators {
		if a.UserId != 0 {
			ids = append(ids, a.UserId)
		}
	}

	adminCache.Set(chatID, ids)
	return ids, nil
}

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
	"slices"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

var adminCache = NewCache[string, []int64](30 * time.Minute)

// Checks if a user is an admin in a chat
func IsChatAdmin(c *telegram.Client, chatID, userID int64) (bool, error) {
	cacheKey := fmt.Sprintf("admins:%d", chatID)

if chatID == userID { // chat anon admin 

return true, nil
}
	ids, ok := adminCache.Get(cacheKey)
	if ok {
		return slices.Contains(ids, userID), nil
	}

	ids, err := ReloadChatAdmin(c, chatID)
	if err != nil {
		return false, err
	}

	return slices.Contains(ids, userID), nil
}

// Reloads the chat admins from Telegram and updates the cache
func ReloadChatAdmin(c *telegram.Client, chatID int64) ([]int64, error) {
	ids, err := fetchAdmins(c, chatID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("admins:%d", chatID)
	if len(ids) == 0 {
		adminCache.Delete(cacheKey)
	} else {
		adminCache.Set(cacheKey, ids)
	}

	return ids, nil
}

// Adds a user to the cached admin list, auto-reloading if cache is missing
func AddChatAdmin(c *telegram.Client, chatID, userID int64) error {
	cacheKey := fmt.Sprintf("admins:%d", chatID)

	ids, ok := adminCache.Get(cacheKey)
	if !ok || len(ids) == 0 {
		var err error
		ids, err = ReloadChatAdmin(c, chatID)
		if err != nil {
			return err
		}
	}

	if !slices.Contains(ids, userID) {
		ids = append(ids, userID)
		adminCache.Set(cacheKey, ids)
	}

	return nil
}

// Removes a user from the cached admin list, auto-reloading if cache is missing
func RemoveChatAdmin(c *telegram.Client, chatID, userID int64) error {
	cacheKey := fmt.Sprintf("admins:%d", chatID)

	ids, ok := adminCache.Get(cacheKey)
	if !ok || len(ids) == 0 {
		var err error
		ids, err = ReloadChatAdmin(c, chatID)
		if err != nil {
			return err
		}
	}

	newIDs := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id != userID {
			newIDs = append(newIDs, id)
		}
	}

	if len(newIDs) == 0 {
		adminCache.Delete(cacheKey)
	} else {
		adminCache.Set(cacheKey, newIDs)
	}

	return nil
}

// Fetches admins from Telegram
func fetchAdmins(c *telegram.Client, chatID int64) ([]int64, error) {
	admins, _, err := c.GetChatMembers(chatID, &telegram.ParticipantOptions{
		Filter:           &telegram.ChannelParticipantsAdmins{},
		SleepThresholdMs: 3000,
		Limit:            -1,
	})
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, p := range admins {
		if p.User.Bot || p.User.Deleted {
			continue
		}
		ids = append(ids, p.User.ID)
	}
	return ids, nil
}

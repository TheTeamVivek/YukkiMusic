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
package database

func IsAuthUser(chatID, userID int64) (bool, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	settings, err := getChatSettings(ctx, chatID)
	if err != nil {
		return false, err
	}

	for _, user := range settings.AuthUsers {
		if user == userID {
			return true, nil
		}
	}
	return false, nil
}

func AddAuthUser(chatID, userID int64) error {
	if is, err := IsAuthUser(chatID, userID); is || err != nil {
		return err
	}

	settings, err := getChatSettings(chatID)
	if err != nil {
		return err
	}

	settings.AuthUsers = append(settings.AuthUsers, userID)
	return updateChatSettings(settings)
}

func RemoveAuthUser(chatID, userID int64) error {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return err
	}

	var newAuthUsers []int64
	var found bool
	for _, user := range settings.AuthUsers {
		if user == userID {
			found = true
			continue
		}
		newAuthUsers = append(newAuthUsers, user)
	}

	if !found {
		return nil // User not in the auth list
	}

	settings.AuthUsers = newAuthUsers
	return updateChatSettings(settings)
}

func GetAuthUsers(chatID int64) ([]int64, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return nil, err
	}

	return settings.AuthUsers, nil
}

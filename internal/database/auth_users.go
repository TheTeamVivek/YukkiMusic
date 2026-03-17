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

func IsAuthorized(chatID, userID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return contains(settings.AuthUsers, userID), nil
}

func Authorize(chatID, userID int64) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		var added bool
		s.AuthUsers, added = addUnique(s.AuthUsers, userID)
		return added
	})
}

func Unauthorize(chatID, userID int64) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		var removed bool
		s.AuthUsers, removed = removeElement(s.AuthUsers, userID)
		return removed
	})
}

func AuthorizedUsers(chatID int64) ([]int64, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return nil, err
	}
	return settings.AuthUsers, nil
}

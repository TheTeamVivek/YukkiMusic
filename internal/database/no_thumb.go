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

// GetNoThumb returns whether thumbnails are disabled for the chat.
// Returns false by default (thumbnails enabled).
func GetNoThumb(chatID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return settings.NoThumb, nil
}

// SetNoThumb sets whether thumbnails should be disabled for the chat.
func SetNoThumb(chatID int64, noThumb bool) error {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return err
	}

	if settings.NoThumb == noThumb {
		return nil
	}

	settings.NoThumb = noThumb
	return updateChatSettings(settings)
}

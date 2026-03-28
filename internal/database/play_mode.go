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

package database

func PlayModeAdminsOnly(chatID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return settings.PlayModeAdminsOnly, nil
}

func SetPlayModeAdminsOnly(chatID int64, adminsOnly bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.PlayModeAdminsOnly == adminsOnly {
			return false
		}
		s.PlayModeAdminsOnly = adminsOnly
		return true
	})
}

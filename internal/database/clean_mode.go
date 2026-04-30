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

func CommandDelete(chatID int64) (bool, error) {
	s, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return s.CommandDelete, nil
}

func SetCommandDelete(chatID int64, enabled bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.CommandDelete == enabled {
			return false
		}
		s.CommandDelete = enabled
		return true
	})
}

func CleanMode(chatID int64) (bool, error) {
	s, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return s.CleanMode, nil
}

func SetCleanMode(chatID int64, enabled bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.CleanMode == enabled {
			return false
		}
		s.CleanMode = enabled
		return true
	})
}

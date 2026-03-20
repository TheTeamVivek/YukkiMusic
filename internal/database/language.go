/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 * ________________________________________________________________________________________
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
 * ________________________________________________________________________________________
 */

package database

import "main/internal/config"

func Language(chatID int64) (string, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return config.DefaultLang, err
	}
	if settings.Language == "" {
		return config.DefaultLang, nil
	}
	return settings.Language, nil
}

func SetLanguage(chatID int64, lang string) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.Language == lang {
			return false
		}
		s.Language = lang
		return true
	})
}

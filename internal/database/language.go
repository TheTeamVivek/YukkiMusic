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

import "main/config"

func GetChatLanguage(chatID int64) (string, error) {
	ctx, cancel := mongoCtx()
	defer cancel()
	settings, err := getChatSettings(ctx, chatID)
	if err != nil || settings.Language == "" {
		return config.DefaultLang, err
	}
	return settings.Language, nil
}

func SetChatLanguage(chatID int64, lang string) error {
	ctx, cancel := mongoCtx()
	defer cancel()
	settings, err := getChatSettings(ctx, chatID)
	if err != nil || settings.Language == lang {
		return err
	}
	settings.Language = lang
	return updateChatSettings(ctx, settings)
}

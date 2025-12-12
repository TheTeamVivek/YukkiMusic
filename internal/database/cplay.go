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

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetCPlayID(chatID int64) (int64, error) {
	settings, err := getChatSettings( chatID)
	if err != nil {
		return 0, err
	}
	return settings.CPlayID, nil
}

func SetCPlayID(chatID, cplayID int64) error {
	settings, err := getChatSettings( chatID)
	if err != nil {
		return err
	}

	if settings.CPlayID == cplayID {
		return nil
	}

	// Proactively update cache to prevent stale entries
	if oldCPlayID := settings.CPlayID; oldCPlayID != 0 {
		oldCacheKey := fmt.Sprintf("cplayid_%d", oldCPlayID)
		dbCache.Delete(oldCacheKey)
	}

	settings.CPlayID = cplayID
	err = updateChatSettings( settings)
	if err == nil {
		newCacheKey := fmt.Sprintf("cplayid_%d", cplayID)
		dbCache.Set(newCacheKey, chatID)
	}
	return err
}

func GetChatIDFromCPlayID(cplayID int64) (int64, error) {
	cacheKey := fmt.Sprintf("cplayid_%d", cplayID)
	if cached, found := dbCache.Get(cacheKey); found {
		if chatID, ok := cached.(int64); ok {
			return chatID, nil
		}
	}

	ctx, cancel := mongoCtx()
	defer cancel()

	var settings ChatSettings
	err := chatSettingsColl.FindOne(ctx, bson.M{"cplay_id": cplayID}).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("no chat found with cplayID %d", cplayID)
		}
		return 0, err
	}

	dbCache.Set(cacheKey, settings.ChatID)
	return settings.ChatID, nil
}

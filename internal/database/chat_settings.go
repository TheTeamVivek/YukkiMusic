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
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TODO: reflect deepequal checked removed so handle caching of same opt in high-level
type RTMPConfig struct {
	RtmpURL string `bson:"rtmp_url"`
	RtmpKey string `bson:"rtmp_key"`
}
type ChatSettings struct {
	ChatID         int64      `bson:"_id"`
	CPlayID        int64      `bson:"cplay_id"`
	AuthUsers      []int64    `bson:"auth_users"`
	Language       string     `bson:"language"`
	RTMPConfig     RTMPConfig `bson:"rtmp_config"`
	AssistantIndex int        `bson:"ass_index,omitempty"`
}

func defaultChatSettings(chatID int64) *ChatSettings {
	return &ChatSettings{
		ChatID:    chatID,
		AuthUsers: []int64{},
	}
}

func getChatSettings(chatID int64) (*ChatSettings, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	cacheKey := "chat_settings_" + strconv.FormatInt(chatID, 10)
	if cached, found := dbCache.Get(cacheKey); found {
		if settings, ok := cached.(*ChatSettings); ok {
			return settings, nil
		}
	}

	var settings ChatSettings
	err := chatSettingsColl.FindOne(ctx, bson.M{"_id": chatID}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		def := defaultChatSettings(chatID)
		dbCache.Set(cacheKey, def)
		return def, nil
	} else if err != nil {
		logger.Error("Failed to get chat settings for chat " + strconv.FormatInt(chatID, 10) + " :" + err.Error())
		return nil, err
	}

	dbCache.Set(cacheKey, &settings)

	// Proactively cache the cplayID -> chatID mapping
	if settings.CPlayID != 0 {
		cplayCacheKey := "cplayid_" + strconv.FormatInt(settings.CPlayID, 10)
		dbCache.Set(cplayCacheKey, settings.ChatID)
	}

	return &settings, nil
}

func updateChatSettings(newSettings *ChatSettings) error {
	cacheKey := "chat_settings_" + strconv.FormatInt(newSettings.ChatID, 10)
	ctx, cancel := mongoCtx()
	defer cancel()

	opts := options.UpdateOne().SetUpsert(true)
	_, err := chatSettingsColl.UpdateOne(ctx, bson.M{"_id": newSettings.ChatID}, bson.M{"$set": newSettings}, opts)
	if err != nil {
		logger.Error("Failed to update chat settings for chat " + strconv.FormatInt(newSettings.ChatID, 10) + " :" + err.Error())
		return err
	}

	dbCache.Set(cacheKey, newSettings)
	if newSettings.CPlayID != 0 {
		cplayCacheKey := "cplayid_" + strconv.FormatInt(newSettings.CPlayID, 10)
		dbCache.Set(cplayCacheKey, newSettings.ChatID)
	}

	return nil
}

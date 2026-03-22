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

import (
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RTMPConfig struct {
	URL string `bson:"rtmp_url"`
	Key string `bson:"rtmp_key"`
}

type ChatSettings struct {
	ChatID             int64      `bson:"_id"`
	ChannelPlayID      int64      `bson:"cplay_id"`
	AuthUsers          []int64    `bson:"auth_users"`
	Language           string     `bson:"language"`
	RTMP               RTMPConfig `bson:"rtmp_config"`
	AssistantIndex     int        `bson:"ass_index,omitempty"`
	ThumbnailsDisabled bool       `bson:"no_thumb"`
}

func defaultChatSettings(chatID int64) *ChatSettings {
	return &ChatSettings{
		ChatID:    chatID,
		AuthUsers: []int64{},
	}
}

func getChatSettings(chatID int64) (*ChatSettings, error) {
	cacheKey := "chat_settings_" + strconv.FormatInt(chatID, 10)
	if cached, found := dbCache.Get(cacheKey); found {
		if settings, ok := cached.(*ChatSettings); ok {
			return settings, nil
		}
	}

	ctx, cancel := ctx()
	defer cancel()

	var settings ChatSettings
	err := chatSettingsColl.FindOne(ctx, bson.M{"_id": chatID}).
		Decode(&settings)

	if err == mongo.ErrNoDocuments {
		def := defaultChatSettings(chatID)
		dbCache.Set(cacheKey, def)
		return def, nil
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get chat settings for %d: %w",
			chatID,
			err,
		)
	}

	dbCache.Set(cacheKey, &settings)
	return &settings, nil
}

func updateChatSettings(settings *ChatSettings) error {
	cacheKey := "chat_settings_" + strconv.FormatInt(settings.ChatID, 10)

	ctx, cancel := ctx()
	defer cancel()

	_, err := chatSettingsColl.UpdateOne(
		ctx,
		bson.M{"_id": settings.ChatID},
		bson.M{"$set": settings},
		upsertOpt,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update chat settings for %d: %w",
			settings.ChatID,
			err,
		)
	}

	dbCache.Set(cacheKey, settings)
	return nil
}

func modifyChatSettings(chatID int64, fn func(*ChatSettings) bool) error {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return err
	}

	if fn(settings) {
		return updateChatSettings(settings)
	}

	return nil
}

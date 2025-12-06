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
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TODO: reflect deepequal checked removed so handle caching of same opt in high-level 


type UsersChats struct {
	Users []int64 `bson:"users"`
	Chats []int64 `bson:"chats"`
}
type Maintenance struct {
	Enabled bool   `bson:"enabled,omitempty"`
	Reason  string `bson:"reason,omitempty"`
}

type BotState struct {
	ID            string      `bson:"_id"`
	Served        UsersChats  `bson:"served"`
	Sudoers       []int64     `bson:"sudoers"`
	AutoLeave     bool        `bson:"autoleave"`
	LoggerEnabled bool        `bson:"logger"`
	Maintenance   Maintenance `bson:"maint,omitempty"`
}

const cacheKey = "bot_state"

var defaultBotState = &BotState{
	ID:            "global",
	Served:        UsersChats{Users: []int64{}, Chats: []int64{}},
	Sudoers:       []int64{},
	LoggerEnabled: true,
}

func getBotState(ctx context.Context) (*BotState, error) {
	if cached, found := dbCache.Get(cacheKey); found {
		if state, ok := cached.(*BotState); ok {
			return state, nil
		}
	}

	var state BotState
	err := settingsColl.FindOne(ctx, bson.M{"_id": "global"}).Decode(&state)
	if err == mongo.ErrNoDocuments {
		dbCache.Set(cacheKey, defaultBotState)
		return defaultBotState, nil
	} else if err != nil {
		logger.ErrorF("Failed to get bot state: %v", err)
		return nil, err
	}

	dbCache.Set(cacheKey, &state)
	return &state, nil
}

func updateBotState(ctx context.Context, newState *BotState) error {
	opts := options.UpdateOne().SetUpsert(true)
	_, err = settingsColl.UpdateOne(ctx, bson.M{"_id": "global"}, bson.M{"$set": newState}, opts)
	if err != nil {
		logger.ErrorF("Failed to update bot state: %v", err)
		return err
	}

	dbCache.Set(cacheKey, newState)
	return nil
}

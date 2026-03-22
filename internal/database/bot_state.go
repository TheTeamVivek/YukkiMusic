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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

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

	// runtime indexes for fast lookup
	servedUsersMap map[int64]struct{} `bson:"-"`
	servedChatsMap map[int64]struct{} `bson:"-"`
}

const botStateCacheKey = "bot_state"

func newDefaultBotState() *BotState {
	s := &BotState{
		ID: "global",
		Served: UsersChats{
			Users: []int64{},
			Chats: []int64{},
		},
		Sudoers:       []int64{},
		LoggerEnabled: true,
	}

	buildIndexes(s)
	return s
}

func getBotState() (*BotState, error) {
	if cached, found := dbCache.Get(botStateCacheKey); found {
		if state, ok := cached.(*BotState); ok {
			return state, nil
		}
	}

	ctx, cancel := ctx()
	defer cancel()

	var state BotState
	err := settingsColl.FindOne(ctx, bson.M{"_id": "global"}).Decode(&state)

	if err == mongo.ErrNoDocuments {
		s := newDefaultBotState()
		dbCache.Set(botStateCacheKey, s)
		return s, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get bot state: %w", err)
	}

	buildIndexes(&state)
	dbCache.Set(botStateCacheKey, &state)
	return &state, nil
}

func updateBotState(newState *BotState) error {
	ctx, cancel := ctx()
	defer cancel()

	_, err := settingsColl.UpdateOne(
		ctx,
		bson.M{"_id": "global"},
		bson.M{"$set": newState},
		upsertOpt,
	)
	if err != nil {
		return fmt.Errorf("failed to update bot state: %w", err)
	}

	dbCache.Set(botStateCacheKey, newState)
	return nil
}

func modifyBotState(fn func(*BotState) bool) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if fn(state) {
		return updateBotState(state)
	}

	return nil
}

func buildIndexes(s *BotState) {
	s.servedUsersMap = make(map[int64]struct{}, len(s.Served.Users))
	for _, u := range s.Served.Users {
		s.servedUsersMap[u] = struct{}{}
	}

	s.servedChatsMap = make(map[int64]struct{}, len(s.Served.Chats))
	for _, c := range s.Served.Chats {
		s.servedChatsMap[c] = struct{}{}
	}
}

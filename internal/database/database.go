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
	"context"
	"time"

	"github.com/Laky-64/gologging"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/internal/utils"
)

var (
	client           *mongo.Client
	database         *mongo.Database
	settingsColl     *mongo.Collection
	chatSettingsColl *mongo.Collection

	logger            = gologging.GetLogger("Database")
	dbCache           = utils.NewCache[string, any](60 * time.Minute)
	chatSettingsCache = utils.NewCache[int64, *ChatSettings](60 * time.Minute)
)

func Init(mongoURL string) (func(), error) {
	var err error
	logger.Debug("Initializing MongoDB...")
	client, err = mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	logger.Debug("Successfully connected to MongoDB.")

	database = client.Database("YukkiMusic")
	settingsColl = database.Collection("bot_settings")
	chatSettingsColl = database.Collection("chat_settings")

	migrateData()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("Error while disconnecting MongoDB: %v", err)
		} else {
			logger.Info("MongoDB disconnected successfully")
		}
	}, nil
}

func GetMongoDBStats() (bson.M, error) {
	ctx, cancel := ctx()
	defer cancel()
	var result bson.M
	err := database.RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&result)
	return result, err
}

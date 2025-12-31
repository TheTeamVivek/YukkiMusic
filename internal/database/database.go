/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package database

import (
	"time"

	"github.com/Laky-64/gologging"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/internal/utils"
)

var (
	client           *mongo.Client
	database         *mongo.Database
	settingsColl     *mongo.Collection
	chatSettingsColl *mongo.Collection

	logger  = gologging.GetLogger("Database")
	dbCache = utils.NewCache[string, any](60 * time.Minute)
)

func Init(mongoURL string) func() {
	var err error
	logger.Debug("Initializing MongoDB...")
	client, err = mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: %v", err)
	}

	logger.Debug("Successfully connected to MongoDB.")

	database = client.Database("YukkiMusic")
	settingsColl = database.Collection("bot_settings")
	chatSettingsColl = database.Collection("chat_settings")

	go migrateData(mongoURL)

	return func() {
		ctx, cancel := mongoCtx()
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("Error while disconnecting MongoDB: %v", err)
		} else {
			logger.Info("MongoDB disconnected successfully")
		}
	}
}

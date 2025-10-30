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
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	oldDBName = "Yukki"
)

// old data structures
type (
	oldCPlay struct {
		ChatID int64 `bson:"chat_id"`
		Mode   int64 `bson:"mode"`
	}
	oldServedUser struct {
		UserID int64 `bson:"user_id"`
	}

	oldServedChat struct {
		ChatID int64 `bson:"chat_id"`
	}

	oldSudoers struct {
		Sudoers []int64 `bson:"sudoers"`
	}
)

func MigrateData(mongoURI string) {
	logger.Info("Checking for old database to migrate...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oldClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.ErrorF("Failed to connect to old MongoDB: %v", err)
		return
	}
	defer oldClient.Disconnect(ctx)

	oldDB := oldClient.Database(oldDBName)

	migrateCPlay(oldDB)
	migrateServedUsers(oldDB)
	migrateServedChats(oldDB)
	migrateSudoers(oldDB)

	logger.Info("Data migration check complete.")
}

func migrateCPlay(db *mongo.Database) {
	coll := db.Collection("cplaymode")
	ctx, cancel := mongoCtx()
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old cplaymode collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var old oldCPlay
		if err := cursor.Decode(&old); err != nil {
			logger.ErrorF("Failed to decode old cplay document: %v", err)
			continue
		}
		if err := SetCPlayID(old.ChatID, old.Mode); err != nil {
			logger.ErrorF("Failed to migrate cplay for chat %d: %v", old.ChatID, err)
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logger.ErrorF("Failed to drop cplaymode collection: %v", err)
	}

	logger.Info("Finished migrating cplay settings.")
}

func migrateServedUsers(db *mongo.Database) {
	coll := db.Collection("tgusersdb")
	ctx, cancel := mongoCtx()
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"user_id": bson.M{"$gt": 0}})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old tgusersdb collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var old oldServedUser
		if err := cursor.Decode(&old); err != nil {
			logger.ErrorF("Failed to decode old user document: %v", err)
			continue
		}
		if err := AddServed(old.UserID, true); err != nil {
			logger.ErrorF("Failed to migrate served user %d: %v", old.UserID, err)
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logger.ErrorF("Failed to drop tgusersdb collection: %v", err)
	}

	logger.Info("Finished migrating served users.")
}

func migrateServedChats(db *mongo.Database) {
	coll := db.Collection("chats")
	ctx, cancel := mongoCtx()
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"chat_id": bson.M{"$lt": 0}})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old chats collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var old oldServedChat
		if err := cursor.Decode(&old); err != nil {
			logger.ErrorF("Failed to decode old chat document: %v", err)
			continue
		}
		if err := AddServed(old.ChatID, false); err != nil {
			logger.ErrorF("Failed to migrate served chat %d: %v", old.ChatID, err)
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logger.ErrorF("Failed to drop chats collection: %v", err)
	}

	logger.Info("Finished migrating served chats.")
}

func migrateSudoers(db *mongo.Database) {
	coll := db.Collection("sudoers")
	ctx, cancel := mongoCtx()
	defer cancel()

var doc oldSudoers
	err := coll.FindOne(ctx, bson.M{"sudo": "sudo"}).Decode(&doc)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old sudoers collection: %v", err)
		}
		return
	}

	for _, sudoerID := range doc.Sudoers {
		if err := AddSudo(sudoerID); err != nil {
			logger.ErrorF("Failed to migrate sudoer %d: %v", sudoerID, err)
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logger.ErrorF("Failed to drop empty sudoers collection: %v", err)
	}

	logger.Info("Finished migrating sudoers.")
}

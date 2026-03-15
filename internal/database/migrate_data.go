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

func migrateData() {
	logger.Info("Checking for old database to migrate...")

	oldDB := client.Database(oldDBName)
	ctx, cancel := mongoCtx()
	defer cancel()

	flagColl := oldDB.Collection("migration_status")
	var result bson.M
	err := flagColl.FindOne(ctx, bson.M{"migrated": true}).Decode(&result)
	if err == nil {
		logger.Info("Migration already completed previously. Skipping.")
		return
	}

	// perform migration
	migrateCPlay(oldDB)
	migrateServedUsers(oldDB)
	migrateServedChats(oldDB)
	migrateSudoers(oldDB)

	_, err = flagColl.InsertOne(ctx, bson.M{
		"migrated":  true,
		"timestamp": time.Now(),
	})
	if err != nil {
		logger.ErrorF("Failed to write migration flag: %v", err)
	}
	logger.Info("Data migration complete.")
}

func migrateCPlay(db *mongo.Database) {
	coll := db.Collection("cplaymode")

	ctx, cancel := mongoCtx()
	defer cancel()

	opts := options.Find().SetBatchSize(500)

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old cplaymode collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	var models []mongo.WriteModel

	for cursor.Next(ctx) {
		var old oldCPlay

		if err := cursor.Decode(&old); err != nil {
			logger.ErrorF("Failed to decode old cplay document: %v", err)
			continue
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": old.ChatID}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"cplay_id": old.Mode,
				},
			}).
			SetUpsert(true)

		models = append(models, model)

		if len(models) >= 1000 {
			_, err := chatSettingsColl.BulkWrite(ctx, models)
			if err != nil {
				logger.ErrorF("Bulk migration for cplay failed: %v", err)
				return
			}

			models = models[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		logger.ErrorF("Cursor error while migrating cplay: %v", err)
		return
	}

	if len(models) > 0 {
		_, err := chatSettingsColl.BulkWrite(ctx, models)
		if err != nil {
			logger.ErrorF("Bulk migration for cplay failed: %v", err)
			return
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

	opts := options.Find().SetBatchSize(500)

	cursor, err := coll.Find(ctx, bson.M{"user_id": bson.M{"$gt": 0}}, opts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old tgusersdb collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	var users []int64

	for cursor.Next(ctx) {
		var old oldServedUser

		if err := cursor.Decode(&old); err != nil {
			logger.ErrorF("Failed to decode old user document: %v", err)
			continue
		}

		users = append(users, old.UserID)
	}

	if err := cursor.Err(); err != nil {
		logger.ErrorF("Cursor error while migrating served users: %v", err)
		return
	}

	if len(users) > 0 {
		_, err := settingsColl.UpdateOne(
			ctx,
			bson.M{"_id": "global"},
			bson.M{
				"$addToSet": bson.M{
					"served.users": bson.M{
						"$each": users,
					},
				},
			},
			upsertOpt,
		)
		if err != nil {
			logger.ErrorF("Bulk migration for served users failed: %v", err)
			return
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

	opts := options.Find().SetBatchSize(500)

	cursor, err := coll.Find(ctx, bson.M{"chat_id": bson.M{"$lt": 0}}, opts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.ErrorF("Failed to query old chats collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	var chats []int64

	for cursor.Next(ctx) {
		var old oldServedChat

		if err := cursor.Decode(&old); err != nil {
			logger.ErrorF("Failed to decode old chat document: %v", err)
			continue
		}

		chats = append(chats, old.ChatID)
	}

	if err := cursor.Err(); err != nil {
		logger.ErrorF("Cursor error while migrating served chats: %v", err)
		return
	}

	if len(chats) > 0 {
		_, err := settingsColl.UpdateOne(
			ctx,
			bson.M{"_id": "global"},
			bson.M{
				"$addToSet": bson.M{
					"served.chats": bson.M{
						"$each": chats,
					},
				},
			},
			upsertOpt,
		)
		if err != nil {
			logger.ErrorF("Bulk migration for served chats failed: %v", err)
			return
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

	if len(doc.Sudoers) > 0 {
		_, err := settingsColl.UpdateOne(
			ctx,
			bson.M{"_id": "global"},
			bson.M{
				"$addToSet": bson.M{
					"sudoers": bson.M{
						"$each": doc.Sudoers,
					},
				},
			},
			upsertOpt,
		)
		if err != nil {
			logger.ErrorF("Bulk migration for sudoers failed: %v", err)
			return
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logger.ErrorF("Failed to drop empty sudoers collection: %v", err)
	}

	logger.Info("Finished migrating sudoers.")
}

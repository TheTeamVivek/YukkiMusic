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

	"yukkimusic/config"

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
	logr.Info("Checking for old database to migrate...")

	if oldDBName == config.DBName {
		logr.Info("Old database name matches the current database name. Skipping migration.")
		return
	}

	oldDB := client.Database(oldDBName)
	ctx := context.Background()

	flagColl := oldDB.Collection("migration_status")
	var result bson.M
	err := flagColl.FindOne(ctx, bson.M{"migrated": true}).Decode(&result)

	if err == nil {
		logr.Info("Migration already completed previously. Skipping.")
		return
	}

	// perform migration
	migrateCPlay(ctx, oldDB)
	migrateServedUsers(ctx, oldDB)
	migrateServedChats(ctx, oldDB)
	migrateSudoers(ctx, oldDB)

	_, err = flagColl.InsertOne(ctx, bson.M{
		"migrated":  true,
		"timestamp": time.Now(),
	})
	if err != nil {
		logr.Errorf("Failed to write migration flag: %v", err)
	}
	logr.Info("Data migration complete.")
}

func migrateCPlay(ctx context.Context, db *mongo.Database) {
	coll := db.Collection("cplaymode")

	opts := options.Find().SetBatchSize(500)

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logr.Errorf("Failed to query old cplaymode collection: %v", err)
		}
		return
	}
	defer cursor.Close(ctx)

	var models []mongo.WriteModel

	for cursor.Next(ctx) {
		var old oldCPlay

		if err := cursor.Decode(&old); err != nil {
			logr.Errorf("Failed to decode old cplay document: %v", err)
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
				logr.Errorf("Bulk migration for cplay failed: %v", err)
				return
			}

			models = models[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		logr.Errorf("Cursor error while migrating cplay: %v", err)
		return
	}

	if len(models) > 0 {
		_, err := chatSettingsColl.BulkWrite(ctx, models)
		if err != nil {
			logr.Errorf("Bulk migration for cplay failed: %v", err)
			return
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logr.Errorf("Failed to drop cplaymode collection: %v", err)
	}

	logr.Info("Finished migrating cplay settings.")
}

func migrateServedUsers(ctx context.Context, db *mongo.Database) {
	migrateServedIDs(
		ctx,
		db,
		"tgusersdb",
		"served.users",
		"user",
		"users",
		bson.M{"user_id": bson.M{"$gt": 0}},
		func(u oldServedUser) int64 { return u.UserID },
	)
}

func migrateServedChats(ctx context.Context, db *mongo.Database) {
	migrateServedIDs(
		ctx,
		db,
		"chats",
		"served.chats",
		"chat",
		"chats",
		bson.M{"chat_id": bson.M{"$lt": 0}},
		func(c oldServedChat) int64 { return c.ChatID },
	)
}

// migrateServedIDs copies IDs from an old collection into the global settings
// document and drops the old collection. It is shared by the served-users and
// served-chats migrations.
func migrateServedIDs[T any](
	ctx context.Context,
	db *mongo.Database,
	collName, target, docLabel, plural string,
	filter bson.M,
	extract func(T) int64,
) {
	coll := db.Collection(collName)

	opts := options.Find().SetBatchSize(500)

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logr.Errorf("Failed to query old %s collection: %v", collName, err)
		}
		return
	}
	defer cursor.Close(ctx)

	var ids []int64

	for cursor.Next(ctx) {
		var old T

		if err := cursor.Decode(&old); err != nil {
			logr.Errorf("Failed to decode old %s document: %v", docLabel, err)
			continue
		}

		ids = append(ids, extract(old))
	}

	if err := cursor.Err(); err != nil {
		logr.Errorf("Cursor error while migrating served %s: %v", plural, err)
		return
	}

	if len(ids) > 0 {
		_, err := settingsColl.UpdateOne(
			ctx,
			bson.M{"_id": "global"},
			bson.M{
				"$addToSet": bson.M{
					target: bson.M{
						"$each": ids,
					},
				},
			},
			upsertOpt,
		)
		if err != nil {
			logr.Errorf("Bulk migration for served %s failed: %v", plural, err)
			return
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logr.Errorf("Failed to drop %s collection: %v", collName, err)
	}

	logr.Infof("Finished migrating served %s.", plural)
}

func migrateSudoers(ctx context.Context, db *mongo.Database) {
	coll := db.Collection("sudoers")

	var doc oldSudoers
	err := coll.FindOne(ctx, bson.M{"sudo": "sudo"}).Decode(&doc)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logr.Errorf("Failed to query old sudoers collection: %v", err)
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
			logr.Errorf("Bulk migration for sudoers failed: %v", err)
			return
		}
	}

	if err := coll.Drop(ctx); err != nil {
		logr.Errorf("Failed to drop empty sudoers collection: %v", err)
	}

	logr.Info("Finished migrating sudoers.")
}

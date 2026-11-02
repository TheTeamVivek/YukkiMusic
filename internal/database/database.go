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
	"sort"
	"time"

	"yukkimusic/config"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"yukkimusic/internal/utils"
)

var (
	client           *mongo.Client
	database         *mongo.Database
	settingsColl     *mongo.Collection
	chatSettingsColl *mongo.Collection

	logr              = utils.Log.With("logger", "database")
	dbCache           = utils.NewCache[string, any](60 * time.Minute)
	chatSettingsCache = utils.NewCache[int64, *ChatSettings](60 * time.Minute)
)

func Init(mongoURL string) (func(), error) {
	var err error
	logr.Debug("Initializing MongoDB...")
	client, err = mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	logr.Debug("Successfully connected to MongoDB.")

	database = client.Database(config.DBName)
	settingsColl = database.Collection("bot_settings")
	chatSettingsColl = database.Collection("chat_settings")

	migrateData()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logr.Errorf("Error while disconnecting MongoDB: %v", err)
		} else {
			logr.Info("MongoDB disconnected successfully")
		}
	}, nil
}

// MongoDatabaseStats holds the on-disk size of a single database.
type MongoDatabaseStats struct {
	Name       string
	SizeOnDisk int64
}

// MongoCollectionStats holds usage of a single collection.
type MongoCollectionStats struct {
	Name        string
	Objects     int64
	StorageSize int64
}

// MongoStats is a display-ready snapshot of MongoDB usage.
type MongoStats struct {
	Databases       []MongoDatabaseStats
	TotalDatabases  int
	CollectionStats []MongoCollectionStats
}

// GetMongoDBStats collects MongoDB usage. Each part is best-effort: a command
// that fails (e.g. missing privileges) simply leaves that part empty.
func GetMongoDBStats() *MongoStats {
	ctx, cancel := ctx()
	defer cancel()

	s := &MongoStats{}

	var list struct {
		Databases []struct {
			Name       string `bson:"name"`
			SizeOnDisk int64  `bson:"sizeOnDisk"`
		} `bson:"databases"`
	}
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "listDatabases", Value: 1}}).Decode(&list); err == nil {
		for _, db := range list.Databases {
			if db.SizeOnDisk > 0 && !isInternalDatabase(db.Name) {
				s.Databases = append(s.Databases, MongoDatabaseStats{Name: db.Name, SizeOnDisk: db.SizeOnDisk})
			}
		}
		sort.Slice(s.Databases, func(i, j int) bool {
			return s.Databases[i].SizeOnDisk > s.Databases[j].SizeOnDisk
		})
		s.TotalDatabases = len(s.Databases)
	}

	if names, err := database.ListCollectionNames(ctx, bson.D{}); err == nil {
		for _, name := range names {
			var coll struct {
				Count       int64 `bson:"count"`
				StorageSize int64 `bson:"storageSize"`
			}
			if err := database.RunCommand(ctx, bson.D{{Key: "collStats", Value: name}}).Decode(&coll); err != nil {
				continue
			}
			s.CollectionStats = append(s.CollectionStats, MongoCollectionStats{
				Name:        name,
				Objects:     coll.Count,
				StorageSize: coll.StorageSize,
			})
		}
		sort.Slice(s.CollectionStats, func(i, j int) bool {
			return s.CollectionStats[i].StorageSize > s.CollectionStats[j].StorageSize
		})
	}

	return s
}

// isInternalDatabase reports whether name is a MongoDB-managed database that
// should not be counted or displayed.
func isInternalDatabase(name string) bool {
	switch name {
	case "admin", "local", "config":
		return true
	default:
		return false
	}
}

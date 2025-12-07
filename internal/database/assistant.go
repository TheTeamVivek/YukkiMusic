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
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetAssistantIndex(chatID int64, assistantCount int) (int, error) {
	if assistantCount <= 0 {
		logger.Error("assistantCount must be positive")
		return 0, fmt.Errorf("assistantCount must be positive")
	}

	ctx, cancel := mongoCtx()
	defer cancel()

	settings, err := getChatSettings(ctx, chatID)
	if err != nil {
		logger.Error("Failed to get chat settings for chat " + strconv.FormatInt(chatID, 10) + ": " + err.Error())
		return 0, err
	}

	if settings.AssistantIndex >= 1 && settings.AssistantIndex <= assistantCount {
		return settings.AssistantIndex, nil
	}

	counts, err := aggregateAssistantUsage(ctx, assistantCount)
	if err != nil {
		return 0, err
	}

	newIndex := pickLeastUsedAssistant(counts)

	logger.Debug(
		"Assigning assistant index " + strconv.Itoa(newIndex) +
			" to chat " + strconv.FormatInt(chatID, 10),
	)

	settings.AssistantIndex = newIndex
	if err := updateChatSettings(ctx, settings); err != nil {
		logger.Error(
			"Failed to update assistant index for chat " +
				strconv.FormatInt(chatID, 10) + ": " + err.Error(),
		)
		return 0, err
	}

	return newIndex, nil
}

func RebalanceAssistantIndexes(assistantCount int) error {
	if assistantCount <= 0 {
		logger.Error("assistantCount must be positive")
		return fmt.Errorf("assistantCount must be positive")
	}

	ctx, cancel := mongoCtx()
	defer cancel()

	cursor, err := chatSettingsColl.Find(ctx, bson.M{})
	if err != nil {
		logger.Error("Failed to fetch chat settings for rebalance: " + err.Error())
		return err
	}
	defer cursor.Close(ctx)

	var all []*ChatSettings
	original := make(map[int64]int)

	for cursor.Next(ctx) {
		var s ChatSettings
		if err := cursor.Decode(&s); err != nil {
			logger.Error("Failed to decode chat setting during rebalance: " + err.Error())
			return err
		}
		all = append(all, &s)
		original[s.ChatID] = s.AssistantIndex
	}

	if err := cursor.Err(); err != nil {
		logger.Error("Rebalance cursor error: " + err.Error())
		return err
	}

	total := len(all)
	if total == 0 {
		logger.Debug("Rebalance: no chats found")
		return nil
	}

	base := total / assistantCount
	rem := total % assistantCount

	desired := make([]int, assistantCount+1)
	for i := 1; i <= assistantCount; i++ {
		desired[i] = base
		if i <= rem {
			desired[i]++
		}
	}

	currentCounts := make([]int, assistantCount+1)
	var unassigned []*ChatSettings

	for _, s := range all {
		idx := s.AssistantIndex
		if idx < 1 || idx > assistantCount {
			unassigned = append(unassigned, s)
			continue
		}
		currentCounts[idx]++
	}

	keepCount := make([]int, assistantCount+1)
	var excess []*ChatSettings

	for _, s := range all {
		idx := s.AssistantIndex
		if idx < 1 || idx > assistantCount {
			continue
		}
		if keepCount[idx] < desired[idx] {
			keepCount[idx]++
		} else {
			excess = append(excess, s)
		}
	}

	pool := make([]*ChatSettings, 0, len(excess)+len(unassigned))
	pool = append(pool, excess...)
	pool = append(pool, unassigned...)

	poolIndex := 0
	for i := 1; i <= assistantCount; i++ {
		need := desired[i] - keepCount[i]
		for need > 0 && poolIndex < len(pool) {
			s := pool[poolIndex]
			poolIndex++
			if s.AssistantIndex == i {
				continue
			}
			s.AssistantIndex = i
			need--
		}
	}

	updated := 0

	for _, s := range all {
		oldIdx := original[s.ChatID]
		if s.AssistantIndex == oldIdx {
			continue
		}

		logger.Debug(
			"Rebalance: updating chat " + strconv.FormatInt(s.ChatID, 10) +
				" from index " + strconv.Itoa(oldIdx) +
				" to " + strconv.Itoa(s.AssistantIndex),
		)

		if err := updateChatSettings(ctx, s); err != nil {
			logger.Error(
				"Rebalance: failed updating chat " +
					strconv.FormatInt(s.ChatID, 10) + ": " + err.Error(),
			)
			return err
		}
		updated++
	}

	logger.Debug(
		"Rebalance complete. total_chats=" + strconv.Itoa(total) +
			", updated_chats=" + strconv.Itoa(updated),
	)

	return nil
}

func aggregateAssistantUsage(ctx context.Context, assistantCount int) ([]int64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{
			"$group", bson.D{
				{"_id", "$ass_index"},
				{"count", bson.D{{"$sum", 1}}},
			},
		}},
	}

	cursor, err := chatSettingsColl.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("Failed to aggregate assistant usage: " + err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make([]int64, assistantCount+1)

	for cursor.Next(ctx) {
		var doc struct {
			ID    *int  `bson:"_id"`
			Count int64 `bson:"count"`
		}

		if err := cursor.Decode(&doc); err != nil {
			logger.Error("Failed to decode aggregation result: " + err.Error())
			return nil, err
		}

		if doc.ID == nil {
			continue
		}

		idx := *doc.ID
		if idx >= 1 && idx <= assistantCount {
			counts[idx] = doc.Count
		}
	}

	if err := cursor.Err(); err != nil {
		logger.Error("Cursor error during assistant usage aggregation: " + err.Error())
		return nil, err
	}

	return counts, nil
}

func pickLeastUsedAssistant(counts []int64) int {
	newIndex := 1
	minCount := counts[1]

	for i := 2; i < len(counts); i++ {
		if counts[i] < minCount {
			minCount = counts[i]
			newIndex = i
		}
	}

	return newIndex
}

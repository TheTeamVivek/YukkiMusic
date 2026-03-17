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
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	assistantUsage []int64 // index 1..assistantCount used
	usageMu        sync.RWMutex
)

func AssistantIndex(chatID int64, assistantCount int) (int, error) {
	if assistantCount <= 0 {
		return 0, fmt.Errorf("assistantCount must be positive")
	}

	settings, err := getChatSettings(chatID)
	if err != nil {
		return 0, err
	}

	if settings.AssistantIndex >= 1 &&
		settings.AssistantIndex <= assistantCount {
		return settings.AssistantIndex, nil
	}

	usageMu.RLock()
	countsCopy := make([]int64, len(assistantUsage))
	copy(countsCopy, assistantUsage)
	usageMu.RUnlock()

	newIndex := pickLeastUsedAssistant(countsCopy)

	settings.AssistantIndex = newIndex
	if err := updateChatSettings(settings); err != nil {
		return 0, err
	}

	usageMu.Lock()
	if len(assistantUsage) > newIndex {
		assistantUsage[newIndex]++
	}
	usageMu.Unlock()

	return newIndex, nil
}

func RebalanceAssistantIndexes(assistantCount int) error {
	if assistantCount <= 0 {
		return fmt.Errorf("assistantCount must be positive")
	}

	all, err := fetchAllChatSettings()
	if err != nil {
		return err
	}

	if len(all) == 0 {
		usageMu.Lock()
		assistantUsage = make([]int64, assistantCount+1)
		usageMu.Unlock()
		return nil
	}

	redistributeAssistants(all, assistantCount)

	if err := saveChangedSettings(all); err != nil {
		return err
	}

	updateAssistantUsage(all, assistantCount)
	return nil
}

func fetchAllChatSettings() ([]*ChatSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cursor, err := chatSettingsColl.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat settings: %w", err)
	}
	defer cursor.Close(ctx)

	var all []*ChatSettings
	if err := cursor.All(ctx, &all); err != nil {
		return nil, fmt.Errorf("failed to decode chat settings: %w", err)
	}

	return all, nil
}

func redistributeAssistants(all []*ChatSettings, assistantCount int) {
	total := len(all)
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
	keepCount := make([]int, assistantCount+1)
	var pool []*ChatSettings

	// Phase 1: Identify who can stay
	for _, s := range all {
		idx := s.AssistantIndex
		if idx >= 1 && idx <= assistantCount && keepCount[idx] < desired[idx] {
			keepCount[idx]++
			currentCounts[idx]++
		} else {
			pool = append(pool, s)
		}
	}

	// Phase 2: Assign from pool to fill gaps
	poolIdx := 0
	for i := 1; i <= assistantCount; i++ {
		for keepCount[i] < desired[i] && poolIdx < len(pool) {
			pool[poolIdx].AssistantIndex = i
			keepCount[i]++
			poolIdx++
		}
	}
}

func saveChangedSettings(all []*ChatSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, s := range all {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": s.ChatID}).
			SetUpdate(bson.M{"$set": bson.M{"ass_index": s.AssistantIndex}}))

		if len(models) >= 500 {
			if _, err := chatSettingsColl.BulkWrite(ctx, models); err != nil {
				return fmt.Errorf("bulk update failed: %w", err)
			}
			models = nil
		}
	}

	if len(models) > 0 {
		if _, err := chatSettingsColl.BulkWrite(ctx, models); err != nil {
			return fmt.Errorf("bulk update failed: %w", err)
		}
	}
	return nil
}

func updateAssistantUsage(all []*ChatSettings, assistantCount int) {
	counts := make([]int64, assistantCount+1)
	for _, s := range all {
		if s.AssistantIndex >= 1 && s.AssistantIndex <= assistantCount {
			counts[s.AssistantIndex]++
		}
	}

	usageMu.Lock()
	assistantUsage = counts
	usageMu.Unlock()
}

func pickLeastUsedAssistant(counts []int64) int {
	if len(counts) <= 1 {
		return 1
	}
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

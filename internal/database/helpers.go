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

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var upsertOpt = options.UpdateOne().SetUpsert(true)

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// addUnique adds an element to a slice if it's not already present.
// Returns the new slice and true if the element was added.
func addUnique[T comparable](slice []T, element T) ([]T, bool) {
	for _, v := range slice {
		if v == element {
			return slice, false
		}
	}
	return append(slice, element), true
}

// removeElement removes an element from a slice if it's present.
// Returns the new slice and true if the element was removed.
func removeElement[T comparable](slice []T, element T) ([]T, bool) {
	for i, v := range slice {
		if v == element {
			return append(slice[:i], slice[i+1:]...), true
		}
	}
	return slice, false
}

// contains checks if a slice contains an element.
func contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

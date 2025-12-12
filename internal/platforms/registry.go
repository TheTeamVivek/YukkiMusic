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
package platforms

import (
	"sort"
	"sync"

	state "main/internal/core/models"
)

type (
	regEntry struct {
		platform state.Platform
		priority int
	}
)

var (
	registry = make(map[state.PlatformName]regEntry)
	regLock  sync.RWMutex
)

func addPlatform(priority int, name state.PlatformName, p state.Platform) {
	regLock.Lock()
	defer regLock.Unlock()
	registry[name] = regEntry{
		platform: p,
		priority: priority,
	}
}

func getOrderedPlatforms() []state.Platform {
	platforms := make([]regEntry, 0, len(registry))
	for _, entry := range registry {
		platforms = append(platforms, entry)
	}

	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].priority > platforms[j].priority
	})

	result := make([]state.Platform, len(platforms))
	for i, entry := range platforms {
		result[i] = entry.platform
	}
	return result
}

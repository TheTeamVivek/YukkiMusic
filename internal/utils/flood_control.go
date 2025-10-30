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
package utils

import (
	"sync"
	"time"
)

var (
	floodMap = make(map[string]time.Time)
	floodMu  sync.Mutex
)

// GetFlood returns the remaining cooldown time for a key.
// If zero or negative, the action is allowed.
func GetFlood(key string) time.Duration {
	floodMu.Lock()
	defer floodMu.Unlock()

	if t, exists := floodMap[key]; exists {
		return time.Until(t)
	}
	return 0
}

// SetFlood sets a flood timeout for the key.
// 'duration' specifies how long the key should be blocked.
func SetFlood(key string, duration time.Duration) {
	floodMu.Lock()
	defer floodMu.Unlock()

	floodMap[key] = time.Now().Add(duration)
}

// CanAct returns true if the key is allowed to act (cooldown expired).
func CanAct(key string) bool {
	return GetFlood(key) <= 0
}

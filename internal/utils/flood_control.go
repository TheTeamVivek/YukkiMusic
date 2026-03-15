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

package utils

import (
	"sync"
	"time"
)

var (
	floodMap = make(map[string]time.Time)
	floodMu  sync.RWMutex
)

// GetFlood returns remaining cooldown for a key.
func GetFlood(key string) time.Duration {
	floodMu.RLock()
	t, ok := floodMap[key]
	floodMu.RUnlock()

	if !ok {
		return 0
	}

	remaining := time.Until(t)

	if remaining <= 0 {
		floodMu.Lock()
		delete(floodMap, key)
		floodMu.Unlock()
	}

	return remaining
}

// SetFlood sets cooldown duration for a key.
func SetFlood(key string, duration time.Duration) {
	floodMu.Lock()
	floodMap[key] = time.Now().Add(duration)
	floodMu.Unlock()
}

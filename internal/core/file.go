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
package core

import (
	"os"
	"path/filepath"

	"main/internal/state"
)

// check if a track is used in any room (other than the given room)
func isTrackUsed(trackID string, skipChatID int64) bool {
	for _, room := range rooms {
		if room == nil || room.Track == nil || room.ChatID == skipChatID {
			continue
		}

		// check current track
		if room.Track.ID == trackID {
			return true
		}

		// check first 2 queued tracks
		n := len(room.Queue)
		if n > 2 {
			n = 2
		}
		for _, q := range room.Queue[:n] {
			if q.ID == trackID {
				return true
			}
		}
	}
	return false
}

func (r *RoomState) releaseFile() {
	if r == nil || r.Track == nil || r.FilePath == "" {
		return
	}

	roomsMu.RLock()
	used := isTrackUsed(r.Track.ID, r.ChatID)
	roomsMu.RUnlock()

	if !used {
		if err := os.Remove(r.FilePath); err != nil && !os.IsNotExist(err) {
			logger.ErrorF("failed to remove file %s: %v", r.FilePath, err)
		} else {
			logger.DebugF("removed unused file: %s", r.FilePath)
		}
	} else {
		logger.DebugF("file still in use, skipped remove: %s", r.FilePath)
	}
}

func (r *RoomState) cleanupFile() {
	if r == nil {
		return
	}

	// Collect current + queued tracks, up to 5
	tracks := []*state.Track{}
	if r.Track != nil {
		tracks = append(tracks, r.Track)
	}
	tracks = append(tracks, r.Queue...)
	if len(tracks) > 2 {
		tracks = tracks[:2]
	}

	roomsMu.RLock()
	for _, t := range tracks {
		if t == nil || t.ID == "" || isTrackUsed(t.ID, r.ChatID) {
			logger.DebugF("track %s still in use, skip delete", t.ID)
			continue
		}

		pattern := filepath.Join("downloads", t.ID+".*")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			logger.ErrorF("glob failed for %s: %v", pattern, err)
			continue
		}

		for _, f := range matches {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				logger.ErrorF("failed to remove file %s: %v", f, err)
			} else {
				logger.DebugF("removed unused file: %s", f)
			}
		}
	}
	roomsMu.RUnlock()
}

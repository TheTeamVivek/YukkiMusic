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

	"github.com/Laky-64/gologging"

	"main/internal/core/models"
)

// check if a track is used in any room (other than the given room)
func isTrackUsed(trackID string, skipChatID int64) bool {
	for _, room := range rooms {
		if room == nil || room.track == nil || room.chatID == skipChatID {
			continue
		}

		// check current track
		if room.track.ID == trackID {
			return true
		}

		// check first 2 queued tracks
		n := len(room.queue)
		if n > 2 {
			n = 2
		}
		for _, q := range room.queue[:n] {
			if q.ID == trackID {
				return true
			}
		}
	}
	return false
}

func (r *RoomState) releaseFile() {
	if r == nil || r.track == nil || r.fpath == "" {
		return
	}

	roomsMu.RLock()
	used := isTrackUsed(r.track.ID, r.chatID)
	roomsMu.RUnlock()

	if !used {
		if err := os.Remove(r.fpath); err != nil && !os.IsNotExist(err) {
			gologging.ErrorF("failed to remove file %s: %v", r.fpath, err)
		} else {
			gologging.DebugF("removed unused file: %s", r.fpath)
		}
	} else {
		gologging.DebugF("file still in use, skipped remove: %s", r.fpath)
	}
}

func (r *RoomState) cleanupFile() {
	if r == nil {
		return
	}

	// Collect current + queued tracks, up to 5
	tracks := []*state.Track{}
	if r.track != nil {
		tracks = append(tracks, r.track)
	}
	tracks = append(tracks, r.queue...)
	if len(tracks) > 2 {
		tracks = tracks[:2]
	}

	roomsMu.RLock()
	for _, t := range tracks {
		if t == nil || t.ID == "" || isTrackUsed(t.ID, r.chatID) {
			gologging.DebugF("track %s still in use, skip delete", t.ID)
			continue
		}

		pattern := filepath.Join("downloads", t.ID+".*")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			gologging.ErrorF("glob failed for %s: %v", pattern, err)
			continue
		}

		for _, f := range matches {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				gologging.ErrorF("failed to remove file %s: %v", f, err)
			} else {
				gologging.DebugF("removed unused file: %s", f)
			}
		}
	}
	roomsMu.RUnlock()
}

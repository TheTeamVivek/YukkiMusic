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
package core

import (
	"os"
	"path/filepath"

	"github.com/Laky-64/gologging"

	state "main/internal/core/models"
)

// check if a track is used in any room (other than the given room)
func isTrackUsed(trackID string, skipChatID int64) bool {
	for _, room := range rooms {
		if room == nil || room.track == nil || room.chatID == skipChatID {
			continue
		}

		if room.track.ID == trackID {
			return true
		}

		if isTrackInQueue(trackID, room.queue) {
			return true
		}
	}
	return false
}

// checks first N (2) queued tracks
func isTrackInQueue(trackID string, queue []*state.Track) bool {
	limit := 2
	if len(queue) < limit {
		limit = len(queue)
	}

	for _, q := range queue[:limit] {
		if q != nil && q.ID == trackID {
			return true
		}
	}
	return false
}

// release current track file if unused elsewhere
func (r *RoomState) releaseFile() {
	if r == nil || r.track == nil {
		return
	}

	track := r.track

	roomsMu.RLock()
	used := isTrackUsed(track.ID, r.chatID)
	roomsMu.RUnlock()

	if used {
		gologging.DebugF(
			"file still in use, skipped remove: %s:%s",
			string(track.Source),
			track.ID,
		)
		return
	}

	findAndRemove(track)
}

// cleanup current + queued track files if unused
func (r *RoomState) cleanupFile() {
	if r == nil {
		return
	}

	// collect current + next tracks (max 2)
	tracks := []*state.Track{}
	if r.track != nil {
		tracks = append(tracks, r.track)
	}
	tracks = append(tracks, r.queue...)
	if len(tracks) > 2 {
		tracks = tracks[:2]
	}

	for _, t := range tracks {
		if t == nil || t.ID == "" {
			continue
		}

		roomsMu.RLock()
		used := isTrackUsed(t.ID, r.chatID)
		roomsMu.RUnlock()

		if used {
			gologging.DebugF(
				"track still in use, skip delete: %s:%s",
				string(t.Source),
				t.ID,
			)
			continue
		}

		findAndRemove(t)
	}
}

func findAndRemove(track *state.Track) {
	t := "audio"
	if track.Video {
		t = "video"
	}

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		return
	}

	for _, f := range files {
		os.Remove(f)
		gologging.DebugF(
			"removed unused file: %s",
			f,
		)
	}
}

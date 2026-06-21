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

package core

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Laky-64/gologging"

	state "yukkimusic/internal/core/models"
)

var FileCacheDuration = 1 * time.Minute

func isTrackUsed(trackID string, skipChatID int64) bool {
	for _, room := range rooms {
		if room == nil || room.track == nil || room.ChatID == skipChatID {
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

func isTrackInQueue(trackID string, queue []*state.Track) bool {
	limit := min(len(queue), 2)
	for _, q := range queue[:limit] {
		if q != nil && q.ID == trackID {
			return true
		}
	}
	return false
}

func (r *RoomState) releaseFile() {
	if r == nil || r.track == nil {
		return
	}
	scheduleRemove(r.track, r.ID)
}

func (r *RoomState) cleanupFile() {
	if r == nil {
		return
	}

	tracks := make([]*state.Track, 0, 3)
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
		scheduleRemove(t, r.ID)
	}
}

// scheduleRemove deletes the track file after FileCacheDuration,
// but only if no other room is using it at deletion time.
func scheduleRemove(track *state.Track, skipChatID int64) {
	if track == nil {
		return
	}

	if FileCacheDuration <= 0 {
		doRemove(track, skipChatID)
		return
	}

	t := *track
	time.AfterFunc(FileCacheDuration, func() {
		doRemove(&t, skipChatID)
	})

	gologging.DebugF(
		"scheduled file removal in %s: %s:%s",
		FileCacheDuration, string(track.Source), track.ID,
	)
}

func doRemove(track *state.Track, skipChatID int64) {
	roomsMu.RLock()
	used := isTrackUsed(track.ID, skipChatID)
	roomsMu.RUnlock()

	if used {
		gologging.DebugF(
			"file still in use, skipped remove: %s:%s",
			string(track.Source), track.ID,
		)
		return
	}

	findAndRemove(track)
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
		gologging.DebugF("removed file: %s", f)
	}
}

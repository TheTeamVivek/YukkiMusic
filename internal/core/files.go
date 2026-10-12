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

	"yukkimusic/internal/utils"

	state "yukkimusic/internal/core/models"
)

const (
	downloadsDir = "downloads"

	audioPrefix = "audio"
	videoPrefix = "video"

	// queueScanLimit bounds how many upcoming tracks are checked when deciding
	// whether a downloaded file is still in use.
	queueScanLimit = 2
)

// isTrackUsed reports whether any room other than skipChatID is currently
// playing or has queued the track.
func isTrackUsed(trackID string, skipChatID int64) bool {
	for _, room := range Rooms.rooms {
		if room == nil || room.player.track == nil || room.ChatID == skipChatID {
			continue
		}
		if room.player.track.ID == trackID || isTrackInQueue(trackID, room.queue.tracks) {
			return true
		}
	}
	return false
}

// isTrackInQueue reports whether the next queueScanLimit queued tracks contain
// the given track.
func isTrackInQueue(trackID string, tracks []*state.Track) bool {
	limit := min(len(tracks), queueScanLimit)
	for _, q := range tracks[:limit] {
		if q != nil && q.ID == trackID {
			return true
		}
	}
	return false
}

// releaseFile schedules the active track's file for delayed removal.
func (r *Room) releaseFile() {
	if r == nil || r.player.track == nil {
		return
	}
	scheduleRemove(r.player.track, r.ChatID)
}

// cleanupFile schedules the active track and the next queued tracks for
// delayed removal.
func (r *Room) cleanupFile() {
	if r == nil {
		return
	}

	tracks := make([]*state.Track, 0, queueScanLimit+1)
	if r.player.track != nil {
		tracks = append(tracks, r.player.track)
	}
	tracks = append(tracks, r.queue.tracks...)
	if len(tracks) > queueScanLimit {
		tracks = tracks[:queueScanLimit]
	}

	for _, t := range tracks {
		if t == nil || t.ID == "" {
			continue
		}
		scheduleRemove(t, r.ChatID)
	}
}

// scheduleRemove removes the file backing track after FileCacheDuration,
// unless another room starts using it in the meantime.
func scheduleRemove(track *state.Track, skipChatID int64) {
	if track == nil {
		return
	}

	if FileCacheDuration <= 0 {
		removeIfUnused(track, skipChatID)
		return
	}

	t := *track
	time.AfterFunc(FileCacheDuration, func() {
		removeIfUnused(&t, skipChatID)
	})

	utils.Log.Debugf(
		"scheduled file removal in %s: %s:%s",
		FileCacheDuration, string(track.Source), track.ID,
	)
}

// removeIfUnused deletes the file backing track unless another room is still
// using it.
func removeIfUnused(track *state.Track, skipChatID int64) {
	Rooms.mu.RLock()
	used := isTrackUsed(track.ID, skipChatID)
	Rooms.mu.RUnlock()

	if used {
		utils.Log.Debugf(
			"file still in use, skipped remove: %s:%s",
			string(track.Source), track.ID,
		)
		return
	}

	deleteTrackFiles(track)
}

// deleteTrackFiles removes every downloaded file belonging to track.
func deleteTrackFiles(track *state.Track) {
	kind := audioPrefix
	if track.Video {
		kind = videoPrefix
	}

	pattern := filepath.Join(downloadsDir, kind+"_"+track.ID+"*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		utils.Log.Debugf("failed to glob %s: %v", pattern, err)
		return
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			utils.Log.Debugf("failed to remove file %s: %v", f, err)
			continue
		}
		utils.Log.Debugf("removed file: %s", f)
	}
}

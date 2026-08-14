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
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	state "yukkimusic/internal/core/models"
)

var (
	rooms   = make(map[int64]*RoomState)
	roomsMu sync.RWMutex

	ErrRoomDestroyed = errors.New("room destroyed")

	FileCacheDuration = 1 * time.Minute
)

// RoomState holds all playback state for a single chat.
type RoomState struct {
	mu sync.RWMutex

	// ID is the canonical playback target id (group or linked channel).
	ID int64
	// ChatID is the UI/context chat id where messages and controls are sent.
	ChatID int64

	filePath  string       // currently playing local media file
	track     *state.Track // active track metadata
	playing   bool         // whether playback is active
	paused    bool         // whether playback is paused
	muted     bool         // whether playback is muted
	speed     float64      // playback speed multiplier
	position  int          // current playback position in seconds
	updatedAt int64        // last state-update timestamp (unix seconds)
	loop      int          // loop mode/state value

	queue   []*state.Track // upcoming tracks
	shuffle bool           // queue shuffle mode

	statusMsg *td.Message // latest status message in chat

	Assistant *Assistant  // assistant client bound to this room
	destroyed atomic.Bool // whether room cleanup has completed
}

// Room management

func DropRoom(chatID int64) bool {
	_, file, line, _ := runtime.Caller(1)
	logger.Debugf("DeleteRoom called from %s:%d", file, line)

	roomsMu.Lock()
	room, ok := rooms[chatID]
	if !ok || room == nil || room.destroyed.Load() {
		roomsMu.Unlock()
		return false
	}

	delete(rooms, chatID)
	roomsMu.Unlock()

	room.cleanupFile()
	room.Stop()
	room.destroyed.Store(true)
	return true
}

// RoomFor returns the existing room for a chat.
func RoomFor(chatID int64) (*RoomState, bool) {
	roomsMu.RLock()
	room, exists := rooms[chatID]
	roomsMu.RUnlock()
	return room, exists
}

// CreateRoom returns the room for a chat, creating it if it does not exist.
func CreateRoom(chatID int64, ass *Assistant) *RoomState {
	room, _ := createNewRoom(chatID, ass)
	return room
}

func createNewRoom(chatID int64, ass *Assistant) (*RoomState, bool) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[chatID]
	if !exists {
		room = &RoomState{
			ID:        chatID,
			ChatID:    chatID,
			queue:     []*state.Track{},
			speed:     1.0,
			Assistant: ass,
		}
		rooms[chatID] = room
	}

	return room, true
}

// SetAssistant rebinds the room to another assistant (used after the
// assistant for the chat was switched).
func (r *RoomState) SetAssistant(ass *Assistant) {
	r.mu.Lock()
	r.Assistant = ass
	r.mu.Unlock()
}

func AllRooms() map[int64]*RoomState {
	roomsMu.RLock()

	out := make(map[int64]*RoomState, len(rooms))
	var dead []int64

	for chatID, room := range rooms {
		if room == nil || room.destroyed.Load() {
			dead = append(dead, chatID)
			continue
		}
		out[chatID] = room
	}

	roomsMu.RUnlock()

	if len(dead) > 0 {
		roomsMu.Lock()
		for _, chatID := range dead {
			if room := rooms[chatID]; room == nil || room.destroyed.Load() {
				delete(rooms, chatID)
			}
		}
		roomsMu.Unlock()
	}

	return out
}

// Helpers

func (r *RoomState) Destroyed() bool {
	return r.destroyed.Load()
}

func (r *RoomState) updatePosition() {
	if r == nil || r.track == nil || r.updatedAt == 0 {
		return
	}

	current := time.Now().Unix()
	elapsed := float64(current - r.updatedAt)

	if r.playing && !r.paused {
		r.position += int(elapsed * r.speed)
		if r.position >= r.track.Duration {
			r.position = r.track.Duration
			r.playing = false
		}
	}
	r.updatedAt = current
}

// Getters

func (r *RoomState) FilePath() string {
	if r.Destroyed() {
		return ""
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.filePath
}

func (r *RoomState) Loop() int {
	if r.Destroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loop
}

func (r *RoomState) Position() int {
	if r.Destroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.position
}

func (r *RoomState) Queue() []*state.Track {
	if r.Destroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := make([]*state.Track, len(r.queue))
	copy(q, r.queue)
	return q
}

func (r *RoomState) Shuffle() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.shuffle
}

func (r *RoomState) Speed() float64 {
	if r.Destroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speed
}

func (r *RoomState) Track() *state.Track {
	if r.Destroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track
}

func (r *RoomState) StatusMsg() *td.Message {
	if r.Destroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.statusMsg
}

// Setters

func (r *RoomState) SetLoop(loop int) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loop = loop
}

func (r *RoomState) SetShuffle(enabled bool) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shuffle = enabled
}

func (r *RoomState) SetStatusMsg(m *td.Message) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusMsg = m
}

// State checks

func (r *RoomState) Active() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePosition()
	return r.track != nil && r.playing
}

func (r *RoomState) Paused() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paused && r.track != nil && r.playing
}

func (r *RoomState) Muted() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.muted && r.track != nil && r.playing
}

func (r *RoomState) Parse() {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePosition()
}

// Queue

// NextTrack retrieves and prepares the next track in queue
func (r *RoomState) NextTrack() *state.Track {
	if r.Destroyed() {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.track != nil && r.loop > 0 {
		r.position = 0
		r.playing = false
		r.paused = false
		r.muted = false
		r.loop--
		r.updatedAt = time.Now().Unix()
		return r.track
	}

	r.releaseFile()

	if len(r.queue) == 0 {
		return nil
	}

	index := 0
	if r.shuffle {
		index = rand.Intn(len(r.queue))
	}

	next := r.queue[index]
	r.queue = append(r.queue[:index], r.queue[index+1:]...)

	r.track = next
	r.position = 0
	r.playing = false
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	return next
}

// RemoveTrack removes track(s) from queue
func (r *RoomState) RemoveTrack(index int) {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if index == -1 {
		r.queue = []*state.Track{}
		return
	}

	if index >= 0 && index < len(r.queue) {
		r.queue = append(r.queue[:index], r.queue[index+1:]...)
	}
}

// MoveTrack moves a track from one position to another
func (r *RoomState) MoveTrack(from, to int) {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if from < 0 || from >= len(r.queue) ||
		to < 0 || to >= len(r.queue) ||
		from == to {
		return
	}

	item := r.queue[from]
	r.queue = append(r.queue[:from], r.queue[from+1:]...)

	if to >= len(r.queue) {
		r.queue = append(r.queue, item)
	} else {
		r.queue = append(r.queue[:to], append([]*state.Track{item}, r.queue[to:]...)...)
	}
}

// AddTracks appends multiple tracks to the queue
func (r *RoomState) AddTracks(tracks []*state.Track) {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue = append(r.queue, tracks...)
}

// Files

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

	logger.Debugf(
		"scheduled file removal in %s: %s:%s",
		FileCacheDuration, string(track.Source), track.ID,
	)
}

func doRemove(track *state.Track, skipChatID int64) {
	roomsMu.RLock()
	used := isTrackUsed(track.ID, skipChatID)
	roomsMu.RUnlock()

	if used {
		logger.Debugf(
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
		logger.Debugf("removed file: %s", f)
	}
}

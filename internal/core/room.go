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
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"yukkimusic/internal/utils"

	td "github.com/AshokShau/gotdbot"

	state "yukkimusic/internal/core/models"
)

var (
	ErrRoomDestroyed = errors.New("room destroyed")

	FileCacheDuration = 1 * time.Minute
)

// Registry holds every live room, keyed by the room's chat id.
type Registry struct {
	mu    sync.RWMutex
	rooms map[int64]*Room
}

// Rooms is the process-wide room registry.
var Rooms = NewRegistry()

// NewRegistry returns an empty room registry.
func NewRegistry() *Registry {
	return &Registry{rooms: make(map[int64]*Room)}
}

// Get returns the existing room for a chat.
func (reg *Registry) Get(chatID int64) (*Room, bool) {
	reg.mu.RLock()
	room, exists := reg.rooms[chatID]
	reg.mu.RUnlock()
	return room, exists
}

// GetOrCreate returns the room for a chat, creating it if it does not exist.
func (reg *Registry) GetOrCreate(chatID int64, ass *Assistant) *Room {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	room, exists := reg.rooms[chatID]
	if !exists {
		room = &Room{
			ChatID:        chatID,
			ControlChatID: chatID,
			player:        player{speed: 1.0},
			Assistant:     ass,
		}
		reg.rooms[chatID] = room
	}

	return room
}

// Delete removes the room for a chat and tears down its playback.
func (reg *Registry) Delete(chatID int64) bool {
	_, file, line, _ := runtime.Caller(1)
	utils.Log.Debugf("Registry.Delete called from %s:%d", file, line)

	reg.mu.Lock()
	room, ok := reg.rooms[chatID]
	if !ok || room == nil || room.destroyed.Load() {
		reg.mu.Unlock()
		return false
	}
	delete(reg.rooms, chatID)
	reg.mu.Unlock()

	room.cleanupFile()
	room.Stop()
	room.destroyed.Store(true)
	return true
}

// All returns a snapshot of every live room, pruning destroyed entries.
func (reg *Registry) All() map[int64]*Room {
	reg.mu.RLock()
	out := make(map[int64]*Room, len(reg.rooms))
	var dead []int64

	for chatID, room := range reg.rooms {
		if room == nil || room.destroyed.Load() {
			dead = append(dead, chatID)
			continue
		}
		out[chatID] = room
	}
	reg.mu.RUnlock()

	if len(dead) > 0 {
		reg.mu.Lock()
		for _, chatID := range dead {
			if room := reg.rooms[chatID]; room == nil || room.destroyed.Load() {
				delete(reg.rooms, chatID)
			}
		}
		reg.mu.Unlock()
	}

	return out
}

// Room holds all playback state for a single chat.
type Room struct {
	mu sync.RWMutex

	// ChatID is the room identity and the canonical playback target
	// (map key; the group's or linked channel's voice chat).
	ChatID int64
	// ControlChatID is the UI/context chat id where messages and controls
	// are sent. It equals ChatID except in channel-play mode, where it is
	// the origin group that invoked the command.
	ControlChatID int64

	player player // current track and its playback state
	queue  queue  // upcoming tracks

	autoplay bool // auto-play recommendations when the queue ends

	statusMsg *td.Message // latest status message in chat

	Assistant *Assistant  // assistant client bound to this room
	destroyed atomic.Bool // whether room cleanup has completed
}

// player holds the state of the currently playing track.
type player struct {
	track     *state.Track // active track metadata
	filePath  string       // currently playing local media file
	playing   bool         // whether playback is active
	paused    bool         // whether playback is paused
	muted     bool         // whether playback is muted
	speed     float64      // playback speed multiplier
	position  int          // current playback position in seconds
	updatedAt int64        // last state-update timestamp (unix seconds)
	loop      int          // loop mode/state value
}

// SetAssistant rebinds the room to another assistant (used after the
// assistant for the chat was switched).
func (r *Room) SetAssistant(ass *Assistant) {
	r.mu.Lock()
	r.Assistant = ass
	r.mu.Unlock()
}

// Helpers

func (r *Room) Destroyed() bool {
	return r.destroyed.Load()
}

// sync advances the playback position to the current wall-clock time.
func (p *player) sync() {
	if p == nil || p.track == nil || p.updatedAt == 0 {
		return
	}

	now := time.Now().Unix()
	elapsed := float64(now - p.updatedAt)

	if p.playing && !p.paused {
		p.position += int(elapsed * p.speed)
		if p.position >= p.track.Duration {
			p.position = p.track.Duration
			p.playing = false
		}
	}
	p.updatedAt = now
}

// reset clears the per-track transient playback state.
func (p *player) reset() {
	p.position = 0
	p.playing = false
	p.paused = false
	p.muted = false
	p.updatedAt = time.Now().Unix()
}

// active reports whether a track is currently playing.
func (p *player) active() bool {
	return p.track != nil && p.playing
}

// isPaused reports whether the active track is paused.
func (p *player) isPaused() bool {
	return p.paused && p.track != nil && p.playing
}

// isMuted reports whether the active track is muted.
func (p *player) isMuted() bool {
	return p.muted && p.track != nil && p.playing
}

// Getters

func (r *Room) FilePath() string {
	if r.Destroyed() {
		return ""
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.filePath
}

func (r *Room) Loop() int {
	if r.Destroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.loop
}

func (r *Room) Position() int {
	if r.Destroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.position
}

func (r *Room) Shuffle() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queue.shuffle
}

func (r *Room) Autoplay() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.autoplay
}

func (r *Room) Speed() float64 {
	if r.Destroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.speed
}

func (r *Room) Track() *state.Track {
	if r.Destroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.track
}

func (r *Room) StatusMsg() *td.Message {
	if r.Destroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.statusMsg
}

// Setters

func (r *Room) SetLoop(loop int) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.player.loop = loop
}

func (r *Room) SetShuffle(enabled bool) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue.shuffle = enabled
}

func (r *Room) SetAutoplay(enabled bool) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoplay = enabled
}

func (r *Room) SetStatusMsg(m *td.Message) {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusMsg = m
}

// State checks

func (r *Room) Active() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.player.sync()
	return r.player.active()
}

func (r *Room) Paused() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.isPaused()
}

func (r *Room) Muted() bool {
	if r.Destroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.player.isMuted()
}

// Refresh advances the playback position to the current time.
func (r *Room) Refresh() {
	if r.Destroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.player.sync()
}

// State is a consistent, point-in-time view of a room's playback state.
type State struct {
	ChatID   int64
	Track    *state.Track
	FilePath string
	Position int
	Speed    float64
	Playing  bool
	Paused   bool
	Muted    bool
	Loop     int
	Shuffle  bool
	Autoplay bool
	QueueLen int
}

// Active reports whether the room is currently playing a track.
func (s State) Active() bool {
	return s.Track != nil && s.Playing
}

// Snapshot returns a consistent view of the room's state, first catching the
// playback position up to the current time.
func (r *Room) Snapshot() State {
	if r.Destroyed() {
		return State{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.player.sync()

	return State{
		ChatID:   r.ChatID,
		Track:    r.player.track,
		FilePath: r.player.filePath,
		Position: r.player.position,
		Speed:    r.player.speed,
		Playing:  r.player.playing,
		Paused:   r.player.paused,
		Muted:    r.player.muted,
		Loop:     r.player.loop,
		Shuffle:  r.queue.shuffle,
		Autoplay: r.autoplay,
		QueueLen: r.queue.len(),
	}
}

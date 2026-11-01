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
	"maps"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"

	state "yukkimusic/internal/core/models"
	"yukkimusic/ntgcalls"
)

var (
	// ErrRoomClosed is returned when a room that was already removed from
	// the registry tries to (re)start a stream.
	ErrRoomClosed = errors.New("room closed")

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
	defer reg.mu.RUnlock()
	room, ok := reg.rooms[chatID]
	return room, ok
}

// GetOrCreate returns the room for a chat, creating it if it does not exist.
func (reg *Registry) GetOrCreate(chatID int64, ass *Assistant) *Room {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	room, ok := reg.rooms[chatID]
	if !ok {
		room = &Room{
			ChatID:        chatID,
			ControlChatID: chatID,
			Assistant:     ass,
			speed:         1.0,
		}
		reg.rooms[chatID] = room
	}
	return room
}

// Delete removes the room for a chat and stops its stream.
func (reg *Registry) Delete(chatID int64) bool {
	reg.mu.Lock()
	room, ok := reg.rooms[chatID]
	delete(reg.rooms, chatID)
	reg.mu.Unlock()

	if !ok {
		return false
	}

	room.cleanupFile()
	room.Stop()
	return true
}

// All returns a copy of the live rooms.
func (reg *Registry) All() map[int64]*Room {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	return maps.Clone(reg.rooms)
}

// Room keeps what ntgcalls cannot know: the track, its queue and the
// options. Whether the stream is playing, paused or muted, and how far it
// has played, is always read back from ntgcalls.
type Room struct {
	mu sync.RWMutex

	// ChatID is the room identity and the canonical playback target
	// (map key; the group's or linked channel's voice chat).
	ChatID int64
	// ControlChatID is the UI/context chat id where messages and controls
	// are sent. It equals ChatID except in channel-play mode, where it is
	// the origin group that invoked the command.
	ControlChatID int64

	Assistant *Assistant

	track    *state.Track
	filePath string
	speed    float64
	loop     int
	queue    queue

	// offset is the position at anchor; anchor is when the stream last
	// started advancing, and is zero while it is paused or idle.
	offset int
	anchor time.Time

	autoplay  bool
	statusMsg *td.Message
}

// Playback is what the stream of a room is doing.
type Playback uint8

const (
	PlaybackIdle Playback = iota
	PlaybackPlaying
	PlaybackPaused
	PlaybackMuted
)

// State is a point-in-time view of a room.
type State struct {
	Track    *state.Track
	Playback Playback
	Position int
	QueueLen int
}

func (s State) Active() bool {
	return s.Playback != PlaybackIdle
}

// Snapshot returns the current state of the room.
func (r *Room) Snapshot() State {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s := State{Track: r.track, QueueLen: r.queue.len()}
	if r.track != nil {
		s.Playback = r.playback()
		s.Position = r.position()
	}
	return s
}

// playback reads the stream status from ntgcalls.
func (r *Room) playback() Playback {
	calls, err := r.Assistant.Ntg.Calls()
	if err != nil {
		return PlaybackIdle
	}

	call, ok := calls[r.ChatID]
	if !ok {
		return PlaybackIdle
	}

	switch call.Capture {
	case ntgcalls.IdlingStream:
		return PlaybackIdle
	case ntgcalls.PausedStream:
		return PlaybackPaused
	}

	if st, err := r.Assistant.Ntg.State(r.ChatID); err == nil && st.Muted {
		return PlaybackMuted
	}
	return PlaybackPlaying
}

// position is the playback position in seconds. Callers must hold r.mu.
func (r *Room) position() int {
	if r.track == nil {
		return 0
	}

	pos := r.offset
	if !r.anchor.IsZero() {
		pos += int(time.Since(r.anchor).Seconds() * r.speed)
	}
	return min(pos, r.track.Duration)
}

// SetAssistant rebinds the room to another assistant (used after the
// assistant for the chat was switched).
func (r *Room) SetAssistant(ass *Assistant) {
	r.mu.Lock()
	r.Assistant = ass
	r.mu.Unlock()
}

func (r *Room) Active() bool {
	return r.Snapshot().Active()
}

func (r *Room) Paused() bool {
	return r.Snapshot().Playback == PlaybackPaused
}

func (r *Room) Muted() bool {
	return r.Snapshot().Playback == PlaybackMuted
}

func (r *Room) Position() int {
	return r.Snapshot().Position
}

func (r *Room) Track() *state.Track {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track
}

func (r *Room) FilePath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.filePath
}

func (r *Room) Speed() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speed
}

func (r *Room) Loop() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loop
}

func (r *Room) SetLoop(loop int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loop = loop
}

func (r *Room) Shuffle() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queue.shuffle
}

func (r *Room) SetShuffle(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue.shuffle = enabled
}

func (r *Room) Autoplay() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.autoplay
}

func (r *Room) SetAutoplay(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoplay = enabled
}

func (r *Room) StatusMsg() *td.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.statusMsg
}

func (r *Room) SetStatusMsg(m *td.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusMsg = m
}

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
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
)

var (
	rooms   = make(map[int64]*RoomState)
	roomsMu sync.RWMutex

	ErrRoomDestroyed = errors.New("room destroyed")
)

type RoomState struct {
	mu sync.RWMutex

	// Identity
	chatID        int64
	channelPlayID int64

	// Playback State
	filePath  string
	track     *state.Track
	playing   bool
	paused    bool
	muted     bool
	speed     float64
	position  int
	updatedAt int64
	loop      int

	// Queue State
	queue   []*state.Track
	shuffle bool

	// Automation
	*scheduledTimers

	// Metadata & UI
	statusMsg *telegram.NewMessage
	Data      map[string]any

	// Internal Components
	Assistant *Assistant
	destroyed atomic.Bool
}

type scheduledTimers struct {
	scheduledUnmuteTimer *time.Timer
	scheduledResumeTimer *time.Timer
	scheduledSpeedTimer  *time.Timer

	scheduledUnmuteUntil time.Time
	scheduledResumeUntil time.Time
	scheduledSpeedUntil  time.Time
}

// Room management functions

func DeleteRoom(chatID int64) bool {
	_, file, line, _ := runtime.Caller(1)
	gologging.DebugF("DeleteRoom called from %s:%d", file, line)

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

// GetRoom retrieves an existing room or creates a new one if requested.
func GetRoom(chatID int64, ass *Assistant, create bool) (*RoomState, bool) {
	roomsMu.RLock()
	room, exists := rooms[chatID]
	roomsMu.RUnlock()

	if exists {
		return room, true
	}

	if create {
		return createNewRoom(chatID, ass)
	}

	return nil, false
}

func createNewRoom(chatID int64, ass *Assistant) (*RoomState, bool) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[chatID]
	if !exists {
		room = &RoomState{
			chatID:    chatID,
			queue:     []*state.Track{},
			speed:     1.0,
			Assistant: ass,
			Data:      make(map[string]any),
		}
		room.destroyed.Store(false)
		rooms[chatID] = room
	}

	return room, true
}

func GetAllRooms() map[int64]*RoomState {
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

func (r *RoomState) IsDestroyed() bool {
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

func (st *scheduledTimers) RemainingUnmuteDuration() time.Duration {
	if st == nil || st.scheduledUnmuteUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledUnmuteUntil)
}

func (st *scheduledTimers) RemainingResumeDuration() time.Duration {
	if st == nil || st.scheduledResumeUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledResumeUntil)
}

func (st *scheduledTimers) RemainingSpeedDuration() time.Duration {
	if st == nil || st.scheduledSpeedUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledSpeedUntil)
}

func (st *scheduledTimers) cancelScheduledUnmute() {
	if st != nil && st.scheduledUnmuteTimer != nil {
		st.scheduledUnmuteTimer.Stop()
		st.scheduledUnmuteTimer = nil
		st.scheduledUnmuteUntil = time.Time{}
	}
}

func (st *scheduledTimers) cancelScheduledResume() {
	if st != nil && st.scheduledResumeTimer != nil {
		st.scheduledResumeTimer.Stop()
		st.scheduledResumeTimer = nil
		st.scheduledResumeUntil = time.Time{}
	}
}

func (st *scheduledTimers) cancelScheduledSpeed() {
	if st != nil && st.scheduledSpeedTimer != nil {
		st.scheduledSpeedTimer.Stop()
		st.scheduledSpeedTimer = nil
		st.scheduledSpeedUntil = time.Time{}
	}
}

// Getters

func (r *RoomState) EffectiveChatID() int64 {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.channelPlayID != 0 {
		return r.channelPlayID
	}
	return r.chatID
}

func (r *RoomState) ChannelPlayID() int64 {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channelPlayID
}

func (r *RoomState) ChatID() int64 {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.chatID
}

func (r *RoomState) FilePath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.filePath
}

func (r *RoomState) Loop() int {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loop
}

func (r *RoomState) Position() int {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.position
}

func (r *RoomState) Queue() []*state.Track {
	if r.IsDestroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := make([]*state.Track, len(r.queue))
	copy(q, r.queue)
	return q
}

func (r *RoomState) Shuffle() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.shuffle
}

func (r *RoomState) Speed() float64 {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speed
}

func (r *RoomState) Track() *state.Track {
	if r.IsDestroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track
}

func (r *RoomState) StatusMsg() *telegram.NewMessage {
	if r.IsDestroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.statusMsg
}

func (r *RoomState) GetData(k string) (bool, any) {
	if r.IsDestroyed() {
		return false, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.Data[k]
	return ok, v
}

// Setters

func (r *RoomState) SetLoop(loop int) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loop = loop
}

func (r *RoomState) SetChannelPlayID(chatID int64) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channelPlayID = chatID
}

func (r *RoomState) SetData(k string, v any) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Data == nil {
		r.Data = make(map[string]any)
	}
	r.Data[k] = v
}

func (r *RoomState) DeleteData(k string) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Data, k)
}

func (r *RoomState) SetShuffle(enabled bool) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shuffle = enabled
}

func (r *RoomState) SetStatusMsg(m *telegram.NewMessage) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusMsg = m
}

// State checks

func (r *RoomState) IsActiveChat() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePosition()
	return r.track != nil && r.playing
}

func (r *RoomState) IsPaused() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paused && r.track != nil && r.playing
}

func (r *RoomState) IsMuted() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.muted && r.track != nil && r.playing
}

func (r *RoomState) Parse() {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePosition()
}

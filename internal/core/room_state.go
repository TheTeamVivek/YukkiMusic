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

type Player interface {
	Play(r *RoomState) error
	Pause(r *RoomState) (bool, error)
	Resume(r *RoomState) (bool, error)
	Stop(r *RoomState) error
	Mute(r *RoomState) (bool, error)
	Unmute(r *RoomState) (bool, error)
}

type RoomState struct {
	mu sync.RWMutex

	fpath string         // current track file path
	track *state.Track   // current track metadata
	queue []*state.Track // playback queue

	loop     int // number of times to replay current track
	position int // current playback position

	updatedAt int64 // last update timestamp
	chatID    int64 // chat where audio is streamed
	cplayID   int64 // chat for service messages ( when channelplay then it will be provided to send service msg in that chat)

	speed float64 // playback speed (0.5–4.0)

	shuffle bool // pick random track from queue
	playing bool // currently playing
	muted   bool // audio muted
	paused  bool // playback paused

	autoplay   bool   // autoplay recommendations
	autoplayHL string // autoplay language
	autoplayGL string // autoplay country

	destroyed atomic.Bool // room destroyed flag

	mystic *telegram.NewMessage // active telegram message

	p                Player // player
	*scheduledTimers        // playback timers
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

func GetRoom(chatID int64, ass *Assistant, create ...bool) (*RoomState, bool) {
	roomsMu.RLock()
	room, exists := rooms[chatID]
	roomsMu.RUnlock()

	if exists {
		return room, true
	}

	if len(create) > 0 && create[0] {
		return createNewRoom(chatID, ass)
	}

	return nil, false
}

// TODO: Take hl, gl as input
func createNewRoom(chatID int64, ass *Assistant) (*RoomState, bool) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[chatID]
	if !exists {
		room = &RoomState{
			chatID:     chatID,
			queue:      []*state.Track{},
			speed:      1.0,
			autoplayHL: "en",
			autoplayGL: "IN",
			p: &NtgPlayer{
				Ntg: ass.Ntg,
			},
		}
		room.destroyed.Store(false)
		rooms[chatID] = room
	}

	return room, true
}

func GetRoomCounts() int {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	return len(rooms)
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

// Getters

// EffectiveChatID returns the chat ID that should be used for sending messages.
func (r *RoomState) EffectiveChatID() int64 {
	if r.destroyed.Load() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.cplayID != 0 {
		return r.cplayID
	}
	return r.chatID
}

func (r *RoomState) CplayID() int64 {
	if r.destroyed.Load() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cplayID
}

func (r *RoomState) ChatID() int64 {
	if r.destroyed.Load() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.chatID
}

func (r *RoomState) FilePath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.fpath
}

func (r *RoomState) Loop() int {
	if r.destroyed.Load() {
		return 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loop
}

func (r *RoomState) Position() int {
	if r.destroyed.Load() {
		return 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.position
}

func (r *RoomState) Queue() []*state.Track {
	if r.destroyed.Load() {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	q := make([]*state.Track, len(r.queue))
	copy(q, r.queue)
	return q
}

func (r *RoomState) Shuffle() bool {
	if r.destroyed.Load() {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.shuffle
}

func (r *RoomState) Speed() float64 {
	if r.destroyed.Load() {
		return 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speed
}

func (r *RoomState) Track() *state.Track {
	if r.destroyed.Load() {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track
}

func (r *RoomState) GetSpeed() float64 {
	if r.destroyed.Load() {
		return 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speed
}

func (r *RoomState) GetMystic() *telegram.NewMessage {
	if r.destroyed.Load() {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mystic
}

func (r *RoomState) Autoplay() bool {
	if r.destroyed.Load() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.autoplay
}

func (r *RoomState) AutoplayHL() string {
	if r.destroyed.Load() {
		return ""
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.autoplayHL
}

func (r *RoomState) AutoplayGL() string {
	if r.destroyed.Load() {
		return ""
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.autoplayGL
}

func (r *RoomState) Destroyed() bool {
	return r.destroyed.Load()
}

// Setters

func (r *RoomState) SetLoop(loop int) {
	if r.destroyed.Load() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.loop = loop
}

func (r *RoomState) SetCplayID(chatID int64) {
	if r.destroyed.Load() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.cplayID = chatID
}

func (r *RoomState) SetShuffle(enabled bool) {
	if r.destroyed.Load() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.shuffle = enabled
}

func (r *RoomState) SetMystic(m *telegram.NewMessage) {
	if r.destroyed.Load() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mystic != nil {
		r.mystic.Delete()
	}
	r.mystic = m
}

func (r *RoomState) SetAutoplay(enabled bool) {
	if r.destroyed.Load() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoplay = enabled
}

func (r *RoomState) SetAutoplayHL(hl string) {
	if r.destroyed.Load() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoplayHL = hl
}

func (r *RoomState) SetAutoplayGL(gl string) {
	if r.destroyed.Load() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoplayGL = gl
}

func (r *RoomState) PrepareForAutoPlay() {
	if r.destroyed.Load() {
		return
	}
	r.releaseFile()
}

// State checks

func (r *RoomState) IsActiveChat() bool {
	if r.destroyed.Load() {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.parse()
	return r.track != nil && r.playing
}

func (r *RoomState) IsPaused() bool {
	if r.destroyed.Load() {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paused && r.track != nil && r.playing
}

func (r *RoomState) IsMuted() bool {
	if r.destroyed.Load() {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.muted && r.track != nil && r.playing
}

// State management
func (r *RoomState) Parse() {
	if r.destroyed.Load() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.parse()
}

func (r *RoomState) parse() {
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

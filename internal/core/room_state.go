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
	sync.RWMutex

	fpath             string
	track             *state.Track
	queue             []*state.Track
	loop, position    int
	updatedAt, chatID int64
	speed             float64

	shuffle, playing,
	muted, paused,
	cplay bool

	destroyed atomic.Bool

	mystic *telegram.NewMessage

	p Player
	*scheduledTimers
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

func createNewRoom(chatID int64, ass *Assistant) (*RoomState, bool) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[chatID]
	if !exists {
		room = &RoomState{
			chatID: chatID,
			queue:  []*state.Track{},
			speed:  1.0,
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

func (r *RoomState) ChatID() int64 {
	if r.destroyed.Load() {
		return 0
	}
	r.RLock()
	defer r.RUnlock()
	return r.chatID
}

func (r *RoomState) FilePath() string {
	r.RLock()
	defer r.RUnlock()
	return r.fpath
}

func (r *RoomState) Loop() int {
	if r.destroyed.Load() {
		return 0
	}

	r.RLock()
	defer r.RUnlock()
	return r.loop
}

func (r *RoomState) Position() int {
	if r.destroyed.Load() {
		return 0
	}

	r.RLock()
	defer r.RUnlock()
	return r.position
}

func (r *RoomState) Queue() []*state.Track {
	if r.destroyed.Load() {
		return nil
	}

	r.RLock()
	defer r.RUnlock()

	q := make([]*state.Track, len(r.queue))
	copy(q, r.queue)
	return q
}

func (r *RoomState) Shuffle() bool {
	if r.destroyed.Load() {
		return false
	}

	r.RLock()
	defer r.RUnlock()
	return r.shuffle
}

func (r *RoomState) Speed() float64 {
	if r.destroyed.Load() {
		return 0
	}

	r.RLock()
	defer r.RUnlock()
	return r.speed
}

func (r *RoomState) Track() *state.Track {
	if r.destroyed.Load() {
		return nil
	}

	r.RLock()
	defer r.RUnlock()
	return r.track
}

func (r *RoomState) GetSpeed() float64 {
	if r.destroyed.Load() {
		return 0
	}

	r.RLock()
	defer r.RUnlock()
	return r.speed
}

func (r *RoomState) GetMystic() *telegram.NewMessage {
	if r.destroyed.Load() {
		return nil
	}

	r.RLock()
	defer r.RUnlock()
	return r.mystic
}

func (r *RoomState) Destroyed() bool {
	return r.destroyed.Load()
}

// Setters

func (r *RoomState) SetCPlay(isCPlay bool) {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()
	r.cplay = isCPlay
}

func (r *RoomState) SetLoop(loop int) {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()
	r.loop = loop
}

func (r *RoomState) SetShuffle(enabled bool) {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()
	r.shuffle = enabled
}

func (r *RoomState) SetMystic(m *telegram.NewMessage) {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()
	if r.mystic != nil {
		r.mystic.Delete()
	}
	r.mystic = m
}

// State checks

func (r *RoomState) IsCPlay() bool {
	if r.destroyed.Load() {
		return false
	}

	r.RLock()
	defer r.RUnlock()
	return r.cplay
}

func (r *RoomState) IsActiveChat() bool {
	if r.destroyed.Load() {
		return false
	}

	r.Lock()
	defer r.Unlock()
	r.parse()
	return r.track != nil && r.playing
}

func (r *RoomState) IsPaused() bool {
	if r.destroyed.Load() {
		return false
	}

	r.RLock()
	defer r.RUnlock()
	return r.paused && r.track != nil && r.playing
}

func (r *RoomState) IsMuted() bool {
	if r.destroyed.Load() {
		return false
	}

	r.RLock()
	defer r.RUnlock()
	return r.muted && r.track != nil && r.playing
}

// State management
func (r *RoomState) Parse() {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()
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

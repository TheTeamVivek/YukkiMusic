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
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/state"
)

var (
	rooms   = make(map[int64]*RoomState)
	roomsMu sync.RWMutex
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

	chatID    int64
	track     *state.Track
	position  int
	playing   bool
	muted     bool
	paused    bool
	updatedAt int64
	fpath  string
	queue     []*state.Track
	speed     float64
	shuffle   bool

	loop   int
	cplay  bool
	mystic *telegram.NewMessage

	rtmpURL string
	rtmpKey string

	p Player
	*scheduledTimers
}

func DeleteRoom(chatID int64) {
	_, file, line, _ := runtime.Caller(1)

	gologging.DebugF("DeleteRoom Called from %s:%d", file, line)

	roomsMu.RLock()
	room, exists := rooms[chatID]
	roomsMu.RUnlock()

	if exists {
		room.Destroy()
	}
}

func GetRoom(chatID int64, create ...bool) (*RoomState, bool) {
	roomsMu.RLock()
	room, exists := rooms[chatID]
	roomsMu.RUnlock()

	if exists {
		return room, true
	}

	if len(create) > 0 && create[0] {
		roomsMu.Lock()
		defer roomsMu.Unlock()

		room, exists = rooms[chatID]
		if !exists {
			room = &RoomState{
				chatID: chatID,
				queue:  []*state.Track{},
				speed:  1.0,
				p:      &NtgPlayer{},
			}
			rooms[chatID] = room
		}
		return room, true
	}
	return nil, false
}

func GetAllRoomIDs() []int64 {
	roomsMu.RLock()
	defer roomsMu.RUnlock()

	ids := make([]int64, 0, len(rooms))
	for chatID := range rooms {
		ids = append(ids, chatID)
	}
	return ids
}


func (r *RoomState) ChatID() int64 {
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
	r.RLock()
	defer r.RUnlock()
	return r.loop
}

func (r *RoomState) Position() int {
	r.RLock()
	defer r.RUnlock()
	return r.position
}

func (r *RoomState) Queue() []*state.Track {
	r.RLock()
	defer r.RUnlock()

	q := make([]*state.Track, len(r.queue))
	copy(q, r.queue)
	return q
}
func (r *RoomState) Shuffle() bool {
	r.RLock()
	defer r.RUnlock()
	return r.shuffle
}

func (r *RoomState) Speed() float64 {
	r.RLock()
	defer r.RUnlock()
	return r.speed
}

func (r  *RoomState) Track() state.Track {
    r.RLock()
	defer r.RUnlock()
	
    return *r.track
}

func (r *RoomState) SetCPlay(isCPlay bool) {
	r.Lock()
	defer r.Unlock()
	r.cplay = isCPlay
}

func (r *RoomState) SetLoop(loop int) {
	r.Lock()
	defer r.Unlock()
	r.loop = loop
}
func (r *RoomState) SetRTMPPlayer(url, key string) {
	r.Lock()
	defer r.Unlock()

	r.rtmpURL = url
	r.rtmpKey = key
	r.p = &RTMPPlayer{}
}

func (r *RoomState) IsCPlay() bool {
	r.RLock()
	defer r.RUnlock()
	return r.cplay
}

func (r *RoomState) IsActiveChat() bool {
	r.Parse()
	r.RLock()
	defer r.RUnlock()
	return r.track != nil && r.playing
}

func (r *RoomState) IsPaused() bool {
	r.RLock()
	defer r.RUnlock()
	return r.paused && r.track != nil && r.playing
}

func (r *RoomState) IsMuted() bool {
	r.RLock()
	defer r.RUnlock()
	return r.muted && r.track != nil && r.playing
}

func (r *RoomState) GetSpeed() float64 {
	r.RLock()
	defer r.RUnlock()
	return r.speed
}

func (r *RoomState) Parse() {
	r.Lock()
	defer r.Unlock()
	r.parse()
}

func (r *RoomState) SetMystic(m *telegram.NewMessage) {
	r.Lock()
	defer r.Unlock()
	if r.mystic != nil {
		r.mystic.Delete()
	}
	r.mystic = m
}

func (r *RoomState) GetMystic() *telegram.NewMessage {
	r.RLock()
	defer r.RUnlock()
	return r.mystic
}

func (r *RoomState) Destroy() {
	_, file, line, _ := runtime.Caller(1)

	gologging.DebugF("Destroy Called from %s:%d", file, line)

	r.Stop()
	r.cleanupFile()
	roomsMu.Lock()
	defer roomsMu.Unlock()
	delete(rooms, r.chatID)
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

func (r *RoomState) SetShuffle(enabled bool) {
	r.Lock()
	defer r.Unlock()
	r.shuffle = enabled
}

func (r *RoomState) Play(t *state.Track, path string, force ...bool) error {
	r.Lock()
	defer r.Unlock()

	forcePlay := len(force) > 0 && force[0]

	if !forcePlay && r.playing && r.track != nil {
		r.queue = append(r.queue, t)
		return nil
	}

	r.track = t
	r.position = 0
	r.playing = true
	r.paused = false
	r.muted = false
	r.fpath = path
	r.updatedAt = time.Now().Unix()

	return r.p.Play(r) // note: when err so must handle and cleanup room
}

func (r *RoomState) Pause(autoResumeAfter ...time.Duration) (bool, error) {
	if r.IsPaused() {
		return true, nil
	}

	paused, err := r.p.Pause(r)
	if err != nil {
		return false, err
	}

	if r.IsMuted() {
		r.Unmute()
	}

	r.Lock()
	defer r.Unlock()

	r.parse()
	r.paused = true
	r.muted = false

	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledResume()

	if len(autoResumeAfter) > 0 && autoResumeAfter[0] > 0 {
		d := autoResumeAfter[0]
		r.scheduledResumeUntil = time.Now().Add(d)
		r.scheduledResumeTimer = time.AfterFunc(d, func() {
			r.Resume()
		})
	}
	return paused, nil
}

func (r *RoomState) Resume() (bool, error) {
	if !r.IsActiveChat() {
		return false, fmt.Errorf("there are no active music playing")
	}
	if !r.IsPaused() {
		return true, nil
	}

	r.Lock()
	defer r.Unlock()

	resumed, err := r.p.Resume(r)
	if err != nil {
		return false, err
	}

	r.paused = false
	r.muted = false
	r.playing = true
	r.updatedAt = time.Now().Unix()

	r.scheduledTimers.cancelScheduledResume()
	return resumed, nil
}

func (r *RoomState) Replay() error {
	r.Lock()
	defer r.Unlock()

	if r.track == nil || r.fpath == "" {
		return fmt.Errorf("no track to replay")
	}

	old := r.position
	r.position = 0

	err := r.p.Play(r)
	if err != nil {
		r.position = old
		return err
	}

	r.playing = true
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	r.scheduledTimers.cancelScheduledResume()
	r.scheduledTimers.cancelScheduledUnmute()
	return nil
}

func (r *RoomState) Seek(seconds int) error {
	r.Lock()
	defer r.Unlock()

	if r.track == nil || r.fpath == "" {
		return fmt.Errorf("no track to seek")
	}

	r.parse()

	if seconds > 0 && r.track.Duration-r.position <= 10 {
		return fmt.Errorf("cannot seek, track is about to end")
	}

	oldPos := r.position
	oldPlaying := r.playing
	oldPaused := r.paused
	oldMuted := r.muted
	oldUpdated := r.updatedAt

	newPos := r.position + seconds
	if newPos >= r.track.Duration {
		newPos = r.track.Duration - 5
	}
	if newPos < 0 {
		newPos = 0
	}

	r.position = newPos
	r.playing = true
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	err := r.p.Play(r)
	if err != nil {
		r.position = oldPos
		r.playing = oldPlaying
		r.paused = oldPaused
		r.muted = oldMuted
		r.updatedAt = oldUpdated
		return err
	}
	if oldMuted {
		r.p.Unmute(r)
	}
	return nil
}

func (r *RoomState) SetSpeed(speed float64, timeAfterNormal ...time.Duration) error {
	r.Lock()
	defer r.Unlock()

	if r.track == nil || r.fpath == "" {
		return fmt.Errorf("no track to adjust speed")
	}
	if speed < 0.50 || speed > 4.0 {
		return fmt.Errorf("invalid speed: must be between 0.50x and 4.0x")
	}
	if r.speed == speed {
		return nil
	}

	r.parse()
	r.speed = speed
	r.playing = true
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	err := r.p.Play(r)
	if err != nil {
		return err
	}

	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledSpeed()

	if len(timeAfterNormal) > 0 && timeAfterNormal[0] > 0 && speed != 1.0 {
		d := timeAfterNormal[0]
		r.scheduledSpeedUntil = time.Now().Add(d)

		r.scheduledSpeedTimer = time.AfterFunc(d, func() {
			r.Lock()
			defer r.Unlock()
			if r.track != nil && r.playing && r.speed != 1.0 {
				r.parse()
				r.speed = 1.0
				r.p.Play(r)
				r.updatedAt = time.Now().Unix()
			}
		})
	}
	return nil
}

func (r *RoomState) Mute(unmuteAfter ...time.Duration) (bool, error) {
	if r.IsMuted() {
		return true, nil
	}

	muted, err := r.p.Mute(r)
	if err != nil {
		return false, err
	}

	if r.IsPaused() {
		r.Resume()
	} else {
		r.Parse()
	}

	r.Lock()
	defer r.Unlock()
	r.muted = true
	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledUnmute()

	if len(unmuteAfter) > 0 && unmuteAfter[0] > 0 {
		duration := unmuteAfter[0]
		r.scheduledUnmuteUntil = time.Now().Add(duration)

		r.scheduledUnmuteTimer = time.AfterFunc(duration, func() {
			r.Parse()
			r.Unmute()
		})
	}
	return muted, nil
}

func (r *RoomState) Unmute() (bool, error) {
	r.Lock()
	defer r.Unlock()

	unmuted, err := r.p.Unmute(r)
	if err != nil {
		return false, err
	}
	r.parse()
	r.muted = false
	r.paused = false
	r.scheduledTimers.cancelScheduledUnmute()
	return unmuted, nil
}

func (r *RoomState) Stop() error {
	r.Lock()
	defer r.Unlock()

	_, file, line, _ := runtime.Caller(1)

	gologging.DebugF("Stop Called from %s:%d", file, line)

	err := r.p.Stop(r)

	r.track = nil
	r.position = 0
	r.playing = false
	r.paused = false
	r.muted = false
	r.updatedAt = 0
	r.scheduledTimers.cancelScheduledUnmute()
	r.scheduledTimers.cancelScheduledResume()
	r.scheduledTimers.cancelScheduledSpeed()
	return err
}

func (r *RoomState) NextTrack() *state.Track {
	r.Lock()
	defer r.Unlock()

	if r.track != nil && r.loop > 0 {
		r.position = 0
		r.playing = true
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

func (r *RoomState) RemoveFromQueue(index int) {
	r.Lock()
	defer r.Unlock()

	if index == -1 {
		r.queue = []*state.Track{}
		return
	}

	if index < 0 || index >= len(r.queue) {
		return
	}

	r.queue = append(r.queue[:index], r.queue[index+1:]...)
}

func (r *RoomState) MoveInQueue(from, to int) {
	r.Lock()
	defer r.Unlock()

	if from < 0 || from >= len(r.queue) || to < 0 || to >= len(r.queue) || from == to {
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

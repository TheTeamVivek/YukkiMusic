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
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/amarnathcjd/gortc/groupcall"
	"github.com/amarnathcjd/gortc/media"

	state "yukkimusic/internal/core/models"
)

const (
	seekEndThreshold = 10
	seekSafetyMargin = 5
)

// FileCleanupDelay is how long a track's downloaded file sits around after
// it's no longer needed before it's actually deleted. Zero deletes immediately.
var FileCleanupDelay = time.Minute

var (
	rooms   = make(map[int64]*RoomState)
	roomsMu sync.RWMutex

	OnStreamEnd func(chatID int64)

	ErrRoomDestroyed     = errors.New("room destroyed")
	ErrCallNotJoined     = errors.New("voice call: not joined")
	ErrConnectionTimeout = errors.New("voice call: connection timeout")
)

// RoomState is an active voice-chat playback session for a single chat. It
// owns the queue/track metadata and the underlying gortc group call, since
// the two always live and die together.
type RoomState struct {
	mu sync.RWMutex

	ID     int64
	ChatID int64

	Assistant *Assistant
	statusMsg *telegram.NewMessage
	Data      map[string]any

	filePath string
	track    *state.Track
	loop     int

	// TODO: gortc has no speed control yet. speed is tracked only for
	// display; SetSpeed doesn't actually change playback.
	speed float64

	queue   []*state.Track
	shuffle bool

	*scheduledTimers

	gc         *groupcall.GroupCall
	player     *media.Player
	joined     bool
	joinCancel context.CancelFunc

	playGen atomic.Uint64

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

// Room management

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
	room.leaveCall()
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
			ID:        chatID,
			ChatID:    chatID,
			queue:     []*state.Track{},
			Assistant: ass,
			gc:        groupcall.New(ass.Client),
			speed:     1.0,
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

func (r *RoomState) watch(player *media.Player, gen uint64) {
	go func() {
		done := player.Done()
		if done == nil {
			return
		}
		err := <-done

		if r.IsDestroyed() || OnStreamEnd == nil {
			return
		}
		if err != nil {
			gologging.Error(err)
			return
		}
		if r.playGen.Load() != gen {
			return
		}
		OnStreamEnd(r.ChatID)
	}()
}

func (r *RoomState) IsDestroyed() bool {
	return r.destroyed.Load()
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

func (r *RoomState) Speed() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speed
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track != nil && r.player != nil
}

func (r *RoomState) IsPaused() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track != nil && r.player != nil && r.player.Paused()
}

func (r *RoomState) IsMuted() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track != nil && r.player != nil && r.player.Muted()
}

func (r *RoomState) Position() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.player == nil {
		return 0
	}
	return int(r.player.Position().Seconds())
}

// Playback control

// Play starts playback of a track, or enqueues it if something is already
// playing (unless force is set).
//
// TODO: auto unmute for fplay.
func (r *RoomState) Play(t *state.Track, path string, force ...bool) error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}

	forcePlay := len(force) > 0 && force[0]

	r.mu.Lock()
	if r.Data != nil {
		delete(r.Data, "last_queue")
	}

	shouldQueue := !forcePlay && r.player != nil && r.track != nil
	if shouldQueue {
		r.queue = append(r.queue, t)
		r.mu.Unlock()
		return nil
	}

	if r.track != t {
		r.loop = 0
	}
	r.track = t
	r.filePath = path
	r.mu.Unlock()

	if err := r.play(); err != nil {
		r.mu.Lock()
		r.track = nil
		r.filePath = ""
		r.mu.Unlock()
		return err
	}
	return nil
}

// Pause pauses playback with an optional auto-resume after a duration.
func (r *RoomState) Pause(autoResumeAfter ...time.Duration) (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}

	r.mu.Lock()
	if r.player == nil {
		r.mu.Unlock()
		return false, ErrCallNotJoined
	}
	if r.player.Paused() {
		r.mu.Unlock()
		return true, nil
	}
	r.player.Pause()

	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledResume()

	if len(autoResumeAfter) > 0 && autoResumeAfter[0] > 0 {
		d := autoResumeAfter[0]
		r.scheduledResumeUntil = time.Now().Add(d)
		r.scheduledResumeTimer = time.AfterFunc(d, func() {
			if !r.IsDestroyed() {
				r.Resume()
			}
		})
	}
	r.mu.Unlock()

	return true, nil
}

// Resume resumes playback.
func (r *RoomState) Resume() (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}

	r.mu.Lock()
	if r.player == nil || r.track == nil {
		r.mu.Unlock()
		return false, fmt.Errorf("there are no active music playing")
	}
	if !r.player.Paused() {
		r.mu.Unlock()
		return true, nil
	}
	r.player.Resume()

	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledResume()
	}
	r.mu.Unlock()

	return true, nil
}

// Replay restarts the current track from the beginning.
func (r *RoomState) Replay() error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}

	r.mu.RLock()
	hasTrack := r.track != nil && r.filePath != ""
	r.mu.RUnlock()
	if !hasTrack {
		return fmt.Errorf("no track to replay")
	}

	if err := r.play(); err != nil {
		return err
	}

	r.mu.Lock()
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledResume()
		// r.scheduledTimers.cancelScheduledUnmute()
	}
	r.mu.Unlock()

	return nil
}

// Stop stops playback but keeps the room (and the call connection) alive.
// Use DeleteRoom to also leave the call.
func (r *RoomState) Stop() error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}

	_, file, line, _ := runtime.Caller(1)
	gologging.DebugF("Stop Called from %s:%d", file, line)

	r.mu.Lock()
	r.stopPlayerLocked()
	r.track = nil
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledUnmute()
		r.scheduledTimers.cancelScheduledResume()
		r.scheduledTimers.cancelScheduledSpeed()
	}
	r.mu.Unlock()

	return nil
}

// Seek moves playback position by the given number of seconds (negative
// seeks backward).
func (r *RoomState) Seek(seconds int) error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}

	r.mu.Lock()
	if r.player == nil {
		r.mu.Unlock()
		return ErrCallNotJoined
	}
	track := r.track
	if track == nil {
		r.mu.Unlock()
		return fmt.Errorf("no track to seek")
	}

	currentPos := int(r.player.Position().Seconds())
	if seconds > 0 && track.Duration-currentPos <= seekEndThreshold {
		r.mu.Unlock()
		return fmt.Errorf("cannot seek, track is about to end")
	}

	newPos := currentPos + seconds
	if newPos >= track.Duration {
		newPos = track.Duration - seekSafetyMargin
	}
	if newPos < 0 {
		newPos = 0
	}

	player := r.player
	r.mu.Unlock()

	if err := player.Seek(time.Duration(newPos) * time.Second); err != nil {
		return fmt.Errorf("seek failed: %w", err)
	}

	r.mu.Lock()
	gen := r.playGen.Add(1)
	r.mu.Unlock()

	r.watch(player, gen)
	return nil
}

// SetSpeed adjusts the tracked playback speed with an optional auto-reset.
//
// TODO: gortc does not support runtime speed changes yet. This only updates
// the displayed value; actual playback speed is unaffected.
func (r *RoomState) SetSpeed(speed float64, timeAfterNormal ...time.Duration) error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}
	return nil
}

func (r *RoomState) resetSpeedToNormal() {
	if r.IsDestroyed() {
		return
	}
}

// Mute mutes the bot's outgoing audio with an optional auto-unmute.
func (r *RoomState) Mute(unmuteAfter ...time.Duration) (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}

	r.mu.Lock()
	if r.player == nil {
		r.mu.Unlock()
		return false, ErrCallNotJoined
	}
	if r.player.Muted() {
		r.mu.Unlock()
		return true, nil
	}
	r.player.Mute()

	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledUnmute()

	if len(unmuteAfter) > 0 && unmuteAfter[0] > 0 {
		d := unmuteAfter[0]
		r.scheduledUnmuteUntil = time.Now().Add(d)
		r.scheduledUnmuteTimer = time.AfterFunc(d, func() {
			if !r.IsDestroyed() {
				r.Unmute()
			}
		})
	}
	r.mu.Unlock()

	return true, nil
}

// Unmute restores the bot's outgoing audio.
func (r *RoomState) Unmute() (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}

	r.mu.Lock()
	if r.player == nil {
		r.mu.Unlock()
		return false, ErrCallNotJoined
	}
	if !r.player.Muted() {
		r.mu.Unlock()
		return true, nil
	}
	r.player.Unmute()

	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledUnmute()
	}
	r.mu.Unlock()

	return true, nil
}

// play (re)starts the call's player against the room's current filePath,
// joining the call first if necessary.
func (r *RoomState) play() error {
	r.mu.RLock()
	gc := r.gc
	joined := r.joined
	path := r.filePath
	isVideo := r.track != nil && r.track.Video
	r.mu.RUnlock()

	if gc == nil {
		return ErrCallNotJoined
	}
	if !joined {
		if err := r.joinCall(); err != nil {
			return err
		}
	}

	var src media.Source
	if isVideo {
		src = media.FromFile(path, media.Res720)
	} else {
		src = media.FromFile(path, media.EncodeOptions{Tracks: media.TrackAudio})
	}

	r.mu.Lock()
	r.stopPlayerLocked()
	player := gc.Play(context.Background(), src)
	r.player = player
	gen := r.playGen.Add(1)
	r.mu.Unlock()

	r.watch(player, gen)
	return nil
}

// joinCall joins the room's gortc group call. The context is cancelled
// from leaveCall, which matters because JoinCall retries internally
// (including multi-second flood-wait sleeps) — without cancellation, a
// pending join for a room that's already being torn down would keep
// running in the background and could connect a call no one wants anymore.
func (r *RoomState) joinCall() error {
	r.mu.Lock()
	gc := r.gc
	if gc == nil {
		r.mu.Unlock()
		return ErrCallNotJoined
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.joinCancel = cancel
	chatID := r.ChatID
	r.mu.Unlock()

	err := gc.JoinCall(ctx, chatID)

	r.mu.Lock()
	r.joined = err == nil
	r.mu.Unlock()

	return err
}

// stopPlayerLocked stops the active player, if any. Callers must hold r.mu.
func (r *RoomState) stopPlayerLocked() {
	if r.player != nil {
		r.player.Stop()
		r.player = nil
	}
	r.playGen.Add(1)
}

// leaveCall leaves the underlying gortc group call entirely.
func (r *RoomState) leaveCall() error {
	r.mu.Lock()
	r.stopPlayerLocked()
	r.joined = false
	gc := r.gc
	if r.joinCancel != nil {
		r.joinCancel()
		r.joinCancel = nil
	}
	r.mu.Unlock()

	if gc == nil {
		return nil
	}
	return gc.Leave()
}

// Queue management

// NextTrack retrieves and prepares the next track in queue.
func (r *RoomState) NextTrack() *state.Track {
	if r.IsDestroyed() {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.track != nil && r.loop > 0 {
		r.loop--
		return r.track
	}

	r.releaseFileLocked()

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

	return next
}

// RemoveFromQueue removes a track at index, or clears the queue if index is -1.
func (r *RoomState) RemoveFromQueue(index int) {
	if r.IsDestroyed() {
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

// MoveInQueue moves a track from one position to another.
func (r *RoomState) MoveInQueue(from, to int) {
	if r.IsDestroyed() {
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

// AddTracksToQueue appends multiple tracks to the queue.
func (r *RoomState) AddTracksToQueue(tracks []*state.Track) {
	if r.IsDestroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue = append(r.queue, tracks...)
}

// File cleanup

// isTrackUsed reports whether trackID is playing or queued in any room
// other than skipChatID. Callers must hold roomsMu (read or write).
func isTrackUsed(trackID string, skipChatID int64) bool {
	for _, room := range rooms {
		if room == nil || room.ChatID == skipChatID {
			continue
		}

		room.mu.RLock()
		track := room.track
		queue := room.queue
		room.mu.RUnlock()

		if track != nil && track.ID == trackID {
			return true
		}
		if isTrackInQueue(trackID, queue) {
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

// releaseFileLocked schedules the current track's file for removal if
// unused elsewhere. Callers must already hold r.mu.
func (r *RoomState) releaseFileLocked() {
	if r.track == nil {
		return
	}
	track := r.track

	roomsMu.RLock()
	used := isTrackUsed(track.ID, r.ChatID)
	roomsMu.RUnlock()

	if used {
		gologging.DebugF("file still in use, skipped remove: %s:%s", string(track.Source), track.ID)
		return
	}
	findAndRemove(track, r.ChatID)
}

// cleanupFile schedules the current and next-queued track files for
// removal if unused.
func (r *RoomState) cleanupFile() {
	r.mu.RLock()
	tracks := make([]*state.Track, 0, 3)
	if r.track != nil {
		tracks = append(tracks, r.track)
	}
	tracks = append(tracks, r.queue...)
	if len(tracks) > 2 {
		tracks = tracks[:2]
	}
	chatID := r.ChatID
	r.mu.RUnlock()

	for _, t := range tracks {
		if t == nil || t.ID == "" {
			continue
		}

		roomsMu.RLock()
		used := isTrackUsed(t.ID, chatID)
		roomsMu.RUnlock()

		if used {
			gologging.DebugF("track still in use, skip delete: %s:%s", string(t.Source), t.ID)
			continue
		}
		findAndRemove(t, chatID)
	}
}

// findAndRemove deletes track's downloaded file(s) after FileCleanupDelay,
// re-checking at delete time in case another room picked up the same
// track while waiting (e.g. the same song requested again).
func findAndRemove(track *state.Track, skipChatID int64) {
	remove := func() {
		roomsMu.RLock()
		used := isTrackUsed(track.ID, skipChatID)
		roomsMu.RUnlock()
		if used {
			gologging.DebugF("file claimed before cleanup, skipped: %s:%s", string(track.Source), track.ID)
			return
		}

		kind := "audio"
		if track.Video {
			kind = "video"
		}
		files, err := filepath.Glob(filepath.Join("downloads", kind+"_"+track.ID+"*"))
		if err != nil {
			return
		}
		for _, f := range files {
			os.Remove(f)
			gologging.DebugF("removed unused file: %s", f)
		}
	}

	if FileCleanupDelay <= 0 {
		remove()
		return
	}
	time.AfterFunc(FileCleanupDelay, remove)
}

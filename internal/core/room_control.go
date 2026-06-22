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
	"fmt"
	"runtime"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gortc/media"

	state "yukkimusic/internal/core/models"
)

const (
	seekEndThreshold = 10
	seekSafetyMargin = 5
)

// Play starts playback of a track.
func (r *RoomState) Play(t *state.Track, path string, force ...bool) error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}

	forcePlay := len(force) > 0 && force[0]

	r.mu.Lock()
	if r.Data != nil {
		delete(r.Data, "last_queue")
	}

	shouldQueue := !forcePlay && r.Call != nil && r.Call.IsPlaying() && r.track != nil
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
	r.muted = false
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

// Pause pauses playback with optional auto-resume.
func (r *RoomState) Pause(autoResumeAfter ...time.Duration) (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}
	if r.Call == nil {
		return false, ErrCallNotJoined
	}

	if r.Call.Paused() {
		return true, nil
	}

	r.Call.Pause()

	r.mu.Lock()
	r.muted = false

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
	if !r.IsActiveChat() {
		return false, fmt.Errorf("there are no active music playing")
	}
	if !r.Call.Paused() {
		return true, nil
	}

	r.Call.Resume()

	r.mu.Lock()
	r.muted = false
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
	r.muted = false
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledResume()
		r.scheduledTimers.cancelScheduledUnmute()
	}
	r.mu.Unlock()

	return nil
}

// Stop stops playback completely.
func (r *RoomState) Stop() error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}

	_, file, line, _ := runtime.Caller(1)
	gologging.DebugF("Stop Called from %s:%d", file, line)

	if r.Call != nil {
		r.Call.Leave()
	}

	r.mu.Lock()
	r.track = nil
	r.muted = false
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledUnmute()
		r.scheduledTimers.cancelScheduledResume()
		r.scheduledTimers.cancelScheduledSpeed()
	}
	r.mu.Unlock()

	return nil
}

// Seek moves playback position by specified seconds.
func (r *RoomState) Seek(seconds int) error {
	if r.IsDestroyed() {
		return ErrRoomDestroyed
	}
	if r.Call == nil {
		return ErrCallNotJoined
	}

	r.mu.RLock()
	track := r.track
	r.mu.RUnlock()
	if track == nil {
		return fmt.Errorf("no track to seek")
	}

	currentPos := r.Call.Position()
	if seconds > 0 && track.Duration-int(currentPos.Seconds()) <= seekEndThreshold {
		return fmt.Errorf("cannot seek, track is about to end")
	}

	newPos := int(currentPos.Seconds()) + seconds
	if newPos >= track.Duration {
		newPos = track.Duration - seekSafetyMargin
	}
	if newPos < 0 {
		newPos = 0
	}

	if err := r.Call.SeekTo(time.Duration(newPos) * time.Second); err != nil {
		return fmt.Errorf("seek failed: %w", err)
	}

	r.mu.Lock()
	r.muted = false
	r.mu.Unlock()

	return nil
}

// SetSpeed adjusts playback speed with optional auto-reset.
//
// TODO: gortc didn't support.
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

// Mute mutes playback with optional auto-unmute.
//
// TODO: gortc didn't support.
func (r *RoomState) Mute(unmuteAfter ...time.Duration) (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}
	return false, nil
}

// Unmute unmutes playback.
//
// TODO: gortc didn't support.
func (r *RoomState) Unmute() (bool, error) {
	if r.IsDestroyed() {
		return false, ErrRoomDestroyed
	}
	return false, nil
}

func (r *RoomState) play() error {
	r.mu.RLock()
	path := r.filePath
	isVideo := r.track != nil && r.track.Video
	r.mu.RUnlock()

	var src media.Source
	if isVideo {
		src = media.FromFile(path, media.Res720)
	} else {
		src = media.FromFile(path, media.EncodeOptions{Tracks: media.TrackAudio})
	}

	return r.PlayTrack(src)
}
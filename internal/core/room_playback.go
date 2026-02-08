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
	"fmt"
	"runtime"
	"time"

	"github.com/Laky-64/gologging"

	state "main/internal/core/models"
)

// Play starts playback of a track
func (r *RoomState) Play(t *state.Track, path string, force ...bool) error {
	if r.destroyed.Load() {
		return ErrRoomDestroyed
	}

	forcePlay := len(force) > 0 && force[0]

	r.mu.Lock()
	shouldQueue := !forcePlay && r.playing && r.track != nil
	if shouldQueue {
		r.queue = append(r.queue, t)
		r.mu.Unlock()
		return nil
	}

	r.track = t
	r.playing = true
	r.fpath = path
	r.position = 0
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()
	r.mu.Unlock()

	err := r.p.Play(r)
	if err != nil {
		r.mu.Lock()
		r.track = nil
		r.playing = false
		r.fpath = ""
		r.mu.Unlock()
		return err
	}

	return nil
}

// Pause pauses playback with optional auto-resume
func (r *RoomState) Pause(autoResumeAfter ...time.Duration) (bool, error) {
	if r.destroyed.Load() {
		return false, ErrRoomDestroyed
	}

	r.mu.RLock()
	alreadyPaused := r.paused
	r.mu.RUnlock()

	if alreadyPaused {
		return true, nil
	}

	paused, err := r.p.Pause(r)
	if err != nil {
		return false, err
	}

	r.mu.RLock()
	isMuted := r.muted
	r.mu.RUnlock()

	if isMuted {
		r.Unmute()
	}

	r.mu.Lock()
	r.parse()
	r.paused = true
	r.muted = false
	r.scheduleAutoResume(autoResumeAfter)
	r.mu.Unlock()

	return paused, nil
}

func (r *RoomState) scheduleAutoResume(autoResumeAfter []time.Duration) {
	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledResume()

	if len(autoResumeAfter) > 0 && autoResumeAfter[0] > 0 {
		d := autoResumeAfter[0]
		r.scheduledResumeUntil = time.Now().Add(d)
		r.scheduledResumeTimer = time.AfterFunc(d, func() {
			if !r.destroyed.Load() {
				r.Resume()
			}
		})
	}
}

// Resume resumes playback
func (r *RoomState) Resume() (bool, error) {
	if r.destroyed.Load() {
		return false, ErrRoomDestroyed
	}

	if !r.IsActiveChat() {
		return false, fmt.Errorf("there are no active music playing")
	}

	r.mu.RLock()
	alreadyPlaying := !r.paused
	r.mu.RUnlock()

	if alreadyPlaying {
		return true, nil
	}

	resumed, err := r.p.Resume(r)
	if err != nil {
		return false, err
	}

	r.mu.Lock()
	r.paused = false
	r.muted = false
	r.playing = true
	r.updatedAt = time.Now().Unix()
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledResume()
	}
	r.mu.Unlock()

	return resumed, nil
}

// Replay restarts the current track
func (r *RoomState) Replay() error {
	if r.destroyed.Load() {
		return ErrRoomDestroyed
	}

	r.mu.RLock()
	hasTrack := r.track != nil && r.fpath != ""
	r.mu.RUnlock()

	if !hasTrack {
		return fmt.Errorf("no track to replay")
	}

	r.mu.Lock()
	oldPos := r.position
	r.position = 0
	r.mu.Unlock()

	err := r.p.Play(r)
	if err != nil {
		r.mu.Lock()
		r.position = oldPos
		r.mu.Unlock()
		return err
	}

	r.mu.Lock()
	r.position = 0
	r.paused = false
	r.muted = false
	r.playing = true
	r.updatedAt = time.Now().Unix()
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledResume()
		r.scheduledTimers.cancelScheduledUnmute()
	}
	r.mu.Unlock()

	return nil
}

// Stop stops playback completely
func (r *RoomState) Stop() error {
	if r.destroyed.Load() {
		return ErrRoomDestroyed
	}

	_, file, line, _ := runtime.Caller(1)
	gologging.DebugF("Stop Called from %s:%d", file, line)

	err := r.p.Stop(r)

	r.mu.Lock()
	r.track = nil
	r.position = 0
	r.playing = false
	r.paused = false
	r.muted = false
	r.updatedAt = 0
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledUnmute()
		r.scheduledTimers.cancelScheduledResume()
		r.scheduledTimers.cancelScheduledSpeed()
	}
	r.mu.Unlock()

	return err
}

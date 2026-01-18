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
	r.Lock()
	defer r.Unlock()
	if r.destroyed {
		return ErrRoomDestroyed
	}
	forcePlay := len(force) > 0 && force[0]

	if !forcePlay && r.playing && r.track != nil {
		r.queue = append(r.queue, t)
		return nil
	}

	return r.startPlayback(t, path)
}

func (r *RoomState) startPlayback(t *state.Track, path string) error {
	r.track = t
	r.playing = true
	r.fpath = path

	if err := r.p.Play(r); err != nil {
		r.track = nil
		r.playing = false
		r.fpath = ""
		return err
	}

	r.resetPlaybackState()
	return nil
}

func (r *RoomState) resetPlaybackState() {
	r.position = 0
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()
}

// Pause pauses playback with optional auto-resume
func (r *RoomState) Pause(autoResumeAfter ...time.Duration) (bool, error) {
	if r.Destroyed() {
		return false, ErrRoomDestroyed
	}

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

	r.scheduleAutoResume(autoResumeAfter)

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
			if !r.Destroyed() {
				r.Resume()
			}
		})
	}
}

// Resume resumes playback
func (r *RoomState) Resume() (bool, error) {
	if r.Destroyed() {
		return false, ErrRoomDestroyed
	}

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

	r.updateResumeState()
	r.scheduledTimers.cancelScheduledResume()

	return resumed, nil
}

func (r *RoomState) updateResumeState() {
	r.paused = false
	r.muted = false
	r.playing = true
	r.updatedAt = time.Now().Unix()
}

// Replay restarts the current track
func (r *RoomState) Replay() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.Lock()
	defer r.Unlock()

	if r.track == nil || r.fpath == "" {
		return fmt.Errorf("no track to replay")
	}

	return r.executeReplay()
}

func (r *RoomState) executeReplay() error {
	old := r.position
	r.position = 0

	if err := r.p.Play(r); err != nil {
		r.position = old
		return err
	}

	r.resetPlaybackState()
	r.playing = true
	r.scheduledTimers.cancelScheduledResume()
	r.scheduledTimers.cancelScheduledUnmute()

	return nil
}

// Stop stops playback completely
func (r *RoomState) Stop() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.Lock()
	defer r.Unlock()

	_, file, line, _ := runtime.Caller(1)
	gologging.DebugF("Stop Called from %s:%d", file, line)

	err := r.p.Stop(r)
	r.clearPlaybackState()

	return err
}

func (r *RoomState) clearPlaybackState() {
	r.track = nil
	r.position = 0
	r.playing = false
	r.paused = false
	r.muted = false
	r.updatedAt = 0
	r.scheduledTimers.cancelScheduledUnmute()
	r.scheduledTimers.cancelScheduledResume()
	r.scheduledTimers.cancelScheduledSpeed()
}

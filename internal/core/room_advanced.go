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
	"time"
)

const (
	minSpeed         = 0.50
	maxSpeed         = 4.0
	seekEndThreshold = 10
	seekSafetyMargin = 5
)

type playbackSnapshot struct {
	position int
	paused   bool
	muted    bool
	updated  int64
}

// Seek moves playback position by specified seconds
func (r *RoomState) Seek(seconds int) error {
	if r.destroyed.Load() {
		return ErrRoomDestroyed
	}

	r.Lock()
	if r.track == nil || r.fpath == "" {
		r.Unlock()
		return fmt.Errorf("no track to seek")
	}

	r.parse()

	if seconds > 0 && r.track.Duration-r.position <= seekEndThreshold {
		r.Unlock()
		return fmt.Errorf("cannot seek, track is about to end")
	}

	snapshot := playbackSnapshot{
		position: r.position,
		paused:   r.paused,
		muted:    r.muted,
		updated:  r.updatedAt,
	}

	newPos := r.position + seconds
	if newPos >= r.track.Duration {
		newPos = r.track.Duration - seekSafetyMargin
	}
	if newPos < 0 {
		newPos = 0
	}

	r.position = newPos
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()
	r.Unlock()

	err := r.p.Play(r)
	if err != nil {
		r.Lock()
		r.position = snapshot.position
		r.paused = snapshot.paused
		r.muted = snapshot.muted
		r.updatedAt = snapshot.updated
		r.Unlock()
		return err
	}

	r.RLock()
	wasMuted := snapshot.muted
	r.RUnlock()

	if wasMuted {
		r.p.Unmute(r)
	}

	return nil
}

// SetSpeed adjusts playback speed with optional auto-reset
func (r *RoomState) SetSpeed(
	speed float64,
	timeAfterNormal ...time.Duration,
) error {
	if r.destroyed.Load() {
		return ErrRoomDestroyed
	}

	r.RLock()
	hasTrack := r.track != nil && r.fpath != ""
	currentSpeed := r.speed
	r.RUnlock()

	if !hasTrack {
		return fmt.Errorf("no track to adjust speed")
	}

	if speed < minSpeed || speed > maxSpeed {
		return fmt.Errorf(
			"invalid speed: must be between %.2fx and %.1fx",
			minSpeed,
			maxSpeed,
		)
	}

	if currentSpeed == speed {
		return nil
	}

	r.Lock()
	r.parse()
	r.speed = speed
	r.playing = true
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()
	r.Unlock()

	err := r.p.Play(r)
	if err != nil {
		return err
	}

	r.Lock()
	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledSpeed()

	shouldSchedule := len(timeAfterNormal) > 0 && timeAfterNormal[0] > 0 &&
		speed != 1.0
	if shouldSchedule {
		d := timeAfterNormal[0]
		r.scheduledSpeedUntil = time.Now().Add(d)
		r.scheduledSpeedTimer = time.AfterFunc(d, func() {
			r.resetSpeedToNormal()
		})
	}
	r.Unlock()

	return nil
}

func (r *RoomState) resetSpeedToNormal() {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	if r.track == nil || !r.playing || r.speed == 1.0 {
		r.Unlock()
		return
	}

	r.parse()
	r.speed = 1.0
	r.updatedAt = time.Now().Unix()
	r.Unlock()

	r.p.Play(r)
}

// Mute mutes playback with optional auto-unmute
func (r *RoomState) Mute(unmuteAfter ...time.Duration) (bool, error) {
	if r.destroyed.Load() {
		return false, ErrRoomDestroyed
	}

	r.RLock()
	alreadyMuted := r.muted
	r.RUnlock()

	if alreadyMuted {
		return true, nil
	}

	muted, err := r.p.Mute(r)
	if err != nil {
		return false, err
	}

	r.RLock()
	isPaused := r.paused
	r.RUnlock()

	if isPaused {
		r.Resume()
	} else {
		r.Parse()
	}

	r.Lock()
	r.muted = true
	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledUnmute()

	if len(unmuteAfter) > 0 && unmuteAfter[0] > 0 {
		duration := unmuteAfter[0]
		r.scheduledUnmuteUntil = time.Now().Add(duration)
		r.scheduledUnmuteTimer = time.AfterFunc(duration, func() {
			if !r.destroyed.Load() {
				r.Parse()
				r.Unmute()
			}
		})
	}
	r.Unlock()

	return muted, nil
}

// Unmute unmutes playback
func (r *RoomState) Unmute() (bool, error) {
	if r.destroyed.Load() {
		return false, ErrRoomDestroyed
	}

	unmuted, err := r.p.Unmute(r)
	if err != nil {
		return false, err
	}

	r.Lock()
	r.parse()
	r.muted = false
	r.paused = false
	if r.scheduledTimers != nil {
		r.scheduledTimers.cancelScheduledUnmute()
	}
	r.Unlock()

	return unmuted, nil
}

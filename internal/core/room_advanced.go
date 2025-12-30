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
	r.Lock()
	defer r.Unlock()

	if r.track == nil || r.fpath == "" {
		return fmt.Errorf("no track to seek")
	}

	r.parse()

	if seconds > 0 && r.track.Duration-r.position <= seekEndThreshold {
		return fmt.Errorf("cannot seek, track is about to end")
	}

	return r.executeSeek(seconds)
}

func (r *RoomState) executeSeek(seconds int) error {
	snapshot := r.createPlaybackSnapshot()
	newPos := r.calculateNewPosition(seconds)

	r.position = newPos
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	if err := r.p.Play(r); err != nil {
		r.restorePlaybackSnapshot(snapshot)
		return err
	}

	if snapshot.muted {
		r.p.Unmute(r)
	}

	return nil
}

func (r *RoomState) calculateNewPosition(seconds int) int {
	newPos := r.position + seconds

	if newPos >= r.track.Duration {
		return r.track.Duration - seekSafetyMargin
	}
	if newPos < 0 {
		return 0
	}

	return newPos
}

func (r *RoomState) createPlaybackSnapshot() playbackSnapshot {
	return playbackSnapshot{
		position: r.position,
		paused:   r.paused,
		muted:    r.muted,
		updated:  r.updatedAt,
	}
}

func (r *RoomState) restorePlaybackSnapshot(snap playbackSnapshot) {
	r.position = snap.position
	r.paused = snap.paused
	r.muted = snap.muted
	r.updatedAt = snap.updated
}

// SetSpeed adjusts playback speed with optional auto-reset
func (r *RoomState) SetSpeed(speed float64, timeAfterNormal ...time.Duration) error {
	r.Lock()
	defer r.Unlock()

	if err := r.validateSpeedChange(speed); err != nil {
		return err
	}

	if r.speed == speed {
		return nil
	}

	return r.executeSpeedChange(speed, timeAfterNormal)
}

func (r *RoomState) validateSpeedChange(speed float64) error {
	if r.track == nil || r.fpath == "" {
		return fmt.Errorf("no track to adjust speed")
	}

	if speed < minSpeed || speed > maxSpeed {
		return fmt.Errorf("invalid speed: must be between %.2fx and %.1fx", minSpeed, maxSpeed)
	}

	return nil
}

func (r *RoomState) executeSpeedChange(speed float64, timeAfterNormal []time.Duration) error {
	r.parse()
	r.speed = speed
	r.playing = true
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	if err := r.p.Play(r); err != nil {
		return err
	}

	r.scheduleSpeedReset(speed, timeAfterNormal)
	return nil
}

func (r *RoomState) scheduleSpeedReset(speed float64, timeAfterNormal []time.Duration) {
	if r.scheduledTimers == nil {
		r.scheduledTimers = &scheduledTimers{}
	}
	r.scheduledTimers.cancelScheduledSpeed()

	if !r.shouldScheduleSpeedReset(speed, timeAfterNormal) {
		return
	}

	d := timeAfterNormal[0]
	r.scheduledSpeedUntil = time.Now().Add(d)
	r.scheduledSpeedTimer = time.AfterFunc(d, r.resetSpeedToNormal)
}

func (r *RoomState) shouldScheduleSpeedReset(speed float64, timeAfterNormal []time.Duration) bool {
	return len(timeAfterNormal) > 0 && timeAfterNormal[0] > 0 && speed != 1.0
}

func (r *RoomState) resetSpeedToNormal() {
	r.Lock()
	defer r.Unlock()

	if r.track != nil && r.playing && r.speed != 1.0 {
		r.parse()
		r.speed = 1.0
		r.p.Play(r)
		r.updatedAt = time.Now().Unix()
	}
}

// Mute mutes playback with optional auto-unmute
func (r *RoomState) Mute(unmuteAfter ...time.Duration) (bool, error) {
	if r.IsMuted() {
		return true, nil
	}

	muted, err := r.p.Mute(r)
	if err != nil {
		return false, err
	}

	r.handleMuteStateTransition()
	r.scheduleAutoUnmute(unmuteAfter)

	return muted, nil
}

func (r *RoomState) handleMuteStateTransition() {
	if r.IsPaused() {
		r.Resume()
	} else {
		r.Parse()
	}

	r.Lock()
	defer r.Unlock()
	r.muted = true
}

func (r *RoomState) scheduleAutoUnmute(unmuteAfter []time.Duration) {
	r.Lock()
	defer r.Unlock()

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
}

// Unmute unmutes playback
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

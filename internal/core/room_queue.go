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
	"math/rand"
	"time"

	state "main/internal/core/models"
)

// NextTrack retrieves and prepares the next track in queue
func (r *RoomState) NextTrack() *state.Track {
	if r.destroyed.Load() {
		return nil
	}

	r.Lock()
	defer r.Unlock()

	if r.shouldLoopCurrentTrack() {
		return r.loopCurrentTrack()
	}

	r.releaseFile()

	if len(r.queue) == 0 {
		return nil
	}

	return r.dequeueNextTrack()
}

func (r *RoomState) shouldLoopCurrentTrack() bool {
	return r.track != nil && r.loop > 0
}

func (r *RoomState) loopCurrentTrack() *state.Track {
	r.position = 0
	r.playing = true
	r.paused = false
	r.muted = false
	r.loop--
	r.updatedAt = time.Now().Unix()
	return r.track
}

func (r *RoomState) dequeueNextTrack() *state.Track {
	index := r.selectNextTrackIndex()
	next := r.queue[index]
	r.removeTrackAtIndex(index)
	r.prepareNextTrack(next)
	return next
}

func (r *RoomState) selectNextTrackIndex() int {
	if r.shuffle {
		return rand.Intn(len(r.queue))
	}
	return 0
}

func (r *RoomState) removeTrackAtIndex(index int) {
	r.queue = append(r.queue[:index], r.queue[index+1:]...)
}

func (r *RoomState) prepareNextTrack(track *state.Track) {
	r.track = track
	r.position = 0
	r.playing = false
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()
}

// RemoveFromQueue removes track(s) from queue
func (r *RoomState) RemoveFromQueue(index int) {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()

	if index == -1 {
		r.queue = []*state.Track{}
		return
	}

	if index >= 0 && index < len(r.queue) {
		r.queue = append(r.queue[:index], r.queue[index+1:]...)
	}
}

// MoveInQueue moves a track from one position to another
func (r *RoomState) MoveInQueue(from, to int) {
	if r.destroyed.Load() {
		return
	}

	r.Lock()
	defer r.Unlock()

	if !r.isValidMove(from, to) {
		return
	}

	r.executeMoveOperation(from, to)
}

func (r *RoomState) isValidMove(from, to int) bool {
	return from >= 0 && from < len(r.queue) &&
		to >= 0 && to < len(r.queue) &&
		from != to
}

func (r *RoomState) executeMoveOperation(from, to int) {
	item := r.queue[from]
	r.queue = append(r.queue[:from], r.queue[from+1:]...)

	if to >= len(r.queue) {
		r.queue = append(r.queue, item)
	} else {
		r.queue = append(r.queue[:to], append([]*state.Track{item}, r.queue[to:]...)...)
	}
}

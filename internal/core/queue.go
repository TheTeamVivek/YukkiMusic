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
	"math/rand"
	"slices"

	state "yukkimusic/internal/core/models"
)

// queue is the room's playback queue. Its methods are lock-free; callers must
// hold the owning Room's mutex.
type queue struct {
	tracks  []*state.Track
	shuffle bool
}

func (q *queue) len() int {
	return len(q.tracks)
}

// all returns a snapshot copy of the queued tracks.
func (q *queue) all() []*state.Track {
	return slices.Clone(q.tracks)
}

// push appends the given tracks, skipping nil entries.
func (q *queue) push(tracks ...*state.Track) {
	for _, t := range tracks {
		if t != nil {
			q.tracks = append(q.tracks, t)
		}
	}
}

// remove deletes the track at the zero-based index, if it is in range.
func (q *queue) remove(index int) {
	if index < 0 || index >= len(q.tracks) {
		return
	}
	q.tracks = slices.Delete(q.tracks, index, index+1)
}

// removeByID deletes the first track matching the given ID.
func (q *queue) removeByID(id string) bool {
	for i, t := range q.tracks {
		if t != nil && t.ID == id {
			q.tracks = slices.Delete(q.tracks, i, i+1)
			return true
		}
	}
	return false
}

// move moves the track at from to position to, both 0-based.
func (q *queue) move(from, to int) {
	if from < 0 || from >= len(q.tracks) ||
		to < 0 || to >= len(q.tracks) ||
		from == to {
		return
	}

	track := q.tracks[from]
	q.tracks = slices.Delete(q.tracks, from, from+1)
	q.tracks = slices.Insert(q.tracks, to, track)
}

func (q *queue) clear() {
	q.tracks = nil
}

// pop removes and returns the next track, honoring shuffle, or nil when the
// queue is empty.
func (q *queue) pop() *state.Track {
	if len(q.tracks) == 0 {
		return nil
	}

	index := 0
	if q.shuffle {
		index = rand.Intn(len(q.tracks))
	}

	next := q.tracks[index]
	q.tracks = slices.Delete(q.tracks, index, index+1)
	return next
}

// Queue returns a snapshot copy of the queued tracks.
func (r *Room) Queue() []*state.Track {
	if r.Destroyed() {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queue.all()
}

// QueueLen returns the number of queued tracks.
func (r *Room) QueueLen() int {
	if r.Destroyed() {
		return 0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queue.len()
}

// AddTracks appends the given tracks to the queue, skipping nil entries.
func (r *Room) AddTracks(tracks ...*state.Track) {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue.push(tracks...)
}

// RemoveTrack removes the queued track at the zero-based index. It is a no-op
// when index is out of range.
func (r *Room) RemoveTrack(index int) {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue.remove(index)
}

// RemoveTrackByID removes the first queued track with the given ID.
// It returns true if a track was found and removed, or false otherwise.
func (r *Room) RemoveTrackByID(id string) bool {
	if r.Destroyed() {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.queue.removeByID(id)
}

// MoveTrack moves the queued track at from to position to, where both indices
// are 0-based. It is a no-op when either index is out of range or they match.
func (r *Room) MoveTrack(from, to int) {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue.move(from, to)
}

// ClearQueue removes every queued track.
func (r *Room) ClearQueue() {
	if r.Destroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue.clear()
}

// NextTrack advances playback and returns the next track to play, or nil when
// the queue is drained.
//
// While a repeat is pending it returns the active track and consumes one
// repeat; otherwise it removes the next queued track (a random one when
// shuffle is enabled) and makes it active.
func (r *Room) NextTrack() *state.Track {
	if r.Destroyed() {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.player.track != nil && r.player.loop > 0 {
		r.player.loop--
		r.player.reset()
		return r.player.track
	}

	r.releaseFile()

	next := r.queue.pop()
	if next == nil {
		return nil
	}

	r.player.track = next
	r.player.reset()
	return next
}

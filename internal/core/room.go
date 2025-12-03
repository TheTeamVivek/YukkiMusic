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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/state"
	"main/ntgcalls"
)

var (
	rooms   = make(map[int64]*RoomState)
	roomsMu sync.RWMutex
	logger  = gologging.GetLogger("roomstate")
)

type RoomState struct {
	sync.RWMutex
	ChatID    int64
	Track     *state.Track
	Position  int
	Playing   bool
	Muted     bool
	Paused    bool
	UpdatedAt int64
	FilePath  string
	Queue     []*state.Track
	Speed     float64
	Shuffle   bool

	Loop  int
	cplay bool

	mystic *telegram.NewMessage
	*ScheduledTimers
}

func (r *RoomState) SetCPlay(isCPlay bool) {
	r.Lock()
	defer r.Unlock()
	r.cplay = isCPlay
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
	return r.Track != nil && r.Playing
}

func (r *RoomState) IsPaused() bool {
	r.RLock()
	defer r.RUnlock()
	return r.Paused && r.Track != nil && r.Playing
}

func (r *RoomState) IsMuted() bool {
	r.RLock()
	defer r.RUnlock()
	return r.Muted && r.Track != nil && r.Playing
}

func (r *RoomState) GetSpeed() float64 {
	r.RLock()
	defer r.RUnlock()
	return r.Speed
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

func DeleteRoom(chatID int64) {
	_, file, line, _ := runtime.Caller(1)

	logger.DebugF("DeleteRoom Called from %s:%d", file, line)

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
				ChatID: chatID,
				Queue:  []*state.Track{},
				Speed:  1.0,
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

func (r *RoomState) Destroy() {
	_, file, line, _ := runtime.Caller(1)

	logger.DebugF("Destroy Called from %s:%d", file, line)

	r.Stop()
	r.cleanupFile()
	roomsMu.Lock()
	defer roomsMu.Unlock()
	delete(rooms, r.ChatID)
}

func (r *RoomState) parse() {
	if r == nil || r.Track == nil || r.UpdatedAt == 0 {
		return
	}

	current := time.Now().Unix()
	elapsed := float64(current - r.UpdatedAt)

	if r.Playing && !r.Paused {
		r.Position += int(elapsed * r.Speed)
		if r.Position >= r.Track.Duration {
			r.Position = r.Track.Duration
			r.Playing = false
		}
	}
	r.UpdatedAt = current
}

func (r *RoomState) SetShuffle(enabled bool) {
	r.Lock()
	defer r.Unlock()
	r.Shuffle = enabled
}

func (r *RoomState) Play(t *state.Track, path string, force ...bool) error {
	r.Lock()
	defer r.Unlock()

	forcePlay := len(force) > 0 && force[0]

	if !forcePlay && r.Playing && r.Track != nil {
		r.Queue = append(r.Queue, t)
		return nil
	}

	err := Ntg.Play(r.ChatID, getMediaDescription(path, 0, r.Speed, t.Video))
	if err != nil {
		return err
	}

	r.Track = t
	r.Position = 0
	r.Playing = true
	r.Paused = false
	r.Muted = false
	r.FilePath = path
	r.UpdatedAt = time.Now().Unix()
	return nil
}

func (r *RoomState) Pause(autoResumeAfter ...time.Duration) (bool, error) {
	if r.IsPaused() {
		return true, nil
	}

	paused, err := Ntg.Pause(r.ChatID)
	if err != nil {
		return false, err
	}

	if r.IsMuted() {
		r.Unmute()
	}

	r.Lock()
	defer r.Unlock()

	r.parse()
	r.Paused = true
	r.Muted = false

	if r.ScheduledTimers == nil {
		r.ScheduledTimers = &ScheduledTimers{}
	}
	r.ScheduledTimers.cancelScheduledResume()

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

	resumed, err := Ntg.Resume(r.ChatID)
	if err != nil {
		return false, err
	}

	r.Paused = false
	r.Muted = false
	r.Playing = true
	r.UpdatedAt = time.Now().Unix()

	r.ScheduledTimers.cancelScheduledResume()
	return resumed, nil
}

func (r *RoomState) Replay() error {
	r.Lock()
	defer r.Unlock()

	if r.Track == nil || r.FilePath == "" {
		return fmt.Errorf("no track to replay")
	}

	err := Ntg.Play(r.ChatID, getMediaDescription(r.FilePath, 0, r.Speed, r.Track.Video))
	if err != nil {
		return err
	}

	r.Position = 0
	r.Playing = true
	r.Paused = false
	r.Muted = false
	r.UpdatedAt = time.Now().Unix()

	r.ScheduledTimers.cancelScheduledResume()
	r.ScheduledTimers.cancelScheduledUnmute()
	return nil
}

func (r *RoomState) Seek(seconds int) error {
	r.Lock()
	defer r.Unlock()

	if r.Track == nil || r.FilePath == "" {
		return fmt.Errorf("no track to seek")
	}

	r.parse()

	if seconds > 0 && r.Track.Duration-r.Position <= 10 {
		return fmt.Errorf("cannot seek, track is about to end")
	}

	newPos := r.Position + seconds
	if newPos >= r.Track.Duration {
		newPos = r.Track.Duration - 5
	}
	if newPos < 0 {
		newPos = 0
	}

	err := Ntg.Play(r.ChatID, getMediaDescription(r.FilePath, newPos, r.Speed))
	if err != nil {
		return err
	}

	if r.Muted {
		Ntg.UnMute(r.ChatID)
	}
	r.Position = newPos
	r.Playing = true
	r.Paused = false
	r.Muted = false
	r.UpdatedAt = time.Now().Unix()
	return nil
}

func (r *RoomState) SetSpeed(speed float64, timeAfterNormal ...time.Duration) error {
	r.Lock()
	defer r.Unlock()

	if r.Track == nil || r.FilePath == "" {
		return fmt.Errorf("no track to adjust speed")
	}
	if speed < 0.50 || speed > 4.0 {
		return fmt.Errorf("invalid speed: must be between 0.50x and 4.0x")
	}
	if r.Speed == speed {
		return nil
	}

	r.parse()
	r.Speed = speed
	file := r.FilePath
	pos := r.Position
	r.Playing = true
	r.Paused = false
	r.Muted = false
	r.UpdatedAt = time.Now().Unix()

	err := Ntg.Play(r.ChatID, getMediaDescription(file, pos, speed, r.Track.Video))
	if err != nil {
		return err
	}

	if r.ScheduledTimers == nil {
		r.ScheduledTimers = &ScheduledTimers{}
	}
	r.ScheduledTimers.cancelScheduledSpeed()

	if len(timeAfterNormal) > 0 && timeAfterNormal[0] > 0 && speed != 1.0 {
		d := timeAfterNormal[0]
		r.scheduledSpeedUntil = time.Now().Add(d)

		r.scheduledSpeedTimer = time.AfterFunc(d, func() {
			r.Lock()
			defer r.Unlock()
			if r.Track != nil && r.Playing && r.Speed != 1.0 {
				r.parse()
				r.Speed = 1.0
				Ntg.Play(r.ChatID, getMediaDescription(r.FilePath, r.Position, 1.0, r.Track.Video))
				r.UpdatedAt = time.Now().Unix()
			}
		})
	}
	return nil
}

func (r *RoomState) Mute(unmuteAfter ...time.Duration) (bool, error) {
	if r.IsMuted() {
		return true, nil
	}

	muted, err := Ntg.Mute(r.ChatID)
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
	r.Muted = true
	if r.ScheduledTimers == nil {
		r.ScheduledTimers = &ScheduledTimers{}
	}
	r.ScheduledTimers.cancelScheduledUnmute()

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

	unmuted, err := Ntg.UnMute(r.ChatID)
	if err != nil {
		return false, err
	}
	r.parse()
	r.Muted = false
	r.Paused = false
	r.ScheduledTimers.cancelScheduledUnmute()
	return unmuted, nil
}

func (r *RoomState) Stop() error {
	r.Lock()
	defer r.Unlock()

	_, file, line, _ := runtime.Caller(1)

	logger.DebugF("Stop Called from %s:%d", file, line)

	err := Ntg.Stop(r.ChatID)

	r.Track = nil
	r.Position = 0
	r.Playing = false
	r.Paused = false
	r.Muted = false
	r.UpdatedAt = 0
	r.ScheduledTimers.cancelScheduledUnmute()
	r.ScheduledTimers.cancelScheduledResume()
	r.ScheduledTimers.cancelScheduledSpeed()
	return err
}

func (r *RoomState) NextTrack() *state.Track {
	r.Lock()
	defer r.Unlock()

	if r.Track != nil && r.Loop > 0 {
		r.Position = 0
		r.Playing = true
		r.Paused = false
		r.Muted = false
		r.Loop--
		r.UpdatedAt = time.Now().Unix()
		return r.Track
	}

	r.releaseFile()

	if len(r.Queue) == 0 {
		return nil
	}

	index := 0
	if r.Shuffle {
		index = rand.Intn(len(r.Queue))
	}

	next := r.Queue[index]
	r.Queue = append(r.Queue[:index], r.Queue[index+1:]...)

	r.Track = next
	r.Position = 0
	r.Playing = false
	r.Paused = false
	r.Muted = false
	r.UpdatedAt = time.Now().Unix()
	return next
}

func (r *RoomState) RemoveFromQueue(index int) {
	r.Lock()
	defer r.Unlock()

	if index == -1 {
		r.Queue = []*state.Track{}
		return
	}

	if index < 0 || index >= len(r.Queue) {
		return
	}

	r.Queue = append(r.Queue[:index], r.Queue[index+1:]...)
}

func (r *RoomState) MoveInQueue(from, to int) {
	r.Lock()
	defer r.Unlock()

	if from < 0 || from >= len(r.Queue) || to < 0 || to >= len(r.Queue) || from == to {
		return
	}
	item := r.Queue[from]
	r.Queue = append(r.Queue[:from], r.Queue[from+1:]...)
	if to >= len(r.Queue) {
		r.Queue = append(r.Queue, item)
	} else {
		r.Queue = append(r.Queue[:to], append([]*state.Track{item}, r.Queue[to:]...)...)
	}
}

func getMediaDescription(url string, seek int, speed float64, isVideo bool) ntgcalls.MediaDescription {
	if speed < 0.5 {
		speed = 0.5
	} else if speed > 4.0 {
		speed = 4.0
	}

	audio := &ntgcalls.AudioDescription{
		MediaSource:  ntgcalls.MediaSourceShell,
		SampleRate:   96000,
		ChannelCount: 2,
	}

	baseCmd := "ffmpeg "
	if seek > 0 {
		baseCmd += "-ss " + strconv.Itoa(seek) + " "
	}
	baseCmd += "-v warning -i \"" + url + "\" "

	audioCmd := baseCmd
	audioCmd += "-filter:a \"atempo=" + strconv.FormatFloat(speed, 'f', 2, 64) + "\" "
	audioCmd += "-f s16le -ac " + strconv.Itoa(int(audio.ChannelCount)) + " "
	audioCmd += "-ar " + strconv.Itoa(int(audio.SampleRate)) + " "
	audioCmd += "pipe:1"
	audio.Input = audioCmd

	if !isVideo {
		return ntgcalls.MediaDescription{
			Microphone: audio,
		}
	}

	w, h := getVideoDimensions(url)
	if w <= 0 || h <= 0 {
		w = 1280
		h = 720
	}

	maxW := 1280
	maxH := 720

	if w > maxW {
		h = h * maxW / w
		w = maxW
	}
	if h > maxH {
		w = w * maxH / h
		h = maxH
	}

	if w%2 != 0 {
		w--
	}
	if h%2 != 0 {
		h--
	}

	video := &ntgcalls.VideoDescription{
		MediaSource: ntgcalls.MediaSourceShell,
		Width:       int16(w),
		Height:      int16(h),
		Fps:         30,
	}

	videoSpeed := 1.0 / speed
	videoFilter := "setpts=" + strconv.FormatFloat(videoSpeed, 'f', 4, 64) + "*PTS,scale=" + strconv.Itoa(w) + ":" + strconv.Itoa(h)

	videoCmd := baseCmd
	videoCmd += "-filter:v \"" + videoFilter + "\" "
	videoCmd += "-f rawvideo -r " + strconv.Itoa(int(video.Fps)) + " -pix_fmt yuv420p "
	videoCmd += "pipe:1"
	video.Input = videoCmd

	return ntgcalls.MediaDescription{
		Microphone: audio,
		Camera:     video,
	}
}

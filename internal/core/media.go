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
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"yukkimusic/internal/utils"

	"yukkimusic/ntgcalls"

	state "yukkimusic/internal/core/models"
)

const (
	minSpeed         = 0.50
	maxSpeed         = 4.0
	seekEndThreshold = 10
	seekSafetyMargin = 5

	audioSampleRate = 96000
	audioChannels   = 2

	defaultVideoWidth  = 1280
	defaultVideoHeight = 720
	defaultVideoFPS    = 30

	ffprobeTimeout = 7 * time.Second
)

// Playback control

// Play starts playback of a track
func (r *Room) Play(t *state.Track, path string, force ...bool) error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	forcePlay := len(force) > 0 && force[0]

	r.mu.Lock()
	shouldQueue := !forcePlay && r.player.playing && r.player.track != nil
	if shouldQueue {
		r.queue.push(t)
		r.mu.Unlock()
		return nil
	}

	if r.player.track != t {
		r.player.loop = 0
	}
	r.player.track = t
	r.player.playing = true
	r.player.filePath = path
	r.player.position = 0
	r.player.paused = false
	r.player.muted = false
	r.player.updatedAt = time.Now().Unix()
	r.mu.Unlock()

	err := r.play()
	if err != nil {
		r.mu.Lock()
		r.player.track = nil
		r.player.playing = false
		r.player.filePath = ""
		r.mu.Unlock()
		return err
	}

	return nil
}

// Pause pauses playback with optional auto-resume
func (r *Room) Pause() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.mu.RLock()
	alreadyPaused := r.player.paused
	r.mu.RUnlock()

	if alreadyPaused {
		return nil
	}

	if _, err := r.Assistant.Ntg.Pause(r.ChatID); err != nil {
		return err
	}

	r.mu.RLock()
	isMuted := r.player.muted
	r.mu.RUnlock()

	if isMuted {
		r.Unmute()
	}

	r.mu.Lock()
	r.player.sync()
	r.player.paused = true
	r.player.muted = false
	r.mu.Unlock()

	return nil
}

// Resume resumes playback
func (r *Room) Resume() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	if !r.Active() {
		return fmt.Errorf("there are no active music playing")
	}

	r.mu.RLock()
	alreadyPlaying := !r.player.paused
	r.mu.RUnlock()

	if alreadyPlaying {
		return nil
	}

	if _, err := r.Assistant.Ntg.Resume(r.ChatID); err != nil {
		return err
	}

	r.mu.Lock()
	r.player.paused = false
	r.player.muted = false
	r.player.playing = true
	r.player.updatedAt = time.Now().Unix()
	r.mu.Unlock()

	return nil
}

// Replay restarts the current track
func (r *Room) Replay() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.mu.RLock()
	hasTrack := r.player.track != nil && r.player.filePath != ""
	r.mu.RUnlock()

	if !hasTrack {
		return fmt.Errorf("no track to replay")
	}

	r.mu.Lock()
	oldPos := r.player.position
	r.player.position = 0
	r.mu.Unlock()

	err := r.play()
	if err != nil {
		r.mu.Lock()
		r.player.position = oldPos
		r.mu.Unlock()
		return err
	}

	r.mu.Lock()
	r.player.position = 0
	r.player.paused = false
	r.player.muted = false
	r.player.playing = true
	r.player.updatedAt = time.Now().Unix()
	r.mu.Unlock()

	return nil
}

// Stop stops playback completely
func (r *Room) Stop() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	_, file, line, _ := runtime.Caller(1)
	utils.Log.Debugf("Stop Called from %s:%d", file, line)

	err := r.Assistant.Ntg.Stop(r.ChatID)

	r.mu.Lock()
	r.player.track = nil
	r.player.position = 0
	r.player.playing = false
	r.player.paused = false
	r.player.muted = false
	r.player.updatedAt = 0
	r.mu.Unlock()

	return err
}

// Seek moves playback position by specified seconds
func (r *Room) Seek(seconds int) error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.mu.Lock()
	if r.player.track == nil || r.player.filePath == "" {
		r.mu.Unlock()
		return fmt.Errorf("no track to seek")
	}

	r.player.sync()

	if seconds > 0 && r.player.track.Duration-r.player.position <= seekEndThreshold {
		r.mu.Unlock()
		return fmt.Errorf("cannot seek, track is about to end")
	}

	snapshot := struct {
		position int
		paused   bool
		muted    bool
		updated  int64
	}{
		position: r.player.position,
		paused:   r.player.paused,
		muted:    r.player.muted,
		updated:  r.player.updatedAt,
	}

	newPos := r.player.position + seconds
	if newPos >= r.player.track.Duration {
		newPos = r.player.track.Duration - seekSafetyMargin
	}
	if newPos < 0 {
		newPos = 0
	}

	r.player.position = newPos
	r.player.paused = false
	r.player.muted = false
	r.player.updatedAt = time.Now().Unix()
	r.mu.Unlock()

	err := r.play()
	if err != nil {
		r.mu.Lock()
		r.player.position = snapshot.position
		r.player.paused = snapshot.paused
		r.player.muted = snapshot.muted
		r.player.updatedAt = snapshot.updated
		r.mu.Unlock()
		return err
	}

	if snapshot.muted {
		r.Assistant.Ntg.Unmute(r.ChatID)
	}

	return nil
}

// SetSpeed adjusts playback speed
func (r *Room) SetSpeed(speed float64) error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.mu.RLock()
	hasTrack := r.player.track != nil && r.player.filePath != ""
	currentSpeed := r.player.speed
	r.mu.RUnlock()

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

	r.mu.Lock()
	r.player.sync()
	r.player.speed = speed
	r.player.playing = true
	r.player.paused = false
	r.player.muted = false
	r.player.updatedAt = time.Now().Unix()
	r.mu.Unlock()

	return r.play()
}

// Mute mutes playback
func (r *Room) Mute() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	r.mu.RLock()
	alreadyMuted := r.player.muted
	r.mu.RUnlock()

	if alreadyMuted {
		return nil
	}

	if _, err := r.Assistant.Ntg.Mute(r.ChatID); err != nil {
		return err
	}

	r.mu.RLock()
	isPaused := r.player.paused
	r.mu.RUnlock()

	if isPaused {
		r.Resume()
	} else {
		r.Refresh()
	}

	r.mu.Lock()
	r.player.muted = true
	r.mu.Unlock()

	return nil
}

// Unmute unmutes playback
func (r *Room) Unmute() error {
	if r.Destroyed() {
		return ErrRoomDestroyed
	}

	if _, err := r.Assistant.Ntg.Unmute(r.ChatID); err != nil {
		return err
	}

	r.mu.Lock()
	r.player.sync()
	r.player.muted = false
	r.player.paused = false
	r.mu.Unlock()

	return nil
}

// play pushes the current track to the assistant's ntgcalls stream.
func (r *Room) play() error {
	desc := getMediaDescription(r.player.filePath, r.player.position, r.player.speed, r.player.track.Video)
	return r.Assistant.Ntg.Play(r.ChatID, desc)
}

// Media pipeline

// getMediaDescription builds the ntgcalls media description for a local file
// or remote stream, seeking to pos seconds and applying the given speed.
func getMediaDescription(
	url string,
	pos int,
	speed float64,
	isVideo bool,
) ntgcalls.MediaDescription {
	if speed < minSpeed {
		speed = minSpeed
	} else if speed > maxSpeed {
		speed = maxSpeed
	}

	baseCmd := "ffmpeg "
	if isStreamURL(url) {
		baseCmd += "-reconnect 1 -reconnect_streamed 1 -reconnect_delay_max 5 "
	}
	if pos > 0 {
		baseCmd += "-ss " + strconv.Itoa(pos) + " "
	}
	baseCmd += "-v warning -i \"" + url + "\" "

	audio := getAudioPipeline(baseCmd, speed)
	if !isVideo {
		return ntgcalls.MediaDescription{
			Microphone: audio,
		}
	}

	return ntgcalls.MediaDescription{
		Microphone: audio,
		Camera:     getVideoPipeline(baseCmd, url, speed),
	}
}

// getAudioPipeline builds the ffmpeg audio pipeline mirroring audio at the
// requested speed.
func getAudioPipeline(
	baseCmd string,
	speed float64,
) *ntgcalls.AudioDescription {
	audio := &ntgcalls.AudioDescription{
		MediaSource:  ntgcalls.MediaSourceShell,
		SampleRate:   audioSampleRate,
		ChannelCount: audioChannels,
	}
	audio.Input = baseCmd +
		"-filter:a \"atempo=" + strconv.FormatFloat(speed, 'f', 2, 64) + "\" " +
		"-f s16le -ac " + strconv.Itoa(audioChannels) + " " +
		"-ar " + strconv.Itoa(audioSampleRate) + " pipe:1"

	return audio
}

// getVideoPipeline builds the ffmpeg video pipeline scaled to at most
// defaultVideoWidth x defaultVideoHeight.
func getVideoPipeline(
	baseCmd string,
	url string,
	speed float64,
) *ntgcalls.VideoDescription {
	w, h, fps, filter := normalizeVideo(url, speed)
	video := &ntgcalls.VideoDescription{
		MediaSource: ntgcalls.MediaSourceShell,
		Width:       int16(w),
		Height:      int16(h),
		FPS:         uint8(fps),
	}
	video.Input = baseCmd +
		"-filter:v \"" + filter + "\" " +
		"-f rawvideo -r " + strconv.Itoa(fps) + " -pix_fmt yuv420p pipe:1"

	return video
}

// normalizeVideo probes the video dimensions and returns the scaled size, fps
// and the ffmpeg filter that applies the playback speed.
func normalizeVideo(path string, speed float64) (int, int, int, string) {
	if speed <= 0 {
		speed = 1.0
	}

	w, h := getVideoDimensions(path)
	if w <= 0 || h <= 0 {
		w = defaultVideoWidth
		h = defaultVideoHeight
	}
	if w > defaultVideoWidth {
		h = h * defaultVideoWidth / w
		w = defaultVideoWidth
	}
	if h > defaultVideoHeight {
		w = w * defaultVideoHeight / h
		h = defaultVideoHeight
	}
	if w%2 != 0 {
		w--
	}
	if h%2 != 0 {
		h--
	}

	filter := fmt.Sprintf("setpts=%.4f*PTS,scale=%d:%d", 1.0/speed, w, h)
	return w, h, defaultVideoFPS, filter
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

// getVideoDimensions probes a media file with ffprobe and returns its
// dimensions, or (0, 0) if the probe fails or times out.
func getVideoDimensions(filePath string) (int, int) {
	ctx, cancel := context.WithTimeout(context.Background(), ffprobeTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		utils.Log.Error("[getVideoDimensions] ffprobe timed out for " + filePath)
		return 0, 0
	}
	if err != nil {
		utils.Log.Error(
			"[getVideoDimensions] ffprobe failed for " + filePath + " : " + err.Error(),
		)
		return 0, 0
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		utils.Log.Error(
			"[getVideoDimensions] failed to parse ffprobe JSON for " + filePath + " : " + err.Error(),
		)
		return 0, 0
	}

	for _, s := range probe.Streams {
		if s.CodecType == "video" && s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}
	for _, s := range probe.Streams {
		if s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}

	utils.Log.Error("[getVideoDimensions] no valid video stream found for " + filePath)
	return 0, 0
}

// isStreamURL reports whether path points to a remote HTTP(S) stream.
func isStreamURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}

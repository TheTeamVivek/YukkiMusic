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
	jsonv2 "encoding/json/v2"
	"fmt"
	"os/exec"
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

// Play starts playback of a track, or queues it when something is playing.
func (r *Room) Play(t *state.Track, path string, force ...bool) error {
	forcePlay := len(force) > 0 && force[0]

	r.mu.Lock()
	if !forcePlay && r.track != nil && r.playback() != PlaybackIdle {
		r.queue.push(t)
		r.mu.Unlock()
		return nil
	}

	if r.track != t {
		r.loop = 0
	}
	r.track = t
	r.filePath = path
	r.offset, r.anchor = 0, time.Time{}
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

// Pause pauses playback until Resume is called.
func (r *Room) Pause() error {
	s := r.playback()
	if s == PlaybackPaused {
		return nil
	}

	if _, err := r.Assistant.Ntg.Pause(r.ChatID); err != nil {
		return err
	}

	r.mu.Lock()
	r.offset = r.position()
	r.anchor = time.Time{}
	r.mu.Unlock()

	if s == PlaybackMuted {
		r.Unmute()
	}
	return nil
}

// Resume resumes a paused stream.
func (r *Room) Resume() error {
	switch r.playback() {
	case PlaybackIdle:
		return fmt.Errorf("there are no active music playing")
	case PlaybackPaused:
		if _, err := r.Assistant.Ntg.Resume(r.ChatID); err != nil {
			return err
		}

		r.mu.Lock()
		r.anchor = time.Now()
		r.mu.Unlock()
	}
	return nil
}

// Mute mutes the stream.
func (r *Room) Mute() error {
	s := r.playback()
	if s == PlaybackMuted {
		return nil
	}

	if _, err := r.Assistant.Ntg.Mute(r.ChatID); err != nil {
		return err
	}

	if s == PlaybackPaused {
		r.Resume()
	}
	return nil
}

// Unmute unmutes the stream.
func (r *Room) Unmute() error {
	_, err := r.Assistant.Ntg.Unmute(r.ChatID)
	return err
}

// Replay restarts the current track.
func (r *Room) Replay() error {
	r.mu.Lock()
	if r.track == nil || r.filePath == "" {
		r.mu.Unlock()
		return fmt.Errorf("no track to replay")
	}
	prevOffset, prevAnchor := r.offset, r.anchor
	r.offset, r.anchor = 0, time.Now()
	r.mu.Unlock()

	if err := r.play(); err != nil {
		r.mu.Lock()
		r.offset, r.anchor = prevOffset, prevAnchor
		r.mu.Unlock()
		return err
	}
	return nil
}

// Stop stops the stream and forgets the current track.
func (r *Room) Stop() error {
	err := r.Assistant.Ntg.Stop(r.ChatID)

	r.mu.Lock()
	r.track = nil
	r.filePath = ""
	r.offset = 0
	r.anchor = time.Time{}
	r.mu.Unlock()

	return err
}

// Seek moves the playback position by the given number of seconds.
func (r *Room) Seek(seconds int) error {
	r.mu.Lock()
	if r.track == nil || r.filePath == "" {
		r.mu.Unlock()
		return fmt.Errorf("no track to seek")
	}

	pos := r.position()
	if seconds > 0 && r.track.Duration-pos <= seekEndThreshold {
		r.mu.Unlock()
		return fmt.Errorf("cannot seek, track is about to end")
	}

	newPos := pos + seconds
	if newPos >= r.track.Duration {
		newPos = r.track.Duration - seekSafetyMargin
	}
	newPos = max(newPos, 0)

	prevOffset, prevAnchor := r.offset, r.anchor
	wasMuted := r.playback() == PlaybackMuted
	r.offset, r.anchor = newPos, time.Now()
	r.mu.Unlock()

	if err := r.play(); err != nil {
		r.mu.Lock()
		r.offset, r.anchor = prevOffset, prevAnchor
		r.mu.Unlock()
		return err
	}

	if wasMuted {
		r.Unmute()
	}
	return nil
}

// SetSpeed changes the playback speed, keeping the current position.
func (r *Room) SetSpeed(speed float64) error {
	r.mu.Lock()
	if r.track == nil || r.filePath == "" {
		r.mu.Unlock()
		return fmt.Errorf("no track to adjust speed")
	}

	if speed < minSpeed || speed > maxSpeed {
		r.mu.Unlock()
		return fmt.Errorf(
			"invalid speed: must be between %.2fx and %.1fx",
			minSpeed,
			maxSpeed,
		)
	}

	if r.speed == speed {
		r.mu.Unlock()
		return nil
	}

	prevOffset, prevAnchor, prevSpeed := r.offset, r.anchor, r.speed
	r.offset, r.anchor, r.speed = r.position(), time.Now(), speed
	r.mu.Unlock()

	if err := r.play(); err != nil {
		r.mu.Lock()
		r.offset, r.anchor, r.speed = prevOffset, prevAnchor, prevSpeed
		r.mu.Unlock()
		return err
	}
	return nil
}

// play (re)starts the current track from its offset. It refuses rooms that
// are no longer registered, so a stale room can't rejoin the voice chat.
func (r *Room) play() error {
	if cur, ok := Rooms.Get(r.ChatID); !ok || cur != r {
		return ErrRoomClosed
	}

	r.mu.RLock()
	desc := getMediaDescription(r.filePath, r.offset, r.speed, r.track.Video)
	r.mu.RUnlock()

	if err := r.Assistant.Ntg.Play(r.ChatID, desc); err != nil {
		return err
	}

	paused := r.playback() == PlaybackPaused

	r.mu.Lock()
	r.anchor = time.Time{}
	if !paused {
		r.anchor = time.Now()
	}
	r.mu.Unlock()
	return nil
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
	baseCmd += "-v error -i \"" + url + "\" "

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
	if err := jsonv2.Unmarshal(out, &probe); err != nil {
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

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
	opusSampleRate  = 48000
	audioChannels   = 2
	audioBitrate    = uint32(192000)

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

	var probe mediaProbe
	if isVideo || (speed == 1.0 && !isStreamURL(url)) {
		probe = probeMedia(url)
	}

	audio := getAudioPipeline(baseCmd, url, pos, speed, probe)
	if !isVideo {
		return ntgcalls.MediaDescription{
			Microphone: audio,
		}
	}

	return ntgcalls.MediaDescription{
		Microphone: audio,
		Camera:     getVideoPipeline(baseCmd, speed, probe),
	}
}

// getAudioPipeline builds an Opus passthrough pipeline when possible,
// otherwise falling back to decoded PCM at the requested speed and bitrate.
func getAudioPipeline(
	baseCmd string,
	url string,
	pos int,
	speed float64,
	probe mediaProbe,
) *ntgcalls.AudioDescription {
	if canAudioPassthrough(url, speed, probe) {
		format := ntgcalls.FormatOgg
		channels := uint8(audioChannels)
		if probe.audioChannels == 1 || probe.audioChannels == 2 {
			channels = uint8(probe.audioChannels)
		}
		if !isStreamURL(url) && pos == 0 && hasMediaExt(url, ".ogg", ".opus") {
			return &ntgcalls.AudioDescription{
				MediaSource:  ntgcalls.MediaSourceFile,
				Input:        url,
				SampleRate:   opusSampleRate,
				ChannelCount: channels,
				Format:       &format,
			}
		}
		return &ntgcalls.AudioDescription{
			MediaSource:  ntgcalls.MediaSourceShell,
			Input:        baseCmd + "-vn -sn -dn -c:a copy -f ogg pipe:1",
			SampleRate:   opusSampleRate,
			ChannelCount: channels,
			Format:       &format,
		}
	}

	bitrate := audioBitrate
	format := ntgcalls.FormatPCM
	audio := &ntgcalls.AudioDescription{
		MediaSource:  ntgcalls.MediaSourceShell,
		SampleRate:   audioSampleRate,
		ChannelCount: audioChannels,
		Bitrate:      &bitrate,
		Format:       &format,
	}
	audio.Input = baseCmd +
		"-filter:a \"atempo=" + strconv.FormatFloat(speed, 'f', 2, 64) + "\" " +
		"-f s16le -ac " + strconv.Itoa(audioChannels) + " " +
		"-ar " + strconv.Itoa(audioSampleRate) + " pipe:1"

	return audio
}

// getVideoPipeline builds a Matroska passthrough pipeline for compatible
// codecs/resolutions at normal speed, or falls back to scaled raw I420 video.
func getVideoPipeline(
	baseCmd string,
	speed float64,
	probe mediaProbe,
) *ntgcalls.VideoDescription {
	if canVideoPassthrough(speed, probe) {
		format := ntgcalls.FormatMatroska
		fps := probe.fps
		if fps <= 0 {
			fps = defaultVideoFPS
		}
		return &ntgcalls.VideoDescription{
			MediaSource: ntgcalls.MediaSourceShell,
			Input:       baseCmd + "-an -sn -dn -c:v copy -f matroska pipe:1",
			Width:       int16(probe.width),
			Height:      int16(probe.height),
			FPS:         uint8(fps),
			Format:      &format,
		}
	}

	w, h, fps, filter := normalizeVideo(probe, speed)
	format := ntgcalls.FormatI420
	video := &ntgcalls.VideoDescription{
		MediaSource: ntgcalls.MediaSourceShell,
		Width:       int16(w),
		Height:      int16(h),
		FPS:         uint8(fps),
		Format:      &format,
	}
	video.Input = baseCmd +
		"-filter:v \"" + filter + "\" " +
		"-f rawvideo -r " + strconv.Itoa(fps) + " -pix_fmt yuv420p pipe:1"

	return video
}

// normalizeVideo returns the scaled size, fps and the ffmpeg filter that
// applies the playback speed from a probed video stream.
func normalizeVideo(probe mediaProbe, speed float64) (int, int, int, string) {
	if speed <= 0 {
		speed = 1.0
	}

	w, h := probe.width, probe.height
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

type mediaProbe struct {
	hasAudio      bool
	audioCodec    string
	audioChannels int
	hasVideo      bool
	videoCodec    string
	pixFmt        string
	width         int
	height        int
	fps           int
	hasBFrames    int
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType    string `json:"codec_type"`
		CodecName    string `json:"codec_name"`
		PixFmt       string `json:"pix_fmt"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Channels     int    `json:"channels"`
		HasBFrames   int    `json:"has_b_frames"`
		AvgFrameRate string `json:"avg_frame_rate"`
		RFrameRate   string `json:"r_frame_rate"`
	} `json:"streams"`
}

// probeMedia probes audio and video stream metadata with ffprobe.
func probeMedia(filePath string) mediaProbe {
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
		utils.Log.Error("[probeMedia] ffprobe timed out for " + filePath)
		return mediaProbe{}
	}
	if err != nil {
		utils.Log.Error(
			"[probeMedia] ffprobe failed for " + filePath + " : " + err.Error(),
		)
		return mediaProbe{}
	}

	var probe ffprobeOutput
	if err := jsonv2.Unmarshal(out, &probe); err != nil {
		utils.Log.Error(
			"[probeMedia] failed to parse ffprobe JSON for " + filePath + " : " + err.Error(),
		)
		return mediaProbe{}
	}

	var result mediaProbe
	for _, s := range probe.Streams {
		switch s.CodecType {
		case "audio":
			if !result.hasAudio {
				result.hasAudio = true
				result.audioCodec = strings.ToLower(s.CodecName)
				result.audioChannels = s.Channels
			}
		case "video":
			if !result.hasVideo && s.Width > 0 && s.Height > 0 {
				result.hasVideo = true
				result.videoCodec = strings.ToLower(s.CodecName)
				result.pixFmt = strings.ToLower(s.PixFmt)
				result.width = s.Width
				result.height = s.Height
				result.hasBFrames = s.HasBFrames
				result.fps = parseFrameRate(s.AvgFrameRate)
				if result.fps <= 0 {
					result.fps = parseFrameRate(s.RFrameRate)
				}
			}
		}
	}

	return result
}

func parseFrameRate(rate string) int {
	numStr, denStr, ok := strings.Cut(rate, "/")
	if !ok {
		v, err := strconv.Atoi(rate)
		if err != nil || v <= 0 || v > 60 {
			return 0
		}
		return v
	}
	num, err1 := strconv.ParseFloat(numStr, 64)
	den, err2 := strconv.ParseFloat(denStr, 64)
	if err1 != nil || err2 != nil || den <= 0 || num <= 0 {
		return 0
	}
	fps := int(num/den + 0.5)
	if fps <= 0 || fps > 60 {
		return 0
	}
	return fps
}

func canAudioPassthrough(url string, speed float64, probe mediaProbe) bool {
	if speed != 1.0 {
		return false
	}
	if probe.hasAudio {
		return probe.audioCodec == "opus"
	}
	return isStreamURL(url) && hasMediaExt(url, ".opus")
}

func canVideoPassthrough(speed float64, probe mediaProbe) bool {
	if speed != 1.0 || !probe.hasVideo {
		return false
	}
	if probe.width <= 0 || probe.height <= 0 ||
		probe.width > defaultVideoWidth || probe.height > defaultVideoHeight {
		return false
	}
	if probe.pixFmt != "" && probe.pixFmt != "yuv420p" && probe.pixFmt != "yuvj420p" {
		return false
	}
	switch probe.videoCodec {
	case "vp8", "vp9", "av1":
		return true
	case "h264":
		return probe.hasBFrames == 0
	default:
		return false
	}
}

func hasMediaExt(path string, exts ...string) bool {
	clean := path
	if idx := strings.IndexAny(clean, "?#"); idx >= 0 {
		clean = clean[:idx]
	}
	clean = strings.ToLower(clean)
	for _, ext := range exts {
		if strings.HasSuffix(clean, ext) {
			return true
		}
	}
	return false
}

// isStreamURL reports whether path points to a remote HTTP(S) stream.
func isStreamURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}

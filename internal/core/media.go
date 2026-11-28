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

	audioSampleRate = 48000
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
	ass := r.assistant()
	if ass == nil || ass.Ntg == nil {
		return ErrRoomClosed
	}

	s := r.playback()
	if s == PlaybackPaused {
		return nil
	}

	if _, err := ass.Ntg.Pause(r.ChatID); err != nil {
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
	ass := r.assistant()
	if ass == nil || ass.Ntg == nil {
		return ErrRoomClosed
	}

	switch r.playback() {
	case PlaybackIdle:
		return fmt.Errorf("there are no active music playing")
	case PlaybackPaused:
		if _, err := ass.Ntg.Resume(r.ChatID); err != nil {
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
	ass := r.assistant()
	if ass == nil || ass.Ntg == nil {
		return ErrRoomClosed
	}

	s := r.playback()
	if s == PlaybackMuted {
		return nil
	}

	if _, err := ass.Ntg.Mute(r.ChatID); err != nil {
		return err
	}

	if s == PlaybackPaused {
		r.Resume()
	}
	return nil
}

// Unmute unmutes the stream.
func (r *Room) Unmute() error {
	ass := r.assistant()
	if ass == nil || ass.Ntg == nil {
		return ErrRoomClosed
	}
	_, err := ass.Ntg.Unmute(r.ChatID)
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
	var err error
	if ass := r.assistant(); ass != nil && ass.Ntg != nil {
		err = ass.Ntg.Stop(r.ChatID)
	}

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
	ass := r.Assistant
	desc := getMediaDescription(r.filePath, r.offset, r.speed, r.track.Video)
	r.mu.RUnlock()

	if ass == nil || ass.Ntg == nil {
		return ErrRoomClosed
	}

	if err := ass.Ntg.Play(r.ChatID, desc); err != nil {
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

// getMediaDescription builds an ntgcalls MediaDescription for a local file
// or remote stream, applying start offset, speed, and native passthrough where supported.
func getMediaDescription(
	url string,
	pos int,
	speed float64,
	isVideo bool,
) ntgcalls.MediaDescription {
	speed = min(max(speed, minSpeed), maxSpeed)
	isRemote := isStreamURL(url)

	var mediaInfo *ntgcalls.MediaInfo
	if !isRemote {
		if info, err := ntgcalls.GetMediaInfo(url, ntgcalls.CaptureStream); err == nil {
			mediaInfo = &info
		}
	}

	var remoteVideo *videoStreamInfo
	if isVideo && mediaInfo == nil {
		remoteVideo = probeVideoStream(url)
	}

	baseCmd := buildFFmpegInputCmd(url, pos, isRemote)
	audio := getAudioPipeline(baseCmd, url, pos, speed, isRemote, mediaInfo)

	hasNoVideo := (mediaInfo != nil && mediaInfo.Video == nil) ||
		(remoteVideo != nil && remoteVideo.audioOnly)
	if !isVideo || hasNoVideo {
		return ntgcalls.MediaDescription{
			Microphone: audio,
		}
	}

	return ntgcalls.MediaDescription{
		Microphone: audio,
		Camera:     getVideoPipeline(baseCmd, url, pos, speed, isRemote, mediaInfo, remoteVideo),
	}
}

func buildFFmpegInputCmd(url string, pos int, isRemote bool) string {
	var b strings.Builder
	b.WriteString("ffmpeg ")
	if isRemote {
		b.WriteString("-reconnect 1 -reconnect_streamed 1 -reconnect_delay_max 5 ")
	}
	if pos > 0 {
		b.WriteString("-ss ")
		b.WriteString(strconv.Itoa(pos))
		b.WriteByte(' ')
	}
	b.WriteString("-v error -i ")
	b.WriteString(shellQuote(url))
	b.WriteByte(' ')
	return b.String()
}

// getAudioPipeline uses ntgcalls MediaSourceFile + FormatAuto for supported
// local files at normal speed from position 0, Ogg/Opus passthrough when seeked
// at 1.0x speed, or decoded 48 kHz PCM otherwise.
func getAudioPipeline(
	baseCmd string,
	url string,
	pos int,
	speed float64,
	isRemote bool,
	mediaInfo *ntgcalls.MediaInfo,
) *ntgcalls.AudioDescription {
	bitrate := audioBitrate

	if !isRemote && mediaInfo != nil && mediaInfo.Audio != nil && mediaInfo.Audio.Supported {
		channels := uint8(audioChannels)
		if mediaInfo.Audio.Channels == 1 {
			channels = 1
		}
		if pos == 0 && speed == 1.0 {
			format := ntgcalls.FormatAuto
			return &ntgcalls.AudioDescription{
				MediaSource:  ntgcalls.MediaSourceFile,
				Input:        url,
				SampleRate:   audioSampleRate,
				ChannelCount: channels,
				Bitrate:      &bitrate,
				Format:       &format,
			}
		}
		if speed == 1.0 && mediaInfo.Audio.Passthrough {
			format := ntgcalls.FormatOgg
			return &ntgcalls.AudioDescription{
				MediaSource:  ntgcalls.MediaSourceShell,
				Input:        baseCmd + "-vn -sn -dn -c:a copy -f ogg pipe:1",
				SampleRate:   audioSampleRate,
				ChannelCount: channels,
				Format:       &format,
			}
		}
	}

	if isRemote && speed == 1.0 && hasMediaExt(url, ".opus") {
		format := ntgcalls.FormatOgg
		return &ntgcalls.AudioDescription{
			MediaSource:  ntgcalls.MediaSourceShell,
			Input:        baseCmd + "-vn -sn -dn -c:a copy -f ogg pipe:1",
			SampleRate:   audioSampleRate,
			ChannelCount: audioChannels,
			Format:       &format,
		}
	}

	format := ntgcalls.FormatPCM
	var filterArg string
	if speed != 1.0 {
		filterArg = "-filter:a \"atempo=" + strconv.FormatFloat(speed, 'f', 2, 64) + "\" "
	}
	return &ntgcalls.AudioDescription{
		MediaSource: ntgcalls.MediaSourceShell,
		Input: fmt.Sprintf(
			"%s-vn -sn -dn %s-f s16le -ac %d -ar %d pipe:1",
			baseCmd, filterArg, audioChannels, audioSampleRate,
		),
		SampleRate:   audioSampleRate,
		ChannelCount: audioChannels,
		Bitrate:      &bitrate,
		Format:       &format,
	}
}

// getVideoPipeline uses ntgcalls MediaSourceFile + FormatAuto for supported
// local files at normal speed from position 0, Matroska passthrough when
// forwardable at 1.0x speed, or scaled raw I420 video otherwise.
func getVideoPipeline(
	baseCmd string,
	url string,
	pos int,
	speed float64,
	isRemote bool,
	mediaInfo *ntgcalls.MediaInfo,
	remoteVideo *videoStreamInfo,
) *ntgcalls.VideoDescription {
	if !isRemote && mediaInfo != nil && mediaInfo.Video != nil {
		v := mediaInfo.Video
		w, h := clampVideoDimensions(int(v.Width), int(v.Height))
		fps := clampVideoFPS(v.FrameRate)

		if pos == 0 && speed == 1.0 && v.Supported {
			format := ntgcalls.FormatAuto
			return &ntgcalls.VideoDescription{
				MediaSource: ntgcalls.MediaSourceFile,
				Input:       url,
				Width:       int16(w),
				Height:      int16(h),
				FPS:         uint8(fps),
				Format:      &format,
			}
		}
		if speed == 1.0 && v.Passthrough &&
			v.Width > 0 && v.Height > 0 &&
			int(v.Width) <= defaultVideoWidth && int(v.Height) <= defaultVideoHeight {
			format := ntgcalls.FormatMatroska
			return &ntgcalls.VideoDescription{
				MediaSource: ntgcalls.MediaSourceShell,
				Input:       baseCmd + "-an -sn -dn -c:v copy -f matroska pipe:1",
				Width:       int16(v.Width),
				Height:      int16(v.Height),
				FPS:         uint8(fps),
				Format:      &format,
			}
		}
		return buildRawVideoDescription(baseCmd, w, h, defaultVideoFPS, speed)
	}

	if remoteVideo != nil && speed == 1.0 && remoteVideo.canPassthrough() {
		format := ntgcalls.FormatMatroska
		fps := remoteVideo.fps
		if fps <= 0 {
			fps = defaultVideoFPS
		}
		return &ntgcalls.VideoDescription{
			MediaSource: ntgcalls.MediaSourceShell,
			Input:       baseCmd + "-an -sn -dn -c:v copy -f matroska pipe:1",
			Width:       int16(remoteVideo.width),
			Height:      int16(remoteVideo.height),
			FPS:         uint8(fps),
			Format:      &format,
		}
	}

	var srcW, srcH int
	if remoteVideo != nil {
		srcW, srcH = remoteVideo.width, remoteVideo.height
	}
	w, h := clampVideoDimensions(srcW, srcH)
	return buildRawVideoDescription(baseCmd, w, h, defaultVideoFPS, speed)
}

func buildRawVideoDescription(baseCmd string, w, h, fps int, speed float64) *ntgcalls.VideoDescription {
	filter := fmt.Sprintf("scale=%d:%d", w, h)
	if speed != 1.0 {
		filter = fmt.Sprintf("setpts=%.4f*PTS,%s", 1.0/speed, filter)
	}
	format := ntgcalls.FormatI420
	return &ntgcalls.VideoDescription{
		MediaSource: ntgcalls.MediaSourceShell,
		Input: fmt.Sprintf(
			"%s-an -sn -dn -filter:v %q -f rawvideo -r %d -pix_fmt yuv420p pipe:1",
			baseCmd, filter, fps,
		),
		Width:  int16(w),
		Height: int16(h),
		FPS:    uint8(fps),
		Format: &format,
	}
}

func clampVideoDimensions(w, h int) (int, int) {
	if w <= 0 || h <= 0 {
		return defaultVideoWidth, defaultVideoHeight
	}
	if w > defaultVideoWidth {
		h = h * defaultVideoWidth / w
		w = defaultVideoWidth
	}
	if h > defaultVideoHeight {
		w = w * defaultVideoHeight / h
		h = defaultVideoHeight
	}
	w &^= 1
	h &^= 1
	if w <= 0 || h <= 0 {
		return defaultVideoWidth, defaultVideoHeight
	}
	return w, h
}

func clampVideoFPS(fps float64) int {
	if fps <= 0 {
		return defaultVideoFPS
	}
	rounded := int(fps + 0.5)
	if rounded <= 0 || rounded > defaultVideoFPS {
		return defaultVideoFPS
	}
	return rounded
}

type videoStreamInfo struct {
	audioOnly  bool
	codecName  string
	pixFmt     string
	width      int
	height     int
	fps        int
	hasBFrames int
}

func (v *videoStreamInfo) canPassthrough() bool {
	if v.width <= 0 || v.height <= 0 ||
		v.width > defaultVideoWidth || v.height > defaultVideoHeight {
		return false
	}
	if v.pixFmt != "" && v.pixFmt != "yuv420p" && v.pixFmt != "yuvj420p" {
		return false
	}
	switch v.codecName {
	case "vp8", "vp9", "av1":
		return true
	case "h264":
		return v.hasBFrames == 0
	default:
		return false
	}
}

type ffprobeStreamOutput struct {
	Streams []struct {
		CodecType    string `json:"codec_type"`
		CodecName    string `json:"codec_name"`
		PixFmt       string `json:"pix_fmt"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		HasBFrames   int    `json:"has_b_frames"`
		AvgFrameRate string `json:"avg_frame_rate"`
		RFrameRate   string `json:"r_frame_rate"`
		Disposition  struct {
			AttachedPic int `json:"attached_pic"`
		} `json:"disposition"`
	} `json:"streams"`
}

// probeVideoStream probes stream metadata via ffprobe for remote HTTP(S) video URLs
// or local inputs that ntgcalls.GetMediaInfo could not inspect.
func probeVideoStream(filePath string) *videoStreamInfo {
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
		utils.Log.Error("[probeVideoStream] ffprobe timed out for " + filePath)
		return nil
	}
	if err != nil {
		utils.Log.Error("[probeVideoStream] ffprobe failed for " + filePath + " : " + err.Error())
		return nil
	}

	var probe ffprobeStreamOutput
	if err := jsonv2.Unmarshal(out, &probe); err != nil {
		utils.Log.Error("[probeVideoStream] failed to parse ffprobe JSON for " + filePath + " : " + err.Error())
		return nil
	}

	hasAudio := false
	for _, s := range probe.Streams {
		switch s.CodecType {
		case "audio":
			hasAudio = true
		case "video":
			if s.Disposition.AttachedPic == 0 && s.Width > 0 && s.Height > 0 {
				fps := parseFrameRate(s.AvgFrameRate)
				if fps <= 0 {
					fps = parseFrameRate(s.RFrameRate)
				}
				return &videoStreamInfo{
					codecName:  strings.ToLower(s.CodecName),
					pixFmt:     strings.ToLower(s.PixFmt),
					width:      s.Width,
					height:     s.Height,
					fps:        fps,
					hasBFrames: s.HasBFrames,
				}
			}
		}
	}

	return &videoStreamInfo{audioOnly: hasAudio}
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

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

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
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/amarnathcjd/gortc/media"
)

type speedSource struct {
	path  string
	video bool
	speed float64
}

func newSpeedSource(path string, video bool, speed float64) *speedSource {
	if speed <= 0 {
		speed = 1.0
	}
	return &speedSource{path: path, video: video, speed: speed}
}

func (s *speedSource) Tracks() media.Track {
	if s.video {
		return media.TrackAudio | media.TrackVideo
	}
	return media.TrackAudio
}

func (s *speedSource) Open(ctx context.Context) (*media.Streams, error) {
	return s.open(ctx, 0)
}

func (s *speedSource) OpenAt(ctx context.Context, offset time.Duration) (*media.Streams, error) {
	return s.open(ctx, offset)
}

func (s *speedSource) open(ctx context.Context, offset time.Duration) (*media.Streams, error) {
	st := &media.Streams{}
	var procs []*exec.Cmd

	seekArgs := []string{}
	if offset > 0 {
		seekArgs = []string{"-ss", strconv.FormatFloat(offset.Seconds(), 'f', 3, 64)}
	}

	startAudio := func() (io.Reader, error) {
		args := []string{"-hide_banner", "-loglevel", "error"}
		args = append(args, seekArgs...)
		args = append(args, "-i", s.path, "-vn")
		args = append(args, "-filter:a", atempoFilter(s.speed))
		args = append(args,
			"-c:a", "libopus",
			"-b:a", "128k",
			"-vbr", "on",
			"-compression_level", "10",
			"-frame_duration", "20",
			"-page_duration", "20000",
			"-application", "audio",
			"-mapping_family", "0",
			"-ac", "2",
			"-ar", "48000",
			"-f", "ogg",
			"pipe:1",
		)
		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		out, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		procs = append(procs, cmd)
		return out, nil
	}

	startVideo := func() (io.Reader, error) {
		args := []string{"-hide_banner", "-loglevel", "error", "-re"}
		args = append(args, seekArgs...)
		args = append(args, "-i", s.path, "-an")
		args = append(args, "-vf", s.buildVideoFilter())
		args = append(args,
			"-c:v", "libvpx",
			"-b:v", "1500k", "-minrate", "1500k", "-maxrate", "1500k", "-bufsize", "1500k",
			"-rc_lookahead", "16",
			"-lag-in-frames", "16",
			"-r", "30",
			"-g", "60", "-keyint_min", "60",
			"-auto-alt-ref", "0",
			"-error-resilient", "1",
			"-deadline", "realtime",
			"-cpu-used", "4",
			"-threads", "4",
			"-f", "ivf",
			"pipe:1",
		)
		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		out, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		procs = append(procs, cmd)
		return out, nil
	}

	audio, err := startAudio()
	if err != nil {
		killAll(procs)
		return nil, fmt.Errorf("start audio ffmpeg: %w", err)
	}
	st.Audio = audio

	if s.video {
		video, err := startVideo()
		if err != nil {
			killAll(procs)
			return nil, fmt.Errorf("start video ffmpeg: %w", err)
		}
		st.Video = video
	}

	return st, nil
}

func killAll(procs []*exec.Cmd) {
	for _, c := range procs {
		if c.Process != nil {
			_ = c.Process.Kill()
		}
	}
}

func (s *speedSource) buildVideoFilter() string {
	scale := "scale=1280:720:force_original_aspect_ratio=decrease:force_divisible_by=2"
	if s.speed != 1.0 {
		return fmt.Sprintf("setpts=PTS/%f,%s", s.speed, scale)
	}
	return scale
}

func atempoFilter(speed float64) string {
	var stages []string
	remaining := speed
	for remaining > 2.0 {
		stages = append(stages, "atempo=2.0")
		remaining /= 2.0
	}
	for remaining < 0.5 {
		stages = append(stages, "atempo=0.5")
		remaining /= 0.5
	}
	stages = append(stages, fmt.Sprintf("atempo=%.3f", remaining))

	out := stages[0]
	for _, st := range stages[1:] {
		out += "," + st
	}
	return out
}

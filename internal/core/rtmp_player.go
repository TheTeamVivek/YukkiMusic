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
	"os/exec"
	"strconv"
	"sync"
)

var onStreamEnd func(chatID int64)

func SetOnStreamEnd(f func(chatID int64)) {
	onStreamEnd = f
}

type RTMPPlayer struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

func (p *RTMPPlayer) killLocked() {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	p.cmd = nil
}

func (p *RTMPPlayer) kill() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.killLocked()
}

func (p *RTMPPlayer) Play(r *RoomState) error {
	if r.FilePath == "" {
		return Err("no file")
	}
	if r.rtmpURL == "" || r.rtmpKey == "" {
		return Err("missing rtmp config")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.killLocked()

	speed := r.Speed
	if speed < 0.5 {
		speed = 0.5
	} else if speed > 4.0 {
		speed = 4.0
	}

	seek := r.Position
	args := []string{"-re"}

	if seek > 0 {
		args = append(args, "-ss", strconv.Itoa(seek))
	}

	args = append(args,
		"-v", "warning",
		"-i", r.FilePath,
	)

	audioFilter := "atempo=" + strconv.FormatFloat(speed, 'f', 2, 64)

	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
	)
	if speed != 1.0 {
		args = append(args, "-filter:a", audioFilter)
	}

	if r.Track != nil && r.Track.Video {
		w, h, fps, filter := normalizeVideo(r.FilePath, speed)

		args = append(args,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-r", strconv.Itoa(fps),
			"-filter:v", filter,
		)
	} else {
		args = append(args, "-vn")
	}

	outputURL := r.rtmpURL + "/" + r.rtmpKey

	args = append(args,
		"-f", "flv",
		outputURL,
	)

	cmd := exec.Command("ffmpeg", args...)
	p.cmd = cmd

	if err := cmd.Start(); err != nil {
		p.cmd = nil
		return err
	}

	go func(chatID int64, c *exec.Cmd) {
		_ = c.Wait()

		p.mu.Lock()
		defer p.mu.Unlock()

		if p.cmd == c {
			p.cmd = nil
			if onStreamEnd != nil {
				onStreamEnd(chatID)
			}
		}
	}(r.ChatID, cmd)

	return nil
}

func (p *RTMPPlayer) Pause(r *RoomState) (bool, error) {
	p.kill()
	return true, nil
}

func (p *RTMPPlayer) Resume(r *RoomState) (bool, error) {
	if err := p.Play(r); err != nil {
		return false, err
	}
	return true, nil
}

func (p *RTMPPlayer) Stop(r *RoomState) error {
	p.kill()
	return nil
}

func (p *RTMPPlayer) Mute(r *RoomState) (bool, error) {
	p.kill()
	return true, nil
}

func (p *RTMPPlayer) Unmute(r *RoomState) (bool, error) {
	if err := p.Play(r); err != nil {
		return false, err
	}
	return true, nil
}

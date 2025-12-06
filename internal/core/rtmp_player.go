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
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
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
		gologging.Debug("killLocked: killing process pid=" + strconv.Itoa(p.cmd.Process.Pid))
		_ = p.cmd.Process.Kill()
	} else {
		gologging.Debug("killLocked: no running process")
	}
	p.cmd = nil
}

func (p *RTMPPlayer) kill() {
	p.mu.Lock()
	defer p.mu.Unlock()
	gologging.Debug("kill called")
	p.killLocked()
}

func (p *RTMPPlayer) Play(r *RoomState) error {
	if r.fpath == "" {
		gologging.Error("no file, chatID=" + strconv.FormatInt(r.chatID, 10))
		return errors.New("no file")
	}
	if r.rtmpURL == "" || r.rtmpKey == "" {
		gologging.Error("missing rtmp config, chatID=" + strconv.FormatInt(r.chatID, 10))
		return errors.New("missing rtmp config")
	}

	p.mu.Lock()
	gologging.Debug("killing existing ffmpeg, chatID=" + strconv.FormatInt(r.chatID, 10))
	p.killLocked()
	p.mu.Unlock()

	speed := r.speed
	if speed < 0.5 {
		gologging.Warn("speed below min, requested=" +
			strconv.FormatFloat(r.speed, 'f', 2, 64) +
			" using=0.50 chatID=" + strconv.FormatInt(r.chatID, 10))
		speed = 0.5
	} else if speed > 4.0 {
		gologging.Warn("speed above max, requested=" +
			strconv.FormatFloat(r.speed, 'f', 2, 64) +
			" using=4.00 chatID=" + strconv.FormatInt(r.chatID, 10))
		speed = 4.0
	}

	seek := r.position
	args := []string{"-re"}
	if seek > 0 {
		args = append(args, "-ss", strconv.Itoa(seek))
	}

	args = append(args, "-v", "warning", "-i", r.fpath)
	audioFilter := buildAudioFilter(speed)

	if r.track != nil && r.track.Video {
		_, _, fps, vfilter := normalizeVideo(r.fpath, speed)
		gologging.Info("stream with video, fps=" + strconv.Itoa(fps) +
			" filter=" + vfilter +
			" chatID=" + strconv.FormatInt(r.chatID, 10))
		args = append(args,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-r", strconv.Itoa(fps),
			"-filter:v", vfilter,
		)
	} else if r.track != nil && r.track.Artwork != "" {
		gologging.Info("audio with artwork, chatID=" + strconv.FormatInt(r.chatID, 10) +
			" artwork=" + r.track.Artwork)
		args = append(args,
			"-loop", "1",
			"-i", downloadThumb(r.track.ID, r.track.Artwork),
		)
		args = append(args,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-shortest",
			"-map", "0:a",
			"-map", "1:v",
		)
	} else {
		gologging.Info("audio only (no artwork), chatID=" + strconv.FormatInt(r.chatID, 10))
		args = append(args, "-vn")
	}

	args = append(args, "-c:a", "aac", "-b:a", "128k")
	if audioFilter != "" {
		args = append(args, "-filter:a", audioFilter)
	}

	outputURL := r.rtmpURL + "/" + r.rtmpKey
	args = append(args, "-f", "flv", outputURL)

	gologging.Debug("ffmpeg args: " + strings.Join(args, " ") +
		" chatID=" + strconv.FormatInt(r.chatID, 10))

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	p.mu.Lock()
	p.cmd = cmd
	p.mu.Unlock()

	gologging.Info("starting ffmpeg, outputURL=" + outputURL +
		" chatID=" + strconv.FormatInt(r.chatID, 10))

	if err := cmd.Start(); err != nil {
		gologging.Error("ffmpeg start error: " + err.Error() +
			" chatID=" + strconv.FormatInt(r.chatID, 10))
		p.mu.Lock()
		if p.cmd == cmd {
			p.cmd = nil
		}
		p.mu.Unlock()
		return err
	}

	done := make(chan error, 1)

	go func(chatID int64, c *exec.Cmd, done chan<- error) {
		err := c.Wait()
		done <- err

		p.mu.Lock()
		defer p.mu.Unlock()

		if err != nil {
			gologging.Error("ffmpeg exited with error: " + err.Error() +
				" chatID=" + strconv.FormatInt(chatID, 10))
		} else {
			gologging.Info("ffmpeg exited normally, chatID=" +
				strconv.FormatInt(chatID, 10))
		}

		if p.cmd == c {
			p.cmd = nil
			gologging.Debug("onStreamEnd callback, chatID=" +
				strconv.FormatInt(chatID, 10))
			if onStreamEnd != nil {
				onStreamEnd(chatID)
			}
		}
	}(r.chatID, cmd, done)

	select {
	case err := <-done:
		gologging.Error("ffmpeg exited quickly, chatID=" +
			strconv.FormatInt(r.chatID, 10) +
			" err=" + err.Error())
		return err
	case <-time.After(3500 * time.Millisecond):
		gologging.Debug("ffmpeg running fine after 5s, chatID=" +
			strconv.FormatInt(r.chatID, 10))
		return nil
	}
}

func (p *RTMPPlayer) Pause(r *RoomState) (bool, error) {
	gologging.Info("pause requested, chatID=" + strconv.FormatInt(r.chatID, 10))
	p.kill()
	return true, nil
}

func (p *RTMPPlayer) Resume(r *RoomState) (bool, error) {
	gologging.Info("resume requested, chatID=" + strconv.FormatInt(r.chatID, 10))
	if err := p.Play(r); err != nil {
		gologging.Error("resume failed, chatID=" + strconv.FormatInt(r.chatID, 10) +
			" error=" + err.Error())
		return false, err
	}
	gologging.Info("resume started, chatID=" + strconv.FormatInt(r.chatID, 10))
	return true, nil
}

func (p *RTMPPlayer) Stop(r *RoomState) error {
	gologging.Info("stop requested, chatID=" + strconv.FormatInt(r.chatID, 10))
	p.kill()
	return nil
}

func (p *RTMPPlayer) Mute(r *RoomState) (bool, error) {
	gologging.Info("mute requested, chatID=" + strconv.FormatInt(r.chatID, 10))
	p.kill()
	return true, nil
}

func (p *RTMPPlayer) Unmute(r *RoomState) (bool, error) {
	gologging.Info("unmute requested, chatID=" + strconv.FormatInt(r.chatID, 10))
	if err := p.Play(r); err != nil {
		gologging.Error("unmute failed, chatID=" + strconv.FormatInt(r.chatID, 10) +
			" error=" + err.Error())
		return false, err
	}
	gologging.Info("unmute started, chatID=" + strconv.FormatInt(r.chatID, 10))
	return true, nil
}

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

package utils

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"
)

func GetDurationByFFProbe(filePath string) (int, error) {
	if filePath == "" {
		return 0, errors.New("empty file path")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	result := strings.TrimSpace(out.String())
	seconds, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return 0, err
	}

	return int(seconds), nil
}

func GetDuration(m *td.Message) int {
	if m == nil || m.Content == nil {
		return 0
	}
	switch c := m.Content.(type) {
	case *td.MessageAudio:
		if c.Audio != nil {
			return int(c.Audio.Duration)
		}
	case *td.MessageVideo:
		if c.Video != nil {
			return int(c.Video.Duration)
		}
	case *td.MessageVoiceNote:
		if c.VoiceNote != nil {
			return int(c.VoiceNote.Duration)
		}
	case *td.MessageVideoNote:
		if c.VideoNote != nil {
			return int(c.VideoNote.Duration)
		}
	case *td.MessageDocument:
		if c.Document != nil {
			if c.Document.FileName == "" {
				return 0
			}
			dur, err := GetDurationByFFProbe(c.Document.FileName)
			if err != nil {
				Log.Errorf(
					"GetDurationByFFProbe failed for %q: %v",
					c.Document.FileName, err,
				)
				return 0
			}
			return dur
		}
	}
	return 0
}

var progressBarCache = [9]string{
	"◉———————",
	"—◉——————",
	"——◉—————",
	"———◉————",
	"————◉———",
	"—————◉——",
	"——————◉—",
	"———————◉",
	"————————◉",
}

func GetProgressBar(playedSec, durationSec int) string {
	if durationSec <= 0 || playedSec <= 0 {
		return progressBarCache[0]
	}

	if playedSec >= durationSec {
		return progressBarCache[8]
	}

	index := min((playedSec*9)/durationSec, 8)

	return progressBarCache[index]
}

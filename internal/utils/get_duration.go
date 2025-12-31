/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package utils

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetDurationByFFProbe(filePath string) (int, error) {
	cmd := exec.Command(
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

func GetDuration(f *telegram.MessageMediaDocument) int {
	if f.Document == nil {
		return 0
	}
	d, ok := f.Document.(*telegram.DocumentObj)

	if !ok {
		return 0
	}

	for _, attr := range d.Attributes {
		switch a := attr.(type) {
		case *telegram.DocumentAttributeAudio:
			return int(a.Duration)
		case *telegram.DocumentAttributeVideo:
			return int(a.Duration)
		}
	}

	return 0
}

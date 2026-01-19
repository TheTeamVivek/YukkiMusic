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
package platforms

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Laky-64/gologging"

	state "main/internal/core/models"
)

func getPath(track *state.Track, ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	mediaType := "audio"
	if track.Video {
		mediaType = "video"
	}

	filename := mediaType + "_" + track.ID + ext

	return filepath.Join("downloads", filename)
}

func fileExists(path string) bool {
	i, err := os.Stat(path)
	if err != nil {
		gologging.ErrorF("os.Stat: %v", err)
		return false
	}

	return i.Size() > 0
}

func findFile(track *state.Track) string {
	t := "audio"
	if track.Video {
		t = "video"
	}

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		gologging.ErrorF("filepath.Glob: %v", err)
		return ""
	}

	for _, f := range files {
		if i, err := os.Stat(f); err == nil && i.Size() > 0 {
			return f
		}
	}

	return ""
}

func findAndRemove(track *state.Track) {
	t := "audio"
	if track.Video {
		t = "video"
	}

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		return
	}

	for _, f := range files {
		os.Remove(f)
	}
}

func sanitizeAPIError(err error, apiKey string) error {
	if err == nil || apiKey == "" {
		return err
	}
	masked := strings.ReplaceAll(err.Error(), apiKey, "***REDACTED***")
	return errors.New(masked)
}

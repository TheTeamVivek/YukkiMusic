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
)

// checkDownloadedFile checks if a file already exists in downloads folder
func checkDownloadedFile(trackID string) (string, error) {
	pattern := filepath.Join("./downloads", trackID+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", errors.New("file not found")
	}
	return matches[0], nil
}

// EnsureDownloadsDir creates the downloads directory if it doesn't exist
func ensureDownloadsDir() error {
	return os.MkdirAll("downloads", os.ModePerm)
}

func sanitizeAPIError(err error, apiKey string) error {
	if err == nil || apiKey == "" {
		return err
	}
	masked := strings.ReplaceAll(err.Error(), apiKey, "***REDACTED***")
	return errors.New(masked)
}

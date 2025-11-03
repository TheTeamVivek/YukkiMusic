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

package cookies

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"

	"main/config"
)

var (
	cachedFiles []string
	cacheOnce   sync.Once
	logger      = gologging.GetLogger("cookies")
)

func Init() {
	if err := os.MkdirAll("internal/cookies", 0o700); err != nil {
		logger.Fatal("Failed to create cookies directory:", err)
	}

	urls := strings.Fields(config.CookiesLink)
	for _, url := range urls {
		if err := downloadCookieFile(url); err != nil {
			logger.WarnF("Failed to download cookie file from %s: %v", url, err)
		}
	}
}

func downloadCookieFile(url string) error {
	id := filepath.Base(url)
	rawURL := "https://batbin.me/raw/" + id

	resp, err := http.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, rawURL)
	}
	filePath := filepath.Join("internal/cookies", id+".txt")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func GetRandomCookieFile() (string, error) {
	var err error

	cacheOnce.Do(func() {
		err = loadCookieCache()
	})

	if err != nil {
		logger.WarnF("Failed to load cookie cache: %v — retrying next call", err)
		cacheOnce = sync.Once{} // allow retry next time
		return "", err
	}

	if len(cachedFiles) == 0 {
		logger.Warn("No cookie files found in cache")
		return "", nil
	}

	return cachedFiles[rand.Intn(len(cachedFiles))], nil
}

func loadCookieCache() error {
	files, err := filepath.Glob("internal/cookies/*.txt")
	if err != nil {
		return err
	}
	cachedFiles = files
	return nil
}

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

package cookies

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"yukkimusic/internal/logger"

	"resty.dev/v3"

	"yukkimusic/config"
)

const cookieDir = "config/cookies"

var (
	cachedFiles []string
	cacheOnce   sync.Once
	client      = resty.New().SetTimeout(30 * time.Second)
)

func init() {
	logger.Debug("🔹 Initializing cookies...")

	urls := strings.FieldsSeq(config.CookiesLink)
	for url := range urls {
		if err := downloadCookieFile(url); err != nil {
			logger.Warnf(
				"Failed to download cookie file from %s: %v",
				url,
				err,
			)
		}
	}
}

func downloadCookieFile(url string) error {
	id := filepath.Base(url)
	rawURL := "https://batbin.me/raw/" + id
	filePath := filepath.Join(cookieDir, id+".txt")

	resp, err := client.R().
		SetResponseSaveFileName(filePath).
		Get(rawURL)
	if err != nil {
		return err
	}

	if resp.IsStatusFailure() {
		return fmt.Errorf(
			"unexpected status %d from %s",
			resp.StatusCode(),
			rawURL,
		)
	}

	return nil
}

func loadCookieCache() error {
	files, err := filepath.Glob(filepath.Join(cookieDir, "*.txt"))
	if err != nil {
		return err
	}

	var filtered []string

	for _, f := range files {
		if filepath.Base(f) == "example.txt" {
			continue
		}
		filtered = append(filtered, f)
	}

	cachedFiles = filtered
	return nil
}

func GetRandomCookieFile() (string, error) {
	var err error

	cacheOnce.Do(func() {
		err = loadCookieCache()
	})

	if err != nil {
		logger.Warnf("Failed to load cookie cache: %v", err)
		return "", err
	}

	if len(cachedFiles) == 0 {
		logger.Warn("No cookie files available")
		return "", nil
	}

	return cachedFiles[rand.Intn(len(cachedFiles))], nil
}

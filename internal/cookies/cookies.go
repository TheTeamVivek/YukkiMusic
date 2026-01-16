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
	"embed"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	"resty.dev/v3"

	"main/internal/config"
)

var (
	cachedFiles []string
	cacheOnce   sync.Once
)

//go:embed *.txt
var embeddedCookies embed.FS

func init() {
	gologging.Debug("🔹 Initializing cookies...")

	if err := copyEmbeddedCookies(); err != nil {
		gologging.Fatal("Failed to copy embedded cookies:", err)
	}

	urls := strings.Fields(config.CookiesLink)
	for _, url := range urls {
		if err := downloadCookieFile(url); err != nil {
			gologging.WarnF(
				"Failed to download cookie file from %s: %v",
				url,
				err,
			)
		}
	}
}

func copyEmbeddedCookies() error {
	entries, err := embeddedCookies.ReadDir(".")
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == "example.txt" {
			continue
		}

		dst := filepath.Join("internal/cookies", e.Name())

		if _, err := os.Stat(dst); err == nil {
			continue
		}

		data, err := embeddedCookies.ReadFile(e.Name())
		if err != nil {
			return err
		}

		if err := os.WriteFile(dst, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func downloadCookieFile(url string) error {
	id := filepath.Base(url)
	rawURL := "https://batbin.me/raw/" + id
	filePath := filepath.Join("internal/cookies", id+".txt")

	client := resty.New()
	defer client.Close()

	resp, err := client.R().
		SetOutputFileName(filePath).
		Get(rawURL)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf(
			"unexpected status %d from %s",
			resp.StatusCode(),
			rawURL,
		)
	}

	return nil
}

func loadCookieCache() error {
	files, err := filepath.Glob("internal/cookies/*.txt")
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
		gologging.WarnF("Failed to load cookie cache: %v", err)
		cacheOnce = sync.Once{}
		return "", err
	}

	if len(cachedFiles) == 0 {
		gologging.Warn("No cookie files available")
		return "", nil
	}

	return cachedFiles[rand.Intn(len(cachedFiles))], nil
}

/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * JioSaavn API logic and decryption methods are based on:
 * https://github.com/sumitkolhe/jiosaavn-api/
 * MIT License - Copyright (c) 2024 Sumit Kolhe
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
package platforms

import (
	"context"
	"crypto/des"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"resty.dev/v3"

	state "main/internal/core/models"
	"main/internal/utils"
)

const PlatformJioSaavn state.PlatformName = "JioSaavn"

type JioSaavnPlatform struct {
	name   state.PlatformName
	client *resty.Client
}

var (
	jiosaavnLinkRegex     = regexp.MustCompile(`(?i)(jiosaavn\.com|saavn\.com)`)
	jiosaavnSongRegex     = regexp.MustCompile(`(?i)jiosaavn\.com/song/[^/]+/([^/]+)`)
	jiosaavnAlbumRegex    = regexp.MustCompile(`(?i)jiosaavn\.com/album/[^/]+/([^/]+)`)
	jiosaavnPlaylistRegex = regexp.MustCompile(`(?i)jiosaavn\.com/(?:featured|s/playlist)/[^/]+/([^/]+)`)
	jiosaavnCache         = utils.NewCache[string, []*state.Track](2 * time.Hour)

	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:136.0) Gecko/20100101 Firefox/136.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.8.4 Chrome/130.0.6723.191 Electron/33.3.2 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:136.0) Gecko/20100101 Firefox/136.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 OPR/117.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.8.3 Chrome/130.0.6723.191 Electron/33.3.2 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 YaBrowser/25.2.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/27.0 Chrome/125.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.8.9 Chrome/132.0.6834.210 Electron/34.3.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.58 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.67 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/134.0.6998.99 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.7.7 Chrome/128.0.6613.186 Electron/32.2.5 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.132 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
		"Mozilla/5.0 (Linux; Android 13; SM-G981B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.8.9 Chrome/132.0.6834.210 Electron/34.3.2 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.8.9 Chrome/132.0.6834.210 Electron/34.3.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.6.5 Chrome/124.0.6367.243 Electron/30.1.2 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.92 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.79 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) obsidian/1.8.4 Chrome/130.0.6723.191 Electron/33.3.2 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	}
)

type JioSaavnSongResponse struct {
	Songs []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Album    string `json:"album"`
		Year     string `json:"year"`
		Duration string `json:"duration"`
		Language string `json:"language"`
		Image    string `json:"image"`
		PermaURL string `json:"perma_url"`
		MoreInfo struct {
			EncryptedMediaURL string `json:"encrypted_media_url"`
			Album             string `json:"album"`
			AlbumID           string `json:"album_id"`
			Duration          string `json:"duration"`
		} `json:"more_info"`
	} `json:"songs"`
}

type JioSaavnAlbumResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Year     string `json:"year"`
	Language string `json:"language"`
	Image    string `json:"image"`
	PermaURL string `json:"perma_url"`
	List     []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Duration string `json:"duration"`
		Image    string `json:"image"`
		PermaURL string `json:"perma_url"`
		MoreInfo struct {
			EncryptedMediaURL string `json:"encrypted_media_url"`
		} `json:"more_info"`
	} `json:"list"`
}

type JioSaavnPlaylistResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Language string `json:"language"`
	Image    string `json:"image"`
	PermaURL string `json:"perma_url"`
	List     []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Duration string `json:"duration"`
		Image    string `json:"image"`
		PermaURL string `json:"perma_url"`
		MoreInfo struct {
			EncryptedMediaURL string `json:"encrypted_media_url"`
		} `json:"more_info"`
	} `json:"list"`
}

func init() {
	Register(88, &JioSaavnPlatform{
		name: PlatformJioSaavn,
	})
}

func (j *JioSaavnPlatform) Name() state.PlatformName {
	return j.name
}

func (j *JioSaavnPlatform) IsValid(query string) bool {
	query = strings.TrimSpace(query)
	return jiosaavnLinkRegex.MatchString(query)
}

func (j *JioSaavnPlatform) GetTracks(query string, _ bool) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	cacheKey := "jiosaavn:" + strings.ToLower(query)
	if cached, ok := jiosaavnCache.Get(cacheKey); ok {
		gologging.Debug("JioSaavn: Using cached tracks")
		return cached, nil
	}

	var tracks []*state.Track
	var err error

	if matches := jiosaavnSongRegex.FindStringSubmatch(query); len(matches) > 1 {
		tracks, err = j.fetchSong(matches[1])
	} else if matches := jiosaavnAlbumRegex.FindStringSubmatch(query); len(matches) > 1 {
		tracks, err = j.fetchAlbum(matches[1])
	} else if matches := jiosaavnPlaylistRegex.FindStringSubmatch(query); len(matches) > 1 {
		tracks, err = j.fetchPlaylist(matches[1])
	} else {
		return nil, errors.New("unsupported JioSaavn URL format")
	}

	if err != nil {
		return nil, err
	}

	if len(tracks) > 0 {
		jiosaavnCache.Set(cacheKey, tracks)
		gologging.InfoF("JioSaavn: Successfully extracted %d track(s)", len(tracks))
	}

	return tracks, nil
}

func (j *JioSaavnPlatform) IsDownloadSupported(source state.PlatformName) bool {
	return source == PlatformJioSaavn
}

func (j *JioSaavnPlatform) Download(ctx context.Context, track *state.Track, _ *telegram.NewMessage) (string, error) {
	if path, err := checkDownloadedFile(track.ID); err == nil {
		gologging.InfoF("JioSaavn: Using cached file for %s", track.ID)
		return path, nil
	}

	gologging.InfoF("JioSaavn: Downloading %s", track.Title)

	if err := ensureDownloadsDir(); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	downloadURL, err := j.getDownloadURL(track.URL)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	filePath := filepath.Join("downloads", track.ID+".m4a")

	client := j.getClient()
	resp, err := client.R().
		SetContext(ctx).
		SetOutputFileName(filePath).
		Get(downloadURL)
	if err != nil {
		os.Remove(filePath)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", fmt.Errorf("download failed: %w", err)
	}

	if resp.IsError() {
		os.Remove(filePath)
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode())
	}

	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("downloaded file not found: %w", err)
	}

	gologging.InfoF("JioSaavn: Successfully downloaded %s", track.Title)
	return filePath, nil
}

func (j *JioSaavnPlatform) getClient() *resty.Client {
	if j.client == nil {
		j.client = resty.New().
			SetTimeout(30*time.Second).
			SetHeader("User-Agent", j.getRandomUserAgent()).
			SetHeader("Accept", "application/json").
			SetHeader("Accept-Language", "en-US,en;q=0.9")
	}
	return j.client
}

func (j *JioSaavnPlatform) getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func (j *JioSaavnPlatform) fetchSong(token string) ([]*state.Track, error) {
	client := j.getClient()

	url := fmt.Sprintf("https://www.jiosaavn.com/api.php?__call=webapi.get&token=%s&type=song&_format=json&_marker=0&ctx=web6dot0", token)

	var resp JioSaavnSongResponse
	apiResp, err := client.R().
		SetResult(&resp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch song: %w", err)
	}

	if apiResp.IsError() {
		return nil, fmt.Errorf("API returned error: %d", apiResp.StatusCode())
	}

	if len(resp.Songs) == 0 {
		return nil, errors.New("no songs found")
	}

	var tracks []*state.Track
	for _, song := range resp.Songs {
		track := j.songToTrack(&song)
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (j *JioSaavnPlatform) fetchAlbum(token string) ([]*state.Track, error) {
	client := j.getClient()

	url := fmt.Sprintf("https://www.jiosaavn.com/api.php?__call=webapi.get&token=%s&type=album&_format=json&_marker=0&ctx=web6dot0", token)

	var resp JioSaavnAlbumResponse
	apiResp, err := client.R().
		SetResult(&resp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch album: %w", err)
	}

	if apiResp.IsError() {
		return nil, fmt.Errorf("API returned error: %d", apiResp.StatusCode())
	}

	if len(resp.List) == 0 {
		return nil, errors.New("no songs found in album")
	}

	var tracks []*state.Track
	for _, song := range resp.List {
		track := &state.Track{
			ID:       song.ID,
			Title:    song.Title,
			Duration: j.parseDuration(song.Duration),
			Artwork:  j.getHighQualityImage(song.Image),
			URL:      song.PermaURL,
			Source:   PlatformJioSaavn,
			Video:    false,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (j *JioSaavnPlatform) fetchPlaylist(token string) ([]*state.Track, error) {
	client := j.getClient()

	url := fmt.Sprintf("https://www.jiosaavn.com/api.php?__call=webapi.get&token=%s&type=playlist&_format=json&_marker=0&ctx=web6dot0", token)

	var resp JioSaavnPlaylistResponse
	apiResp, err := client.R().
		SetResult(&resp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist: %w", err)
	}

	if apiResp.IsError() {
		return nil, fmt.Errorf("API returned error: %d", apiResp.StatusCode())
	}

	if len(resp.List) == 0 {
		return nil, errors.New("no songs found in playlist")
	}

	var tracks []*state.Track
	for _, song := range resp.List {
		track := &state.Track{
			ID:       song.ID,
			Title:    song.Title,
			Duration: j.parseDuration(song.Duration),
			Artwork:  j.getHighQualityImage(song.Image),
			URL:      song.PermaURL,
			Source:   PlatformJioSaavn,
			Video:    false,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (j *JioSaavnPlatform) songToTrack(song *struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Album    string `json:"album"`
	Year     string `json:"year"`
	Duration string `json:"duration"`
	Language string `json:"language"`
	Image    string `json:"image"`
	PermaURL string `json:"perma_url"`
	MoreInfo struct {
		EncryptedMediaURL string `json:"encrypted_media_url"`
		Album             string `json:"album"`
		AlbumID           string `json:"album_id"`
		Duration          string `json:"duration"`
	} `json:"more_info"`
},
) *state.Track {
	return &state.Track{
		ID:       song.ID,
		Title:    song.Title,
		Duration: j.parseDuration(song.MoreInfo.Duration),
		Artwork:  j.getHighQualityImage(song.Image),
		URL:      song.PermaURL,
		Source:   PlatformJioSaavn,
		Video:    false,
	}
}

func (j *JioSaavnPlatform) parseDuration(duration string) int {
	var dur int
	fmt.Sscanf(duration, "%d", &dur)
	return dur
}

func (j *JioSaavnPlatform) getHighQualityImage(imageURL string) string {
	imageURL = strings.Replace(imageURL, "150x150", "500x500", 1)
	imageURL = strings.Replace(imageURL, "50x50", "500x500", 1)
	imageURL = strings.Replace(imageURL, "http://", "https://", 1)
	return imageURL
}

func (j *JioSaavnPlatform) getDownloadURL(songURL string) (string, error) {
	var token string
	if matches := jiosaavnSongRegex.FindStringSubmatch(songURL); len(matches) > 1 {
		token = matches[1]
	} else {
		return "", errors.New("invalid song URL")
	}

	client := j.getClient()
	url := fmt.Sprintf("https://www.jiosaavn.com/api.php?__call=song.getDetails&pids=%s&_format=json&_marker=0&ctx=web6dot0", token)

	var resp struct {
		Songs []struct {
			MoreInfo struct {
				EncryptedMediaURL string `json:"encrypted_media_url"`
			} `json:"more_info"`
		} `json:"songs"`
	}

	apiResp, err := client.R().
		SetResult(&resp).
		Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch song details: %w", err)
	}

	if apiResp.IsError() || len(resp.Songs) == 0 {
		return "", errors.New("failed to get song details")
	}

	encryptedURL := resp.Songs[0].MoreInfo.EncryptedMediaURL
	if encryptedURL == "" {
		return "", errors.New("no encrypted media URL found")
	}

	decryptedURL, err := j.decryptURL(encryptedURL)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt URL: %w", err)
	}

	decryptedURL = strings.Replace(decryptedURL, "_96", "_320", 1)

	return decryptedURL, nil
}

func (j *JioSaavnPlatform) decryptURL(encryptedURL string) (string, error) {
	key := []byte("38346591")

	encrypted, err := base64.StdEncoding.DecodeString(encryptedURL)
	if err != nil {
		return "", err
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(encrypted))

	for i := 0; i < len(encrypted); i += block.BlockSize() {
		block.Decrypt(decrypted[i:i+block.BlockSize()], encrypted[i:i+block.BlockSize()])
	}

	decrypted = j.removePadding(decrypted)

	return string(decrypted), nil
}

func (j *JioSaavnPlatform) removePadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padding := data[len(data)-1]
	if int(padding) > len(data) {
		return data
	}

	return data[:len(data)-int(padding)]
}

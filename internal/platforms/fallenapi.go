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
package platforms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/state"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

type FallenApiPlatform struct{}

func init() {
	addPlatform(80, state.PlatformFallenApi, &FallenApiPlatform{})
}

func (*FallenApiPlatform) Name() state.PlatformName {
	return state.PlatformFallenApi
}

func (*FallenApiPlatform) IsValid(query string) bool {
	return false
}

func (*FallenApiPlatform) GetTracks(query string) ([]*state.Track, error) {
	return nil, errors.New("FallenApi is a download-only platform")
}

func (*FallenApiPlatform) IsDownloadSupported(source state.PlatformName) bool {
	if config.ApiURL == "" || config.ApiKEY == "" {
		return false
	}
	return source == state.PlatformYouTube
}

var telegramDLRegex = regexp.MustCompile(`https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)`)

type APIResponse struct {
	CdnUrl string `json:"cdnurl"`
}

func (f *FallenApiPlatform) Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
	pm := utils.GetProgress(mystic)

	if path, err := f.checkDownloadedFile(track.ID); err == nil {
		return path, nil
	}

	dlURL, err := f.getDownloadURL(ctx, track.URL)
	if err != nil {
		return "", err
	}

	os.MkdirAll("downloads", os.ModePerm)
	filePath := filepath.Join("downloads", track.ID+".webm")

	var downloadErr error
	if telegramDLRegex.MatchString(dlURL) {
		filePath, downloadErr = f.downloadFromTelegram(ctx, mystic.Client, dlURL, track.ID, pm)
	} else {
		downloadErr = f.downloadFromURL(ctx, dlURL, filePath)
	}

	if downloadErr != nil {
		if _, err := os.Stat(filePath); err == nil {
			os.Remove(filePath)
		}
		return "", downloadErr
	}

	return filePath, nil
}

func (f *FallenApiPlatform) getDownloadURL(ctx context.Context, mediaURL string) (string, error) {
	apiReqURL := fmt.Sprintf("%s/track?api_key=%s&url=%s", config.ApiURL, config.ApiKEY, url.QueryEscape(mediaURL))
	req, err := http.NewRequestWithContext(ctx, "GET", apiReqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("invalid API response: %w", err)
	}

	if apiResp.CdnUrl == "" {
		return "", fmt.Errorf("empty API response")
	}

	return apiResp.CdnUrl, nil
}

func (f *FallenApiPlatform) downloadFromURL(ctx context.Context, dlURL, filePath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", dlURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (f *FallenApiPlatform) downloadFromTelegram(ctx context.Context, client *telegram.Client, dlURL, videoId string, pm *telegram.ProgressManager) (string, error) {
	matches := telegramDLRegex.FindStringSubmatch(dlURL)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid telegram download url: %s", dlURL)
	}

	username := matches[1]
	messageID, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid message ID: %v", err)
	}

	msg, err := client.GetMessageByID(username, int32(messageID))
	if err != nil {
		return "", fmt.Errorf("failed to fetch Telegram message: %w", err)
	}

	rawFile := filepath.Join("downloads", videoId+msg.File.Ext)

	dOpts := &telegram.DownloadOptions{FileName: rawFile, Threads: 3}
	if pm != nil {
		dOpts.ProgressManager = pm
	}

	done := make(chan error)
	go func() {
		_, err := msg.Download(dOpts)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("telegram download failed: %w", err)
		}
	}

	return rawFile, nil
}

func (f *FallenApiPlatform) checkDownloadedFile(videoId string) (string, error) {
	outputDir := "./downloads"
	pattern := filepath.Join(outputDir, videoId+".*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to search files: %v", err)
	}

	if len(matches) == 0 {
		return "", errors.New("❌ file not found")
	}

	// If multiple matches, pick the first one
	return matches[0], nil
}

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
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
	"resty.dev/v3"

	"main/internal/core"
	"main/internal/core/models"
	"main/internal/utils"
)

var (
	telegramDLRegex = regexp.MustCompile(`https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)`)
	fallenAPIURL    = os.Getenv("FALLEN_API_URL")
	fallenAPIKey    = os.Getenv("FALLEN_API_KEY")
)

const PlatformFallenApi state.PlatformName = "FallenApi"

type APIResponse struct {
	CdnUrl string `json:"cdnurl"`
}

type FallenApiPlatform struct{}

func init() {
	addPlatform(80, PlatformFallenApi, &FallenApiPlatform{})
	if fallenAPIURL == "" {
		fallenAPIURL = "https://tgmusic.fallenapi.fun"
	}
}

func (*FallenApiPlatform) Name() state.PlatformName {
	return PlatformFallenApi
}

func (*FallenApiPlatform) IsValid(query string) bool {
	return false
}

func (*FallenApiPlatform) GetTracks(_ string, _ bool) ([]*state.Track, error) {
	return nil, errors.New("fallenapi is a download-only platform")
}

func (*FallenApiPlatform) IsDownloadSupported(source state.PlatformName) bool {
	if fallenAPIURL == "" || fallenAPIKey == "" {
		return false
	}
	return source == PlatformYouTube
}

func (f *FallenApiPlatform) Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
	// fallen api didn't support video downloads so disable it
	track.Video = false
	var pm *telegram.ProgressManager
	if mystic != nil {
		pm = utils.GetProgress(mystic)
	}

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
		filePath, downloadErr = f.downloadFromTelegram(ctx, dlURL, track.ID, pm)
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
	apiReqURL := fmt.Sprintf("%s/track?api_key=%s&url=%s", fallenAPIURL, fallenAPIKey, url.QueryEscape(mediaURL))

	client := resty.New()
	var apiResp APIResponse

	resp, err := client.R().
		SetContext(ctx).
		SetResult(&apiResp).
		Get(apiReqURL)
	if err != nil {
		return "", fmt.Errorf("api request failed: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("api request failed with status: %d", resp.StatusCode())
	}

	if apiResp.CdnUrl == "" {
		return "", fmt.Errorf("empty API response")
	}

	return apiResp.CdnUrl, nil
}

func (f *FallenApiPlatform) downloadFromURL(ctx context.Context, dlURL, filePath string) error {
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		SetOutputFileName(filePath).
		Get(dlURL)
	if err != nil {
		return fmt.Errorf("http download failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode())
	}

	return nil
}

func (f *FallenApiPlatform) downloadFromTelegram(ctx context.Context, dlURL, videoId string, pm *telegram.ProgressManager) (string, error) {
	matches := telegramDLRegex.FindStringSubmatch(dlURL)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid telegram download url: %s", dlURL)
	}

	username := matches[1]
	messageID, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid message ID: %v", err)
	}

	msg, err := core.Bot.GetMessageByID(username, int32(messageID))
	if err != nil {
		return "", fmt.Errorf("failed to fetch Telegram message: %w", err)
	}

	rawFile := filepath.Join("downloads", videoId+msg.File.Ext)

	dOpts := &telegram.DownloadOptions{
		FileName: rawFile,
		Threads:  3,
		Ctx:      ctx,
	}
	if pm != nil {
		dOpts.ProgressManager = pm
	}
	_, err = msg.Download(dOpts)
	if err != nil {
		os.Remove(rawFile)
		return "", err
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
		return "", errors.New("file not found")
	}

	// If multiple matches, pick the first one
	return matches[0], nil
}

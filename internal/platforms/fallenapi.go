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

package platforms

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/utils"
)

const PlatformFallenApi state.PlatformName = "FallenApi"

type FallenApiPlatform struct{}

var telegramDLRegex = regexp.MustCompile(
	`https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)`,
)

type fallenAPIResponse struct {
	CdnUrl string `json:"cdnurl"`
}

func init() {
	Register(&FallenApiPlatform{})
}

func (f *FallenApiPlatform) Name() state.PlatformName { return PlatformFallenApi }
func (f *FallenApiPlatform) Priority() int             { return 80 }

func (f *FallenApiPlatform) CanGet(_ string) bool { return false }

func (f *FallenApiPlatform) Get(_ string, _ bool) ([]*state.Track, error) {
	return nil, errors.New("fallenapi is download-only")
}

func (f *FallenApiPlatform) CanDownload(source state.PlatformName) bool {
	if config.FallenAPIURL == "" || config.FallenAPIKey == "" {
		return false
	}
	return source == PlatformYouTube
}

func (f *FallenApiPlatform) Download(
	ctx context.Context,
	track *state.Track,
	statusMsg *telegram.NewMessage,
) (string, error) {
	track.Video = false

	if p := findFile(track); p != "" {
		gologging.Debug("FallenApi: cache hit " + p)
		return p, nil
	}

	var pm *telegram.ProgressManager
	if statusMsg != nil {
		pm = utils.GetProgress(statusMsg)
	}

	dlURL, err := f.getDownloadURL(ctx, track.URL)
	if err != nil {
		return "", err
	}

	path := getPath(track, ".mp3")

	if telegramDLRegex.MatchString(dlURL) {
		return f.downloadFromTelegram(ctx, dlURL, path, pm)
	}

	if err := f.downloadFromURL(ctx, dlURL, path); err != nil {
		return "", err
	}

	if !fileExists(path) {
		return "", errors.New("API returned empty file")
	}

	return path, nil
}

func (f *FallenApiPlatform) getDownloadURL(ctx context.Context, mediaURL string) (string, error) {
	apiURL := fmt.Sprintf(
		"%s/api/track?api_key=%s&url=%s",
		config.FallenAPIURL,
		config.FallenAPIKey,
		url.QueryEscape(mediaURL),
	)

	var resp fallenAPIResponse
	r, err := rc.R().
		SetContext(ctx).
		SetResult(&resp).
		Get(apiURL)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", sanitizeAPIError(
			fmt.Errorf("API request failed: %w", err),
			config.FallenAPIKey,
		)
	}

	if r.IsError() {
		return "", sanitizeAPIError(fmt.Errorf(
			"API returned %d: %s", r.StatusCode(), r.String(),
		), config.FallenAPIKey)
	}

	if resp.CdnUrl == "" {
		return "", sanitizeAPIError(
			fmt.Errorf("empty cdnurl in response: %s", r.String()),
			config.FallenAPIKey,
		)
	}

	return resp.CdnUrl, nil
}

func (f *FallenApiPlatform) downloadFromURL(ctx context.Context, dlURL, path string) error {
	r, err := rc.R().
		SetContext(ctx).
		SetOutputFileName(path).
		Get(dlURL)
	if err != nil {
		os.Remove(path)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("http download failed: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("download returned %d", r.StatusCode())
	}
	return nil
}

func (f *FallenApiPlatform) downloadFromTelegram(
	ctx context.Context,
	dlURL, path string,
	pm *telegram.ProgressManager,
) (string, error) {
	matches := telegramDLRegex.FindStringSubmatch(dlURL)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid telegram download url: %s", dlURL)
	}

	username := matches[1]
	msgID, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid message ID: %w", err)
	}

	msg, err := core.Bot.GetMessageByID(username, int32(msgID))
	if err != nil {
		return "", fmt.Errorf("failed to fetch Telegram message: %w", err)
	}

	dOpts := &telegram.DownloadOptions{FileName: path, Ctx: ctx}
	if pm != nil {
		dOpts.ProgressManager = pm
	}

	if _, err = msg.Download(dOpts); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}
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

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
)

const PlatformFallenApi state.PlatformName = "FallenApi"

type FallenApiPlatform struct{}

var (
	telegramDLRegex = regexp.MustCompile(
		`https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)`,
	)

	// fallenAPIURLRe matches URLs of the platforms the Fallen API can resolve
	// and download. YouTube and Telegram are excluded because they have their
	// own dedicated platforms.
	fallenAPIURLRe = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^https?:\/\/music\.apple\.com\/[a-zA-Z-]+\/(?:song\/(?:[^\/]+\/)?\d+|album\/[^\/]+\/\d+(?:\?i=\d+)?|playlist\/[^\/]+\/pl\.[\w.-]+|artist\/[^\/]+\/\d+)(?:\?.*)?$`),
		regexp.MustCompile(`(?i)^(https?://)?([a-z0-9-]+\.)*spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+(\?.*)?$`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?(?:jiosaavn|saavn)\.com\/(?:s\/)?(song|album|playlist|featured)(?:\/[^\/]+)*\/([A-Za-z0-9_,-]+)(?:\/)?(?:\?.*)?$`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?deezer\.com\/(?:[a-z]{2}\/)?(track|album|playlist)\/(\d+)`),
		regexp.MustCompile(`(?i)^(https?://)?(www\.)?soundcloud\.com/[a-zA-Z0-9_-]+/(sets/)?[a-zA-Z0-9._-]+(\?.*)?$`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?gaana\.com\/(song|album|playlist|artist)\/([A-Za-z0-9\-]+)`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.|listen\.)?tidal\.com\/(?:browse\/)?(track|album|playlist)\/([a-zA-Z0-9-]+)(?:[\/?].*)?`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?mxplayer\.in\/(?:show|movie)\/.*`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?twitch\.tv\/(?:videos|[\w._-]+\/video)\/\d+`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?(?:twitch\.tv\/clip\/[\w-]+|clips\.twitch\.tv\/[\w-]+|twitch\.tv\/[\w-]+\/clip\/[\w-]+)`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/videos\/[a-fA-F0-9-]+`),
		regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/clips\/[\w-]+`),
	}
)

type fallenAPITrack struct {
	Title     string `json:"title"`
	ID        string `json:"id"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Duration  int    `json:"duration"`
}

type fallenAPIResults struct {
	Results []fallenAPITrack `json:"results"`
}

type fallenAPIResponse struct {
	CdnUrl string `json:"cdnurl"`
}

func init() {
	Register(&FallenApiPlatform{})
}

func (f *FallenApiPlatform) Name() state.PlatformName { return PlatformFallenApi }
func (f *FallenApiPlatform) Priority() int            { return 80 }

func (f *FallenApiPlatform) CanGet(query string) bool {
	if config.FallenAPIURL == "" || config.FallenAPIKey == "" {
		return false
	}
	for _, re := range fallenAPIURLRe {
		if re.MatchString(query) {
			return true
		}
	}
	return false
}

// Get resolves a supported platform URL (Apple Music, Spotify, Deezer,
// SoundCloud, etc.) through the Fallen API into one or more tracks.
func (f *FallenApiPlatform) Get(query string, _ bool) ([]*state.Track, error) {
	if config.FallenAPIURL == "" || config.FallenAPIKey == "" {
		return nil, errors.New("FALLEN_API_KEY not configured")
	}

	var resp fallenAPIResults
	apiURL := fmt.Sprintf(
		"%s/api/get_url?api_key=%s&url=%s",
		config.FallenAPIURL,
		config.FallenAPIKey,
		url.QueryEscape(query),
	)

	r, err := rc.R().
		SetHeader("X-API-Key", config.FallenAPIKey).
		SetResult(&resp).
		Get(apiURL)
	if err != nil {
		return nil, sanitizeAPIError(
			fmt.Errorf("API request failed: %w", err),
			config.FallenAPIKey,
		)
	}
	if r.IsStatusFailure() {
		return nil, sanitizeAPIError(fmt.Errorf(
			"API returned %d: %s", r.StatusCode(), r.String(),
		), config.FallenAPIKey)
	}

	tracks := make([]*state.Track, 0, len(resp.Results))
	for _, mt := range resp.Results {
		if mt.Title == "" || mt.URL == "" {
			continue
		}
		tracks = append(tracks, &state.Track{
			ID:       mt.ID,
			Title:    mt.Title,
			Duration: mt.Duration,
			Artwork:  mt.Thumbnail,
			URL:      mt.URL,
			Source:   PlatformFallenApi,
		})
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no results for: %s", query)
	}
	return tracks, nil
}

func (f *FallenApiPlatform) CanDownload(source state.PlatformName) bool {
	if config.FallenAPIURL == "" || config.FallenAPIKey == "" {
		return false
	}
	return source == PlatformFallenApi || source == PlatformYouTube
}

func (f *FallenApiPlatform) Download(
	ctx context.Context,
	track *state.Track,
	statusMsg *td.Message,
) (string, error) {
	track.Video = false

	if p := findFile(track); p != "" {
		logger.Debug("FallenApi: cache hit " + p)
		return p, nil
	}

	dlURL, err := f.getDownloadURL(ctx, track.URL)
	if err != nil {
		return "", err
	}

	path := getPath(track, ".mp3")

	if telegramDLRegex.MatchString(dlURL) {
		return f.downloadFromTelegram(ctx, dlURL, path, statusMsg)
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
		SetHeader("X-API-Key", config.FallenAPIKey).
		SetResult(&resp).
		Get(apiURL)
	if err != nil {
		if isDownloadCancelled(err) {
			return "", err
		}
		return "", sanitizeAPIError(
			fmt.Errorf("API request failed: %w", err),
			config.FallenAPIKey,
		)
	}

	if r.IsStatusFailure() {
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
		SetResponseSaveFileName(path).
		Get(dlURL)
	if err != nil {
		os.Remove(path)
		if isDownloadCancelled(err) {
			return err
		}
		return fmt.Errorf("http download failed: %w", err)
	}
	if r.IsStatusFailure() {
		return fmt.Errorf("download returned %d", r.StatusCode())
	}
	return nil
}

func (f *FallenApiPlatform) downloadFromTelegram(
	ctx context.Context,
	dlURL, path string,
	statusMsg *td.Message,
) (string, error) {
	if !telegramDLRegex.MatchString(dlURL) {
		return "", fmt.Errorf("invalid telegram download url: %s", dlURL)
	}

	info, err := core.Bot.GetMessageLinkInfo(dlURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Telegram message: %w", err)
	}
	if info == nil || info.Message == nil {
		return "", errors.New("failed to fetch Telegram message")
	}
	msg := info.Message

	fileID := msg.RemoteFileID()
	if OnDownloadStart != nil {
		OnDownloadStart(fileID, statusMsg)
	}
	file, err := msg.Download(core.Bot, 1, 0, 0, true)
	if err != nil {
		os.Remove(path)
		if isDownloadCancelled(err) {
			return "", context.Canceled
		}
		return "", err
	}
	if file == nil || file.Local == nil || file.Local.Path == "" {
		return "", errors.New("downloaded file missing")
	}

	if err := copyFile(file.Local.Path, path); err != nil {
		os.Remove(path)
		return "", err
	}
	os.Remove(file.Local.Path)
	return path, nil
}

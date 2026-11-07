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

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
)

const PlatformFallenApi state.PlatformName = "FallenApi"

type FallenApiPlatform struct{}

var (
	telegramDLRegex = regexp.MustCompile(`^https://t\.me/([a-zA-Z0-9_]{5,})/(\d+)`)

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

type fallenTrack struct {
	Title     string `json:"title"`
	ID        string `json:"id"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Duration  int    `json:"duration"`
}

func init() {
	Register(&FallenApiPlatform{})
}

func (f *FallenApiPlatform) Name() state.PlatformName { return PlatformFallenApi }
func (f *FallenApiPlatform) Priority() int            { return 80 }

func fallenConfigured() bool {
	return config.FallenAPIURL != "" && config.FallenAPIKey != ""
}

func (f *FallenApiPlatform) CanGet(query string) bool {
	if !fallenConfigured() {
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
	if !fallenConfigured() {
		return nil, errors.New("FALLEN_API_KEY not configured")
	}

	var resp struct {
		Results []fallenTrack `json:"results"`
	}
	if _, err := fallenRequest(context.Background(), "get_url", query, &resp); err != nil {
		return nil, err
	}

	tracks := make([]*state.Track, 0, len(resp.Results))
	for _, t := range resp.Results {
		if t.Title == "" || t.URL == "" {
			continue
		}
		tracks = append(tracks, &state.Track{
			ID:       t.ID,
			Title:    t.Title,
			Duration: t.Duration,
			Artwork:  t.Thumbnail,
			URL:      t.URL,
			Source:   PlatformFallenApi,
		})
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no results for: %s", query)
	}
	return tracks, nil
}

func (f *FallenApiPlatform) CanDownload(source state.PlatformName) bool {
	return fallenConfigured() && (source == PlatformFallenApi || source == PlatformYouTube)
}

// Download always fetches audio. The track is flipped to audio only once that
// worked, so a failure leaves the video request intact for the next platform.
func (f *FallenApiPlatform) Download(
	ctx context.Context,
	track *state.Track,
	statusMsg *td.Message,
) (string, error) {
	audio := *track
	audio.Video = false

	path, err := f.download(ctx, &audio, statusMsg)
	if err != nil {
		return "", err
	}

	track.Video = false
	return path, nil
}

func (f *FallenApiPlatform) download(
	ctx context.Context,
	track *state.Track,
	statusMsg *td.Message,
) (string, error) {
	if path := findFile(track); path != "" {
		return path, nil
	}

	var resp struct {
		CdnURL string `json:"cdnurl"`
	}
	body, err := fallenRequest(ctx, "track", track.URL, &resp)
	if err != nil {
		return "", err
	}
	if resp.CdnURL == "" {
		return "", sanitizeAPIError(
			fmt.Errorf("empty cdnurl in response: %s", body),
			config.FallenAPIKey,
		)
	}

	path := trackFile(track, ".mp3")
	if telegramDLRegex.MatchString(resp.CdnURL) {
		return downloadFromTelegram(ctx, resp.CdnURL, path, statusMsg)
	}

	if err := downloadFromURL(ctx, resp.CdnURL, path); err != nil {
		return "", err
	}
	if !fileExists(path) {
		return "", errors.New("API returned empty file")
	}
	return path, nil
}

// fallenRequest calls an API endpoint for mediaURL, decodes the JSON reply into
// out and returns the raw body.
func fallenRequest(ctx context.Context, endpoint, mediaURL string, out any) (string, error) {
	apiURL := fmt.Sprintf(
		"%s/api/%s?api_key=%s&url=%s",
		config.FallenAPIURL,
		endpoint,
		config.FallenAPIKey,
		url.QueryEscape(mediaURL),
	)

	r, err := rc.R().
		SetContext(ctx).
		SetHeader("X-API-Key", config.FallenAPIKey).
		SetResult(out).
		Get(apiURL)
	if err != nil {
		if canceled(ctx, err) {
			return "", context.Canceled
		}
		return "", sanitizeAPIError(fmt.Errorf("API request failed: %w", err), config.FallenAPIKey)
	}
	if r.IsStatusFailure() {
		return "", sanitizeAPIError(
			fmt.Errorf("API returned %d: %s", r.StatusCode(), r.String()),
			config.FallenAPIKey,
		)
	}
	return r.String(), nil
}

func downloadFromURL(ctx context.Context, dlURL, path string) error {
	r, err := dl.R().
		SetContext(ctx).
		SetResponseSaveFileName(path).
		Get(dlURL)
	if err == nil && r.IsStatusFailure() {
		err = fmt.Errorf("download returned %d", r.StatusCode())
	} else if err != nil {
		if canceled(ctx, err) {
			err = context.Canceled
		} else {
			err = fmt.Errorf("http download failed: %w", err)
		}
	}

	if err != nil {
		_ = os.Remove(path)
	}
	return err
}

func downloadFromTelegram(
	ctx context.Context,
	dlURL, path string,
	statusMsg *td.Message,
) (string, error) {
	info, err := core.Bot.GetMessageLinkInfo(dlURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Telegram message: %w", err)
	}
	if info == nil || info.Message == nil {
		return "", errors.New("failed to fetch Telegram message")
	}

	return downloadTDLibMessage(ctx, info.Message, path, statusMsg)
}

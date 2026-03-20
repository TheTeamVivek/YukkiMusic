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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	state "main/internal/core/models"
	"main/internal/utils"
)

type YouTubePlatform struct {
	name state.PlatformName
}

var (
	youtubeLinkRegex = regexp.MustCompile(
		`(?i)^(?:https?:\/\/)?(?:www\.|m\.|music\.)?(?:youtube\.com|youtu\.be)\/\S+`,
	)
	youtubeCache = utils.NewCache[string, []*state.Track](1 * time.Hour)
)

const (
	PlatformYouTube        state.PlatformName = "YouTube"
	innerTubeKey                              = "AIzaSyBOti4mM-6x9WDnZIjIeyEU21OpBXqWBgw"
	innerTubeClientVersion                    = "2.20250101.01.00"
	innerTubeClientName                       = "WEB"
)

var yt = &YouTubePlatform{
	name: PlatformYouTube,
}

func init() {
	Register(90, yt)
}

func (yp *YouTubePlatform) Name() state.PlatformName {
	return yp.name
}

func (yp *YouTubePlatform) CanGetTracks(link string) bool {
	return youtubeLinkRegex.MatchString(link)
}

func (yp *YouTubePlatform) GetTracks(
	input string,
	video bool,
) ([]*state.Track, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, errors.New("empty query")
	}

	var tracks []*state.Track
	var err error

	if youtubeLinkRegex.MatchString(trimmed) {
		u, err := url.Parse(trimmed)
		if err != nil {
			return nil, fmt.Errorf("failed to parse url %w", err)
		}

		q := u.Query()

		if q.Get("list") != "" && q.Get("v") == "" {
			tracks, err = yp.handlePlaylist(trimmed)
		} else {
			tracks, err = yp.handleTrackURL(trimmed)
		}
	} else {
		tracks, err = yp.VideoSearch(trimmed, false)
	}

	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, errors.New("no tracks found")
	}

	return updateCached(tracks, video), nil
}

func (yp *YouTubePlatform) handlePlaylist(
	rawURL string,
) ([]*state.Track, error) {
	cacheKey := "playlist:" + strings.ToLower(rawURL)
	if cached, ok := youtubeCache.Get(cacheKey); ok {
		return cached, nil
	}

	playlistID := yp.extractPlaylistID(rawURL)
	if playlistID != "" {
		tracks, err := scrapePlaylistYouTube(playlistID, config.QueueLimit)
		if err == nil && len(tracks) > 0 {
			youtubeCache.Set(cacheKey, tracks)
			return tracks, nil
		}
	}

	videoIDs, err := getPlaylist(rawURL)
	if err != nil {
		return nil, err
	}

	type result struct {
		index int
		track *state.Track
	}

	resChan := make(chan result, len(videoIDs))
	sem := make(chan struct{}, 3) // Semantic 3 concurrency

	for i, id := range videoIDs {
		go func(idx int, vID string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			if cached, ok := youtubeCache.Get("track:" + vID); ok &&
				len(cached) > 0 {
				resChan <- result{idx, cached[0]}
				return
			}

			fullURL := "https://youtube.com/watch?v=" + vID
			found := false
			for _, query := range []string{vID, fullURL} {
				searchResult, err := yp.VideoSearch(query, true)
				if err != nil || len(searchResult) == 0 {
					continue
				}

				for _, t := range searchResult {
					if t.ID == vID {
						youtubeCache.Set("track:"+vID, []*state.Track{t})
						resChan <- result{idx, t}
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				gologging.ErrorF(
					"[YouTube] Failed to resolve track metadata for ID: %s\n",
					vID,
				)
				resChan <- result{idx, nil}
			}
		}(i, id)
	}

	orderedTracks := make([]*state.Track, len(videoIDs))
	for i := 0; i < len(videoIDs); i++ {
		res := <-resChan
		orderedTracks[res.index] = res.track
	}

	var finalTracks []*state.Track
	for _, t := range orderedTracks {
		if t != nil {
			finalTracks = append(finalTracks, t)
		}
	}

	if len(finalTracks) > 0 {
		youtubeCache.Set(cacheKey, finalTracks)
	}
	return finalTracks, nil
}

func (yp *YouTubePlatform) handleTrackURL(
	rawURL string,
) ([]*state.Track, error) {
	_, videoID, err := yp.normalizeYouTubeURL(rawURL)
	if err != nil {
		return nil, err
	}

	if cached, ok := youtubeCache.Get("track:" + videoID); ok &&
		len(cached) > 0 {
		return cached, nil
	}

	for _, query := range []string{videoID, rawURL} {
		results, err := yp.VideoSearch(query, true)
		if err != nil {
			continue
		}

		for _, t := range results {
			if t.ID == videoID {
				youtubeCache.Set("track:"+videoID, []*state.Track{t})
				return []*state.Track{t}, nil
			}
		}
	}

	return nil, errors.New("track not found")
}

func (yp *YouTubePlatform) CanDownload(source state.PlatformName) bool {
	return false
}

func (yt *YouTubePlatform) Download(
	ctx context.Context,
	track *state.Track,
	statusMsg *telegram.NewMessage,
) (string, error) {
	return "", errors.New("youtube platform does not support downloading")
}

func (*YouTubePlatform) CanSearch() bool { return true }

func (y *YouTubePlatform) Search(
	q string,
	video bool,
) ([]*state.Track, error) {
	return y.GetTracks(q, video)
}

func (yp *YouTubePlatform) VideoSearch(
	query string,
	singleOpt ...bool,
) ([]*state.Track, error) {
	single := false
	limit := config.QueueLimit
	if len(singleOpt) > 0 && singleOpt[0] {
		single = true
		limit = 1
	}

	cacheKey := "search:" + strings.TrimSpace(strings.ToLower(query))
	if arr, ok := youtubeCache.Get(cacheKey); ok {
		if single && len(arr) > 0 {
			return []*state.Track{arr[0]}, nil
		}
		if !single && len(arr) == 1 {
			// goto Search
		} else {
			return arr, nil
		}
	}

	var tracks []*state.Track
	var err error

	tracks, err = searchYouTube(query, limit)
	if err != nil {
		return nil, fmt.Errorf("ytsearch failed: %w", err)
	}

	if len(tracks) == 0 {
		return nil, errors.New("no tracks found")
	}

	youtubeCache.Set(cacheKey, tracks)

	if single {
		return []*state.Track{tracks[0]}, nil
	}

	return tracks, nil
}

func (yt *YouTubePlatform) extractPlaylistID(input string) string {
	u, err := url.Parse(strings.TrimSpace(input))
	if err != nil {
		return ""
	}
	return u.Query().Get("list")
}

func (yt *YouTubePlatform) normalizeYouTubeURL(
	input string,
) (string, string, error) {
	u, err := url.Parse(strings.TrimSpace(input))
	if err != nil {
		return "", "", err
	}

	host := strings.ToLower(u.Host)
	path := strings.Trim(u.Path, "/")

	if host == "youtu.be" {
		id := strings.Split(path, "/")[0]
		if len(id) == 11 {
			return "https://www.youtube.com/watch?v=" + id, id, nil
		}
	}

	if host == "youtube.com" ||
		host == "www.youtube.com" ||
		host == "m.youtube.com" ||
		host == "music.youtube.com" {

		if v := u.Query().Get("v"); len(v) == 11 {
			return "https://www.youtube.com/watch?v=" + v, v, nil
		}

		parts := strings.Split(path, "/")

		if len(parts) >= 2 {
			id := parts[1]

			switch parts[0] {
			case "shorts":
				if len(id) == 11 {
					return "https://www.youtube.com/watch?v=" + id, id, nil
				}
			case "embed":
				if len(id) == 11 {
					return "https://www.youtube.com/watch?v=" + id, id, nil
				}
			case "live":
				if len(id) == 11 {
					return "https://www.youtube.com/watch?v=" + id, id, nil
				}
			case "v":
				if len(id) == 11 {
					return "https://www.youtube.com/watch?v=" + id, id, nil
				}
			}
		}
	}

	return "", "", errors.New("unsupported YouTube URL or missing video ID")
}

func getPlaylist(pUrl string) ([]string, error) {
	if strings.Contains(pUrl, "&") {
		pUrl = strings.Split(pUrl, "&")[0]
	}

	cmd := exec.Command(
		"yt-dlp",
		"-i",
		"--compat-options",
		"no-youtube-unavailable-videos",
		"--get-id",
		"--flat-playlist",
		"--skip-download",
		"--playlist-end",
		strconv.Itoa(config.QueueLimit),
		pUrl,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp error: %v\n%s", err, stderr.String())
	}

	return strings.Split(strings.TrimSpace(out.String()), "\n"), nil
}

func updateCached(tracks []*state.Track, video bool) []*state.Track {
	if len(tracks) == 0 {
		return nil
	}

	out := make([]*state.Track, 0, len(tracks))
	for _, t := range tracks {
		if t == nil {
			continue
		}

		trackCopy := *t
		trackCopy.Video = video

		out = append(out, &trackCopy)
	}

	return out
}

// The following search functions are adapted from TgMusicBot.
// Copyright (c) 2025 Ashok Shau
// Licensed under GNU GPL v3
// See https://github.com/AshokShau/TgMusicBot
//
// searchYouTube scrapes YouTube results page

func searchYouTube(query string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] Searching for: %s (limit: %d)", query, limit)
	searchURL := "https://m.youtube.com/youtubei/v1/search?key=" + innerTubeKey
	var result map[string]any

	resp, err := rc.
		R().
		SetResult(&result).
		SetBody(map[string]any{
			"context": map[string]any{
				"client": map[string]any{
					"clientName":       innerTubeClientName,
					"clientVersion":    innerTubeClientVersion,
					"newVisitorCookie": true,
					"acceptHeader":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
					"hl":               "en-IN",
					"gl":               "IN",
				},
			},
			"request": map[string]any{
				"useSsl": true,
			},
			"user": map[string]any{
				"lockedSafetyMode": false,
			},
			"params": "CAASAhAB",
			"query":  query,
		}).
		SetHeaderMultiValues(map[string][]string{
			"User-Agent": {
				"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36",
			},
			"Accept": {
				"*/*",
			},
			"Content-Type": {
				"application/json",
			},
			"x-origin": {
				"https://m.youtube.com",
			},
			"origin": {
				"https://m.youtube.com",
			},
			"accept-language": {
				"en-IN",
			},
		}).Post(searchURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	contents := dig(
		result,
		"contents",
		"twoColumnSearchResultsRenderer",
		"primaryContents",
		"sectionListRenderer",
		"contents",
	)
	if contents == nil {
		return nil, fmt.Errorf("no contents found")
	}

	var tracks []*state.Track
	parseSearchResults(contents, &tracks, limit)
	gologging.DebugF("[YouTube] Search found %d tracks", len(tracks))
	return tracks, nil
}

func scrapePlaylistYouTube(playlistID string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] Scraping playlist: %s (limit: %d)", playlistID, limit)
	browseURL := "https://m.youtube.com/youtubei/v1/browse?key=" + innerTubeKey
	var result map[string]any

	browseId := playlistID
	if !strings.HasPrefix(playlistID, "VL") {
		browseId = "VL" + playlistID
	}

	resp, err := rc.
		R().
		SetResult(&result).
		SetBody(map[string]any{
			"context": map[string]any{
				"client": map[string]any{
					"clientName":       innerTubeClientName,
					"clientVersion":    innerTubeClientVersion,
					"newVisitorCookie": true,
					"acceptHeader":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
					"hl":               "en-IN",
					"gl":               "IN",
				},
			},
			"browseId": browseId,
		}).
		SetHeaderMultiValues(map[string][]string{
			"User-Agent": {
				"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36",
			},
			"Accept": {
				"*/*",
			},
			"Content-Type": {
				"application/json",
			},
			"x-origin": {
				"https://m.youtube.com",
			},
			"origin": {
				"https://m.youtube.com",
			},
			"accept-language": {
				"en-IN",
			},
		}).Post(browseURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	var tracks []*state.Track
	parsePlaylistContents(result, &tracks, limit)
	gologging.DebugF("[YouTube] Playlist scraper found %d tracks", len(tracks))
	return tracks, nil
}

func parsePlaylistContents(node any, tracks *[]*state.Track, limit int) {
	if limit > 0 && len(*tracks) >= limit {
		return
	}

	switch v := node.(type) {
	case []any:
		for _, item := range v {
			if limit > 0 && len(*tracks) >= limit {
				return
			}
			parsePlaylistContents(item, tracks, limit)
		}
	case map[string]any:
		if vid, ok := dig(v, "playlistVideoRenderer").(map[string]any); ok {
			id := safeString(vid["videoId"])
			title := safeString(dig(vid, "title", "runs", 0, "text"))
			thumb := getThumbnailURL(vid)
			durationText := safeString(dig(vid, "lengthText", "simpleText"))

			if durationText == "" {
				return
			}

			duration := parseDuration(durationText)
			t := &state.Track{
				URL:      "https://www.youtube.com/watch?v=" + id,
				Title:    title,
				ID:       id,
				Artwork:  thumb,
				Duration: duration,
				Source:   PlatformYouTube,
			}
			*tracks = append(*tracks, t)
			youtubeCache.Set("track:"+t.ID, []*state.Track{t})
		} else {
			for _, child := range v {
				parsePlaylistContents(child, tracks, limit)
			}
		}
	}
}

func parseSearchResults(node any, tracks *[]*state.Track, limit int) {
	if limit > 0 && len(*tracks) >= limit {
		return
	}

	switch v := node.(type) {

	case []any:
		for _, item := range v {
			if limit > 0 && len(*tracks) >= limit {
				return
			}
			parseSearchResults(item, tracks, limit)
		}

	case map[string]any:
		if vid, ok := dig(v, "videoRenderer").(map[string]any); ok {

			if isLiveVideo(vid) {
				return
			}

			id := safeString(vid["videoId"])
			title := safeString(dig(vid, "title", "runs", 0, "text"))
			thumb := getThumbnailURL(vid)
			durationText := safeString(dig(vid, "lengthText", "simpleText"))

			if durationText == "" {
				return
			}

			duration := parseDuration(durationText)
			t := &state.Track{
				URL:      "https://www.youtube.com/watch?v=" + id,
				Title:    title,
				ID:       id,
				Artwork:  thumb,
				Duration: duration,
				Source:   PlatformYouTube,
			}
			*tracks = append(*tracks, t)
			youtubeCache.Set("track:"+t.ID, []*state.Track{t})
		} else {
			for _, child := range v {
				parseSearchResults(child, tracks, limit)
			}
		}
	}
}

func isLiveVideo(videoRenderer map[string]any) bool {
	if badges, ok := dig(videoRenderer, "badges").([]any); ok {
		for _, badge := range badges {
			if badgeMap, ok := badge.(map[string]any); ok {
				if metadataBadge, ok := dig(badgeMap, "metadataBadgeRenderer").(map[string]any); ok {

					style := safeString(metadataBadge["style"])
					label := safeString(metadataBadge["label"])

					if style == "BADGE_STYLE_TYPE_LIVE_NOW" || label == "LIVE" {
						return true
					}
				}
			}
		}
	}

	if viewCountText, ok := dig(videoRenderer, "viewCountText", "runs").([]any); ok {
		for _, run := range viewCountText {
			if runMap, ok := run.(map[string]any); ok {
				text := safeString(runMap["text"])
				if strings.Contains(strings.ToLower(text), "watching") {
					return true
				}
			}
		}
	}
	return false
}

func getThumbnailURL(vid map[string]any) string {
	thumbs, ok := dig(vid, "thumbnail", "thumbnails").([]any)
	if !ok || len(thumbs) == 0 {
		return ""
	}

	last := thumbs[len(thumbs)-1]
	if m, ok := last.(map[string]any); ok {
		return safeString(m["url"])
	}
	return ""
}

func dig(m any, path ...any) any {
	curr := m
	for _, p := range path {
		switch key := p.(type) {
		case string:
			if mm, ok := curr.(map[string]any); ok {
				curr = mm[key]
			} else {
				return nil
			}
		case int:
			if arr, ok := curr.([]any); ok && len(arr) > key {
				curr = arr[key]
			} else {
				return nil
			}
		}
	}
	return curr
}

func safeString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func parseDuration(s string) int {
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ":")
	total := 0
	multiplier := 1
	for i := len(parts) - 1; i >= 0; i-- {
		total += atoi(parts[i]) * multiplier
		multiplier *= 60
	}
	return total
}

func atoi(s string) int {
	var n int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		}
	}
	return n
}

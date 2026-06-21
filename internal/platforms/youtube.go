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
	"regexp"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/config"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/utils"
)

const (
	PlatformYouTube        state.PlatformName = "YouTube"
	innerTubeKey                              = "AIzaSyBOti4mM-6x9WDnZIjIeyEU21OpBXqWBgw"
	innerTubeClientVersion                    = "2.20250101.01.00"
	innerTubeClientName                       = "WEB"
)

type YouTubePlatform struct {
	cache *utils.Cache[string, []*state.Track]
}

var (
	youtubeLinkRe = regexp.MustCompile(
		`(?i)^(?:https?:\/\/)?(?:www\.|m\.|music\.)?(?:youtube\.com|youtu\.be)\/\S+`,
	)
	videoIDRe1 = regexp.MustCompile(
		`(?i)(?:youtube\.com/(?:watch\?v=|embed/|shorts/|live/)|youtu\.be/)([A-Za-z0-9_-]{11})`,
	)
	videoIDRe2    = regexp.MustCompile(`(?:v=|\/)([0-9A-Za-z_-]{11})`)
	playlistIDRe1 = regexp.MustCompile(
		`(?i)(?:youtube\.com|music\.youtube\.com).*(?:\?|&)list=([A-Za-z0-9_-]+)`,
	)
	playlistIDRe2 = regexp.MustCompile(`list=([0-9A-Za-z_-]+)`)
)

func init() {
	Register(&YouTubePlatform{
		cache: utils.NewCache[string, []*state.Track](1 * time.Hour),
	})
}

func (p *YouTubePlatform) Name() state.PlatformName { return PlatformYouTube }
func (p *YouTubePlatform) Priority() int            { return 90 }

func (p *YouTubePlatform) CanGet(query string) bool {
	return youtubeLinkRe.MatchString(query)
}

func (p *YouTubePlatform) Get(input string, video bool) ([]*state.Track, error) {
	query := strings.TrimSpace(input)
	if query == "" {
		return nil, errors.New("empty query")
	}

	var (
		tracks []*state.Track
		err    error
	)

	if !youtubeLinkRe.MatchString(query) {
		tracks, err = p.VideoSearch(query)
	} else {
		playlistID := p.extractPlaylistID(query)
		videoID := p.extractVideoID(query)

		switch {
		case playlistID != "" && videoID != "":
			tracks, err = p.handleCombined(query, videoID)
		case playlistID != "":
			tracks, err = p.handlePlaylist(query)
		default:
			tracks, err = p.handleTrackURL(query)
		}
	}

	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, errors.New("no tracks found")
	}

	return withVideo(tracks, video), nil
}

func (p *YouTubePlatform) CanDownload(_ state.PlatformName) bool { return false }

func (p *YouTubePlatform) Download(_ context.Context, _ *state.Track, _ *telegram.NewMessage) (string, error) {
	return "", errors.New("youtube platform does not support downloading")
}

// VideoSearch is exported for Spotify to use.
func (p *YouTubePlatform) VideoSearch(query string, single ...bool) ([]*state.Track, error) {
	limit := config.QueueLimit
	onlyOne := len(single) > 0 && single[0]
	if onlyOne {
		limit = 1
	}

	cacheKey := "search:" + strings.ToLower(strings.TrimSpace(query))
	if arr, ok := p.cache.Get(cacheKey); ok {
		if onlyOne && len(arr) > 0 {
			return []*state.Track{arr[0]}, nil
		}
		if !onlyOne && len(arr) > 1 {
			return arr, nil
		}
	}

	tracks, err := p.performSearch(query, limit)
	if err != nil {
		return nil, fmt.Errorf("ytsearch failed: %w", err)
	}
	if len(tracks) == 0 {
		return nil, errors.New("no tracks found")
	}

	p.cache.Set(cacheKey, tracks)

	if onlyOne {
		return []*state.Track{tracks[0]}, nil
	}
	return tracks, nil
}

func (p *YouTubePlatform) handlePlaylist(rawURL string) ([]*state.Track, error) {
	cacheKey := "playlist:" + strings.ToLower(rawURL)
	if cached, ok := p.cache.Get(cacheKey); ok {
		return cached, nil
	}

	playlistID := p.extractPlaylistID(rawURL)
	if playlistID == "" {
		return nil, errors.New("invalid playlist url")
	}

	var (
		tracks []*state.Track
		err    error
	)

	if strings.HasPrefix(playlistID, "RD") {
		tracks, err = p.fetchMixPlaylist(playlistID, config.QueueLimit)
	} else {
		tracks, err = p.fetchPlaylist(playlistID, config.QueueLimit)
	}

	if err != nil {
		return nil, fmt.Errorf("playlist fetch failed: %w", err)
	}

	if len(tracks) > 0 {
		p.cache.Set(cacheKey, tracks)
	}
	return tracks, nil
}

func (p *YouTubePlatform) handleCombined(rawURL, videoID string) ([]*state.Track, error) {
	vTracks, vErr := p.handleTrackURL(rawURL)
	pTracks, pErr := p.handlePlaylist(rawURL)

	if vErr == nil && pErr == nil && len(vTracks) > 0 {
		vid := vTracks[0].ID
		out := []*state.Track{vTracks[0]}
		for _, t := range pTracks {
			if t.ID != vid {
				out = append(out, t)
			}
		}
		return out, nil
	}
	if vErr == nil {
		return vTracks, nil
	}
	if pErr == nil {
		gologging.WarnF("[YouTube] video fetch failed for %s: %v", videoID, vErr)
		return pTracks, nil
	}
	return nil, fmt.Errorf("video (%v) and playlist (%v) both failed", vErr, pErr)
}

func (p *YouTubePlatform) handleTrackURL(rawURL string) ([]*state.Track, error) {
	videoID := p.extractVideoID(rawURL)
	if videoID == "" {
		return nil, errors.New("invalid video url")
	}

	if cached, ok := p.cache.Get("track:" + videoID); ok && len(cached) > 0 {
		return cached, nil
	}

	track, err := p.fetchVideo(videoID)
	if err == nil && track != nil {
		p.cache.Set("track:"+videoID, []*state.Track{track})
		return []*state.Track{track}, nil
	}

	for _, q := range []string{videoID, rawURL} {
		results, err := p.VideoSearch(q)
		if err != nil {
			continue
		}
		for _, t := range results {
			if t.ID == videoID {
				p.cache.Set("track:"+videoID, []*state.Track{t})
				return []*state.Track{t}, nil
			}
		}
	}

	return nil, errors.New("track not found")
}

func (p *YouTubePlatform) extractPlaylistID(input string) string {
	if m := playlistIDRe1.FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}
	if m := playlistIDRe2.FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}
	return ""
}

func (p *YouTubePlatform) extractVideoID(u string) string {
	if m := videoIDRe1.FindStringSubmatch(u); len(m) > 1 {
		return m[1]
	}
	if m := videoIDRe2.FindStringSubmatch(u); len(m) > 1 {
		return m[1]
	}
	return ""
}

func (p *YouTubePlatform) performSearch(query string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] search: %s", query)
	var result map[string]any

	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    innerTubeClientName,
				"clientVersion": innerTubeClientVersion,
				"hl":            "en-IN",
				"gl":            "IN",
			},
		},
		"query":  query,
		"params": "CAASAhAB",
	}

	if err := p.callInnerTube("search", payload, &result); err != nil {
		return nil, err
	}

	contents, ok := dig(
		result,
		"contents", "twoColumnSearchResultsRenderer",
		"primaryContents", "sectionListRenderer", "contents",
	).([]any)
	if !ok {
		return nil, errors.New("invalid search results structure")
	}

	var tracks []*state.Track
	p.parseNodes(contents, &tracks, limit, "videoRenderer")
	return tracks, nil
}

func (p *YouTubePlatform) fetchVideo(videoID string) (*state.Track, error) {
	gologging.DebugF("[YouTube] fetchVideo: %s", videoID)
	var result map[string]any

	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    innerTubeClientName,
				"clientVersion": innerTubeClientVersion,
			},
		},
		"videoId": videoID,
	}

	if err := p.callInnerTube("player", payload, &result); err != nil {
		return nil, err
	}

	details, ok := dig(result, "videoDetails").(map[string]any)
	if !ok {
		return nil, errors.New("videoDetails not found")
	}

	id := safeStr(details["videoId"])
	return &state.Track{
		URL:      "https://www.youtube.com/watch?v=" + id,
		Title:    safeStr(details["title"]),
		ID:       id,
		Artwork:  getThumbnailURL(result),
		Duration: atoi(safeStr(details["lengthSeconds"])),
		Source:   PlatformYouTube,
	}, nil
}

func (p *YouTubePlatform) fetchPlaylist(playlistID string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] fetchPlaylist: %s", playlistID)
	var result map[string]any

	browseID := playlistID
	if !strings.HasPrefix(playlistID, "VL") {
		browseID = "VL" + playlistID
	}

	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    innerTubeClientName,
				"clientVersion": innerTubeClientVersion,
			},
		},
		"browseId": browseID,
	}

	if err := p.callInnerTube("browse", payload, &result); err != nil {
		return nil, err
	}

	var tracks []*state.Track
	p.parseNodes(result, &tracks, limit, "playlistVideoRenderer")
	return tracks, nil
}

func (p *YouTubePlatform) fetchMixPlaylist(playlistID string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] fetchMix: %s", playlistID)
	var result map[string]any

	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    innerTubeClientName,
				"clientVersion": innerTubeClientVersion,
			},
		},
		"playlistId": playlistID,
	}

	if err := p.callInnerTube("next", payload, &result); err != nil {
		return nil, err
	}

	items, ok := dig(
		result,
		"contents", "twoColumnWatchNextResults",
		"playlist", "playlist", "contents",
	).([]any)
	if !ok {
		return nil, errors.New("mix contents not found")
	}

	var tracks []*state.Track
	for _, item := range items {
		if limit > 0 && len(tracks) >= limit {
			break
		}
		vid, ok := dig(item, "playlistPanelVideoRenderer").(map[string]any)
		if !ok {
			continue
		}
		id := safeStr(vid["videoId"])
		if id == "" {
			continue
		}
		t := &state.Track{
			URL:      "https://www.youtube.com/watch?v=" + id,
			Title:    safeStr(dig(vid, "title", "simpleText")),
			ID:       id,
			Artwork:  getThumbnailURL(vid),
			Duration: parseDuration(safeStr(dig(vid, "lengthText", "simpleText"))),
			Source:   PlatformYouTube,
		}
		tracks = append(tracks, t)
		p.cache.Set("track:"+id, []*state.Track{t})
	}

	return tracks, nil
}

func (p *YouTubePlatform) callInnerTube(endpoint string, body, result any) error {
	apiURL := fmt.Sprintf(
		"https://m.youtube.com/youtubei/v1/%s?key=%s",
		endpoint, innerTubeKey,
	)
	resp, err := rc.R().
		SetBody(body).
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36").
		Post(apiURL)
	if err != nil {
		return fmt.Errorf("innertube request failed: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("innertube error: %d", resp.StatusCode())
	}
	return nil
}

func (p *YouTubePlatform) parseNodes(node any, tracks *[]*state.Track, limit int, rendererKey string) {
	if limit > 0 && len(*tracks) >= limit {
		return
	}

	switch v := node.(type) {
	case []any:
		for _, item := range v {
			p.parseNodes(item, tracks, limit, rendererKey)
		}
	case map[string]any:
		if vid, ok := v[rendererKey].(map[string]any); ok {
			if rendererKey == "videoRenderer" && isLiveVideo(vid) {
				return
			}
			id := safeStr(vid["videoId"])
			if id == "" {
				return
			}
			durationText := safeStr(dig(vid, "lengthText", "simpleText"))
			if durationText == "" {
				return
			}
			t := &state.Track{
				URL:      "https://www.youtube.com/watch?v=" + id,
				Title:    safeStr(dig(vid, "title", "runs", 0, "text")),
				ID:       id,
				Artwork:  getThumbnailURL(vid),
				Duration: parseDuration(durationText),
				Source:   PlatformYouTube,
			}
			*tracks = append(*tracks, t)
			p.cache.Set("track:"+id, []*state.Track{t})
		} else {
			for _, val := range v {
				p.parseNodes(val, tracks, limit, rendererKey)
			}
		}
	}
}

func isLiveVideo(vid map[string]any) bool {
	if badges, ok := dig(vid, "badges").([]any); ok {
		for _, b := range badges {
			if safeStr(dig(b, "metadataBadgeRenderer", "style")) == "BADGE_STYLE_TYPE_LIVE_NOW" {
				return true
			}
		}
	}
	return strings.Contains(
		strings.ToLower(safeStr(dig(vid, "viewCountText", "runs", 0, "text"))),
		"watching",
	)
}

func getThumbnailURL(vid map[string]any) string {
	thumbs, ok := dig(vid, "thumbnail", "thumbnails").([]any)
	if !ok {
		thumbs, ok = dig(vid, "videoDetails", "thumbnail", "thumbnails").([]any)
	}
	if ok && len(thumbs) > 0 {
		if last, ok := thumbs[len(thumbs)-1].(map[string]any); ok {
			return safeStr(last["url"])
		}
	}
	return ""
}

func dig(m any, path ...any) any {
	curr := m
	for _, key := range path {
		switch k := key.(type) {
		case string:
			mm, ok := curr.(map[string]any)
			if !ok {
				return nil
			}
			curr = mm[k]
		case int:
			arr, ok := curr.([]any)
			if !ok || k >= len(arr) {
				return nil
			}
			curr = arr[k]
		}
	}
	return curr
}

func safeStr(v any) string {
	s, _ := v.(string)
	return s
}

func parseDuration(s string) int {
	parts := strings.Split(s, ":")
	total, mult := 0, 1
	for i := len(parts) - 1; i >= 0; i-- {
		total += atoi(parts[i]) * mult
		mult *= 60
	}
	return total
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		}
	}
	return n
}

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
	videoIDRe1 = regexp.MustCompile(
		`(?i)(?:youtube\.com/(?:watch\?v=|embed/|shorts/|live/)|youtu\.be/)([A-Za-z0-9_-]{11})`,
	)
	videoIDRe2    = regexp.MustCompile(`(?:v=|\/)([0-9A-Za-z_-]{11})`)
	playlistIDRe1 = regexp.MustCompile(
		`(?i)(?:youtube\.com|music\.youtube\.com).*(?:\?|&)list=([A-Za-z0-9_-]+)`,
	)
	playlistIDRe2 = regexp.MustCompile(`list=([0-9A-Za-z_-]+)`)
	youtubeCache  = utils.NewCache[string, []*state.Track](1 * time.Hour)
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

func (p *YouTubePlatform) Name() state.PlatformName {
	return p.name
}

func (p *YouTubePlatform) CanGetTracks(link string) bool {
	return youtubeLinkRegex.MatchString(link)
}

func (p *YouTubePlatform) GetTracks(
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
		playlistID := p.extractPlaylistID(trimmed)
		videoID := p.extractVideoID(trimmed)

		if playlistID != "" && videoID == "" {
			tracks, err = p.handlePlaylist(trimmed)
		} else {
			tracks, err = p.handleTrackURL(trimmed)
		}
	} else {
		tracks, err = p.VideoSearch(trimmed, false)
	}

	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, errors.New("no tracks found")
	}

	return updateCached(tracks, video), nil
}

func (p *YouTubePlatform) handlePlaylist(
	rawURL string,
) ([]*state.Track, error) {
	cacheKey := "playlist:" + strings.ToLower(rawURL)
	if cached, ok := youtubeCache.Get(cacheKey); ok {
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

	if err == nil && len(tracks) > 0 {
		youtubeCache.Set(cacheKey, tracks)
		return tracks, nil
	}

	return nil, fmt.Errorf("failed to fetch playlist: %w", err)
}

func (p *YouTubePlatform) handleTrackURL(
	rawURL string,
) ([]*state.Track, error) {
	videoID := p.extractVideoID(rawURL)
	if videoID == "" {
		return nil, errors.New("invalid video url")
	}

	if cached, ok := youtubeCache.Get("track:" + videoID); ok &&
		len(cached) > 0 {
		return cached, nil
	}

	track, err := p.fetchVideo(videoID)
	if err == nil && track != nil {
		youtubeCache.Set("track:"+videoID, []*state.Track{track})
		return []*state.Track{track}, nil
	}

	for _, query := range []string{videoID, rawURL} {
		results, err := p.VideoSearch(query, true)
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

func (p *YouTubePlatform) CanDownload(source state.PlatformName) bool {
	return false
}

func (p *YouTubePlatform) Download(
	_ context.Context,
	_ *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	return "", errors.New("youtube platform does not support downloading")
}

func (*YouTubePlatform) CanSearch() bool { return true }

func (p *YouTubePlatform) Search(
	q string,
	video bool,
) ([]*state.Track, error) {
	return p.GetTracks(q, video)
}

func (p *YouTubePlatform) VideoSearch(
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
		if !single && len(arr) > 1 {
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

	youtubeCache.Set(cacheKey, tracks)

	if single {
		return []*state.Track{tracks[0]}, nil
	}

	return tracks, nil
}

func (p *YouTubePlatform) extractPlaylistID(input string) string {
	m0 := playlistIDRe1.FindStringSubmatch(input)
	if len(m0) > 1 {
		return m0[1]
	}
	m := playlistIDRe2.FindStringSubmatch(input)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func (p *YouTubePlatform) extractVideoID(u string) string {
	m := videoIDRe1.FindStringSubmatch(u)
	if len(m) > 1 {
		return m[1]
	}
	m2 := videoIDRe2.FindStringSubmatch(u)
	if len(m2) > 1 {
		return m2[1]
	}
	return ""
}

func (p *YouTubePlatform) performSearch(query string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] Searching: %s", query)
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

	err := p.callInnerTube("search", payload, &result)
	if err != nil {
		return nil, err
	}

	contents, ok := dig(
		result,
		"contents",
		"twoColumnSearchResultsRenderer",
		"primaryContents",
		"sectionListRenderer",
		"contents",
	).([]any)

	if !ok {
		return nil, errors.New("invalid search results")
	}

	var tracks []*state.Track
	p.parseNodes(contents, &tracks, limit, "videoRenderer")
	return tracks, nil
}

func (p *YouTubePlatform) fetchVideo(videoID string) (*state.Track, error) {
	gologging.DebugF("[YouTube] Fetching video: %s", videoID)
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

	err := p.callInnerTube("player", payload, &result)
	if err != nil {
		return nil, err
	}

	details, ok := dig(result, "videoDetails").(map[string]any)
	if !ok {
		return nil, errors.New("video details not found")
	}

	id := safeString(details["videoId"])
	title := safeString(details["title"])
	duration := atoi(safeString(details["lengthSeconds"]))
	thumb := getThumbnailURL(result)

	return &state.Track{
		URL:      "https://www.youtube.com/watch?v=" + id,
		Title:    title,
		ID:       id,
		Artwork:  thumb,
		Duration: duration,
		Source:   PlatformYouTube,
	}, nil
}

func (p *YouTubePlatform) fetchPlaylist(playlistID string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] Fetching playlist: %s", playlistID)
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

	err := p.callInnerTube("browse", payload, &result)
	if err != nil {
		return nil, err
	}

	var tracks []*state.Track
	p.parseNodes(result, &tracks, limit, "playlistVideoRenderer")
	return tracks, nil
}

func (p *YouTubePlatform) fetchMixPlaylist(playlistID string, limit int) ([]*state.Track, error) {
	gologging.DebugF("[YouTube] Fetching mix: %s", playlistID)
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

	err := p.callInnerTube("next", payload, &result)
	if err != nil {
		return nil, err
	}

	items, ok := dig(
		result,
		"contents",
		"twoColumnWatchNextResults",
		"playlist",
		"playlist",
		"contents",
	).([]any)

	if !ok {
		return nil, errors.New("mix contents not found")
	}

	var tracks []*state.Track
	for _, item := range items {
		if limit > 0 && len(tracks) >= limit {
			break
		}

		if vid, ok := dig(item, "playlistPanelVideoRenderer").(map[string]any); ok {
			id := safeString(vid["videoId"])
			if id == "" {
				continue
			}

			title := safeString(dig(vid, "title", "simpleText"))
			thumb := getThumbnailURL(vid)
			duration := parseDuration(safeString(dig(vid, "lengthText", "simpleText")))

			t := &state.Track{
				URL:      "https://www.youtube.com/watch?v=" + id,
				Title:    title,
				ID:       id,
				Artwork:  thumb,
				Duration: duration,
				Source:   PlatformYouTube,
			}
			tracks = append(tracks, t)
			youtubeCache.Set("track:"+id, []*state.Track{t})
		}
	}

	return tracks, nil
}

func (p *YouTubePlatform) callInnerTube(endpoint string, body any, result any) error {
	apiURL := fmt.Sprintf("https://m.youtube.com/youtubei/v1/%s?key=%s", endpoint, innerTubeKey)
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

			id := safeString(vid["videoId"])
			if id == "" {
				return
			}

			title := safeString(dig(vid, "title", "runs", 0, "text"))
			thumb := getThumbnailURL(vid)
			durationText := safeString(dig(vid, "lengthText", "simpleText"))
			if durationText == "" {
				return
			}

			t := &state.Track{
				URL:      "https://www.youtube.com/watch?v=" + id,
				Title:    title,
				ID:       id,
				Artwork:  thumb,
				Duration: parseDuration(durationText),
				Source:   PlatformYouTube,
			}
			*tracks = append(*tracks, t)
			youtubeCache.Set("track:"+id, []*state.Track{t})
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
			if style := safeString(dig(b, "metadataBadgeRenderer", "style")); style == "BADGE_STYLE_TYPE_LIVE_NOW" {
				return true
			}
		}
	}
	return strings.Contains(strings.ToLower(safeString(dig(vid, "viewCountText", "runs", 0, "text"))), "watching")
}

func getThumbnailURL(vid map[string]any) string {
	thumbs, ok := dig(vid, "thumbnail", "thumbnails").([]any)
	if !ok || len(thumbs) == 0 {
		// Try player response structure
		thumbs, ok = dig(vid, "videoDetails", "thumbnail", "thumbnails").([]any)
	}

	if ok && len(thumbs) > 0 {
		if last, ok := thumbs[len(thumbs)-1].(map[string]any); ok {
			return safeString(last["url"])
		}
	}
	return ""
}

func updateCached(tracks []*state.Track, video bool) []*state.Track {
	out := make([]*state.Track, 0, len(tracks))
	for _, t := range tracks {
		if t == nil {
			continue
		}
		tc := *t
		tc.Video = video
		out = append(out, &tc)
	}
	return out
}

func dig(m any, path ...any) any {
	curr := m
	for _, p := range path {
		switch k := p.(type) {
		case string:
			if mm, ok := curr.(map[string]any); ok {
				curr = mm[k]
			} else {
				return nil
			}
		case int:
			if arr, ok := curr.([]any); ok && k < len(arr) {
				curr = arr[k]
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
	parts := strings.Split(s, ":")
	total := 0
	mult := 1
	for i := len(parts) - 1; i >= 0; i-- {
		n := atoi(parts[i])
		total += n * mult
		mult *= 60
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

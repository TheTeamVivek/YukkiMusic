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
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"

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
	youtube = &YouTubePlatform{
		cache: utils.NewCache[string, []*state.Track](time.Hour),
	}

	errLiveStream = errors.New("live streams are not supported")

	youtubeLinkRe = regexp.MustCompile(
		`(?i)^(?:https?:\/\/)?(?:www\.|m\.|music\.)?(?:youtube\.com|youtu\.be)\/\S+`,
	)
	videoIDRe = regexp.MustCompile(
		`(?i)(?:youtu\.be/|youtube\.com/(?:embed|shorts|live|v|e|vi)/|[?&]v=)([A-Za-z0-9_-]{11})`,
	)
	playlistIDRe = regexp.MustCompile(`[?&]list=([A-Za-z0-9_-]+)`)
)

func init() {
	Register(youtube)
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
		playlistID, videoID := extractPlaylistID(query), extractVideoID(query)

		switch {
		case playlistID != "" && videoID != "":
			tracks, err = p.handleCombined(query)
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
		return nil, ErrNoTracks
	}

	return withVideo(tracks, video), nil
}

func (p *YouTubePlatform) CanDownload(_ state.PlatformName) bool { return false }

func (p *YouTubePlatform) Download(_ context.Context, _ *state.Track, _ *td.Message) (string, error) {
	return "", errors.New("youtube platform does not support downloading")
}

func withVideo(tracks []*state.Track, video bool) []*state.Track {
	out := make([]*state.Track, 0, len(tracks))
	for _, t := range tracks {
		clone := *t
		clone.Video = video
		out = append(out, &clone)
	}
	return out
}

func ytTrack(id, title, artwork string, duration int) *state.Track {
	return &state.Track{
		ID:       id,
		Title:    title,
		Artwork:  artwork,
		Duration: duration,
		URL:      "https://www.youtube.com/watch?v=" + id,
		Source:   PlatformYouTube,
	}
}

// remember caches a track under its id so a later link to it skips the lookup.
func (p *YouTubePlatform) remember(t *state.Track) []*state.Track {
	out := []*state.Track{t}
	p.cache.Set("track:"+t.ID, out)
	return out
}

func (p *YouTubePlatform) VideoSearch(query string) ([]*state.Track, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("empty search query")
	}

	key := "search:" + strings.ToLower(query)
	if tracks, ok := p.cache.Get(key); ok && len(tracks) > 0 {
		return tracks, nil
	}

	limit := config.QueueLimit
	if limit <= 0 {
		limit = 10
	}

	tracks, err := p.performSearch(query, limit)
	if err != nil {
		return nil, fmt.Errorf("ytsearch failed: %w", err)
	}
	if len(tracks) == 0 {
		return nil, ErrNoTracks
	}

	p.cache.Set(key, tracks)
	return tracks, nil
}

func (p *YouTubePlatform) handlePlaylist(rawURL string) ([]*state.Track, error) {
	id := extractPlaylistID(rawURL)
	if id == "" {
		return nil, errors.New("invalid playlist url")
	}

	key := "playlist:" + id
	if cached, ok := p.cache.Get(key); ok {
		return cached, nil
	}

	limit := config.QueueLimit
	if limit <= 0 {
		limit = 24
	}

	fetch := p.fetchPlaylist
	if strings.HasPrefix(id, "RD") {
		fetch = p.fetchMixPlaylist
	}

	tracks, err := fetch(id, limit)
	if err != nil {
		return nil, fmt.Errorf("playlist fetch failed: %w", err)
	}

	if len(tracks) > 0 {
		p.cache.Set(key, tracks)
	}
	return tracks, nil
}

// handleCombined serves a watch link that also carries a list: the video
// first, then the rest of the playlist.
func (p *YouTubePlatform) handleCombined(rawURL string) ([]*state.Track, error) {
	vTracks, vErr := p.handleTrackURL(rawURL)
	pTracks, pErr := p.handlePlaylist(rawURL)

	switch {
	case vErr == nil && pErr == nil && len(vTracks) > 0:
		out := []*state.Track{vTracks[0]}
		for _, t := range pTracks {
			if t.ID != out[0].ID {
				out = append(out, t)
			}
		}
		return out, nil
	case vErr == nil:
		return vTracks, nil
	case pErr == nil:
		utils.Log.Warnf("[YouTube] video fetch failed for %s: %v", rawURL, vErr)
		return pTracks, nil
	}
	return nil, fmt.Errorf("video (%v) and playlist (%v) both failed", vErr, pErr)
}

func (p *YouTubePlatform) handleTrackURL(rawURL string) ([]*state.Track, error) {
	id := extractVideoID(rawURL)
	if id == "" {
		return nil, errors.New("invalid video url")
	}

	if cached, ok := p.cache.Get("track:" + id); ok && len(cached) > 0 {
		return cached, nil
	}

	track, err := p.fetchVideo(id)
	if err == nil {
		return p.remember(track), nil
	}
	if errors.Is(err, errLiveStream) {
		return nil, err
	}

	if t := p.searchByID(id, id, rawURL); t != nil {
		return p.remember(t), nil
	}

	if title, err := oembedTitle(id); err == nil {
		queries := []string{title}
		if runes := []rune(title); len(runes) > 35 {
			queries = append(queries, string(runes[:35]))
		}
		if t := p.searchByID(id, queries...); t != nil {
			return p.remember(t), nil
		}
	}

	return nil, errors.New("track not found")
}

// searchByID runs the queries in turn and returns the result with the given video id.
func (p *YouTubePlatform) searchByID(id string, queries ...string) *state.Track {
	for _, q := range queries {
		results, err := p.VideoSearch(q)
		if err != nil {
			continue
		}
		for _, t := range results {
			if t.ID == id {
				return t
			}
		}
	}
	return nil
}

func oembedTitle(videoID string) (string, error) {
	var result map[string]any

	resp, err := rc.R().SetResult(&result).Get(
		"https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v=" + videoID + "&format=json",
	)
	if err != nil {
		return "", fmt.Errorf("oembed request failed: %w", err)
	}
	if resp.IsStatusFailure() {
		return "", fmt.Errorf("oembed error: %d", resp.StatusCode())
	}

	title := safeStr(result["title"])
	if title == "" {
		return "", errors.New("oembed: title not found")
	}
	return title, nil
}

func extractPlaylistID(input string) string {
	if m := playlistIDRe.FindStringSubmatch(input); m != nil {
		return m[1]
	}
	return ""
}

func extractVideoID(input string) string {
	if m := videoIDRe.FindStringSubmatch(input); m != nil {
		return m[1]
	}
	return ""
}

func (p *YouTubePlatform) performSearch(query string, limit int) ([]*state.Track, error) {
	var result map[string]any

	err := p.callInnerTube(
		"search",
		map[string]any{"hl": "en-IN", "gl": "IN"},
		map[string]any{"query": query, "params": "CAASAhAB"},
		&result,
	)
	if err != nil {
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
	var result map[string]any

	err := p.callInnerTube("player", nil, map[string]any{"videoId": videoID}, &result)
	if err != nil {
		return nil, err
	}

	details, ok := dig(result, "videoDetails").(map[string]any)
	if !ok {
		return nil, errors.New("videoDetails not found")
	}
	if live, _ := details["isLive"].(bool); live {
		return nil, errLiveStream
	}

	return ytTrack(
		safeStr(details["videoId"]),
		safeStr(details["title"]),
		thumbnailURL(result),
		atoi(safeStr(details["lengthSeconds"])),
	), nil
}

func (p *YouTubePlatform) fetchPlaylist(playlistID string, limit int) ([]*state.Track, error) {
	var result map[string]any

	browseID := playlistID
	if !strings.HasPrefix(playlistID, "VL") {
		browseID = "VL" + playlistID
	}

	err := p.callInnerTube("browse", nil, map[string]any{"browseId": browseID}, &result)
	if err != nil {
		return nil, err
	}

	var tracks []*state.Track
	p.parseNodes(result, &tracks, limit, "playlistVideoRenderer")
	return tracks, nil
}

func (p *YouTubePlatform) fetchMixPlaylist(playlistID string, limit int) ([]*state.Track, error) {
	var result map[string]any

	err := p.callInnerTube("next", nil, map[string]any{"playlistId": playlistID}, &result)
	if err != nil {
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
	p.parseNodes(items, &tracks, limit, "playlistPanelVideoRenderer")
	return tracks, nil
}

// callInnerTube posts fields to an InnerTube endpoint; client adds to the
// default client context.
func (p *YouTubePlatform) callInnerTube(endpoint string, client, fields map[string]any, result any) error {
	clientCtx := map[string]any{
		"clientName":    innerTubeClientName,
		"clientVersion": innerTubeClientVersion,
	}
	maps.Copy(clientCtx, client)

	body := map[string]any{"context": map[string]any{"client": clientCtx}}
	maps.Copy(body, fields)

	resp, err := rc.R().
		SetBody(body).
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36").
		Post(fmt.Sprintf("https://m.youtube.com/youtubei/v1/%s?key=%s", endpoint, innerTubeKey))
	if err != nil {
		return fmt.Errorf("innertube request failed: %w", err)
	}
	if resp.IsStatusFailure() {
		return fmt.Errorf("innertube error: %d", resp.StatusCode())
	}
	return nil
}

// parseNodes collects every rendererKey node under node, up to limit tracks.
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
		vid, ok := v[rendererKey].(map[string]any)
		if !ok {
			for _, val := range v {
				p.parseNodes(val, tracks, limit, rendererKey)
			}
			return
		}

		if t := parseRenderer(vid, rendererKey); t != nil {
			*tracks = append(*tracks, t)
			p.remember(t)
		}
	}
}

func parseRenderer(vid map[string]any, rendererKey string) *state.Track {
	if rendererKey == "videoRenderer" && isLiveVideo(vid) {
		return nil
	}

	id := safeStr(vid["videoId"])
	duration := extractText(dig(vid, "lengthText"))
	if id == "" || duration == "" {
		return nil
	}

	title := cmp.Or(extractText(dig(vid, "title")), "YouTube Track")
	return ytTrack(id, title, thumbnailURL(vid), parseDuration(duration))
}

func extractText(node any) string {
	switch v := node.(type) {
	case string:
		return v
	case map[string]any:
		if s := cmp.Or(safeStr(v["simpleText"]), safeStr(v["text"])); s != "" {
			return s
		}
		runs, _ := v["runs"].([]any)
		var sb strings.Builder
		for _, r := range runs {
			sb.WriteString(safeStr(dig(r, "text")))
		}
		return sb.String()
	}
	return ""
}

func isLiveVideo(vid map[string]any) bool {
	badges, _ := dig(vid, "badges").([]any)
	for _, b := range badges {
		if safeStr(dig(b, "metadataBadgeRenderer", "style")) == "BADGE_STYLE_TYPE_LIVE_NOW" {
			return true
		}
	}
	viewCount := strings.ToLower(extractText(dig(vid, "viewCountText")))
	return strings.Contains(viewCount, "watching")
}

func thumbnailURL(vid map[string]any) string {
	thumbs, ok := dig(vid, "thumbnail", "thumbnails").([]any)
	if !ok {
		thumbs, _ = dig(vid, "videoDetails", "thumbnail", "thumbnails").([]any)
	}
	if len(thumbs) == 0 {
		return ""
	}
	return safeStr(dig(thumbs[len(thumbs)-1], "url"))
}

func dig(m any, path ...string) any {
	for _, key := range path {
		mm, ok := m.(map[string]any)
		if !ok {
			return nil
		}
		m = mm[key]
	}
	return m
}

func safeStr(v any) string {
	s, _ := v.(string)
	return s
}

func parseDuration(s string) int {
	total, mult := 0, 1
	for _, part := range slices.Backward(strings.Split(strings.TrimSpace(s), ":")) {
		total += atoi(part) * mult
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

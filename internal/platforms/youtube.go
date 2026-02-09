/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
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

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	state "main/internal/core/models"
	"main/internal/utils"
)

type YouTubePlatform struct {
	name state.PlatformName
}

var (
	playlistRegex    = regexp.MustCompile(`(?i)(?:list=)([A-Za-z0-9_-]+)`)
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

	if youtubeLinkRegex.MatchString(trimmed) {

		if playlistRegex.MatchString(trimmed) {
			cacheKey := "playlist:" + strings.ToLower(trimmed)
			if cached, ok := youtubeCache.Get(cacheKey); ok {
				return updateCached(cached, video), nil
			}

			videoIDs, err := getPlaylist(trimmed)
			if err != nil {
				return nil, err
			}

			var tracks []*state.Track
			for _, videoID := range videoIDs {
				if cached, ok := youtubeCache.Get("track:" + videoID); ok &&
					len(cached) > 0 {
					tracks = append(tracks, cached[0])
					continue
				}

				trackList, err := yp.VideoSearch(
					"https://youtube.com/watch?v="+videoID,
					true,
				)
				if err != nil || len(trackList) == 0 {
					continue
				}

				t := trackList[0]
				youtubeCache.Set("track:"+videoID, []*state.Track{t})
				tracks = append(tracks, t)
			}

			if len(tracks) > 0 {
				youtubeCache.Set(cacheKey, tracks)
			}

			return updateCached(tracks, video), nil
		}

		normalizedURL, videoID, err := yp.normalizeYouTubeURL(trimmed)
		if err != nil {
			return nil, err
		}

		if cached, ok := youtubeCache.Get("track:" + videoID); ok &&
			len(cached) > 0 {
			return updateCached(cached, video), nil
		}

		trackList, err := yp.VideoSearch(normalizedURL, true)
		if err != nil {
			return nil, err
		}
		if len(trackList) == 0 {
			return nil, errors.New("track not found for the given url")
		}

		youtubeCache.Set("track:"+videoID, trackList)
		return updateCached(trackList, video), nil
	}

	tracks, err := yp.VideoSearch(trimmed, true)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, errors.New("no tracks found for the given query")
	}

	return updateCached(tracks, video), nil
}

func (yp *YouTubePlatform) CanDownload(source state.PlatformName) bool {
	return false
}

func (yt *YouTubePlatform) Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	return "", errors.New("youtube platform does not support downloading")
}

func (yp *YouTubePlatform) CanGetRecommendations() bool {
	return true
}

func (yp *YouTubePlatform) GetRecommendations(
	track *state.Track,
	hl, gl string,
) ([]*state.Track, error) {
	if hl == "" {
		hl = "en"
	}
	if gl == "" {
		gl = "IN"
	}

	nextURL := "https://m.youtube.com/youtubei/v1/next?key=" + innerTubeKey
	var result map[string]any

	resp, err := rc.R().
		SetResult(&result).
		SetBody(map[string]any{
			"context": map[string]any{
				"client": map[string]any{
					"clientName":    innerTubeClientName,
					"clientVersion": innerTubeClientVersion,
					"hl":            hl,
					"gl":            gl,
				},
			},
			"videoId": track.ID,
		}).
		SetHeader("Content-Type", "application/json").
		Post(nextURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	contents := dig(
		result,
		"contents",
		"twoColumnWatchNextResults",
		"secondaryResults",
		"secondaryResults",
		"results",
	)
	if contents == nil {
		return nil, fmt.Errorf("no contents found")
	}

	var tracks []*state.Track
	yp.parseNextResults(contents, &tracks, track.Video, track.Requester)

	if len(tracks) == 0 {
		return nil, errors.New("no recommendations found")
	}

	return tracks, nil
}

func (yp *YouTubePlatform) parseNextResults(
	node any,
	tracks *[]*state.Track,
	video bool,
	requester string,
) {
	switch v := node.(type) {
	case []any:
		for _, item := range v {
			yp.parseNextResults(item, tracks, video, requester)
		}
	case map[string]any:
		if _, ok := v["continuationItemRenderer"]; ok {
			return
		}
		if _, ok := v["itemSectionRenderer"]; ok {
			return
		}

		if lockup, ok := dig(v, "lockupViewModel").(map[string]any); ok {
			contentType := safeString(lockup["contentType"])
			if contentType != "LOCKUP_CONTENT_TYPE_VIDEO" {
				return
			}

			id := safeString(lockup["contentId"])
			if id == "" {
				return
			}

			meta := dig(lockup, "metadata", "lockupMetadataViewModel")
			title := safeString(dig(meta, "title", "content"))

			thumbVM := dig(lockup, "contentImage", "thumbnailViewModel")
			sources, _ := dig(thumbVM, "image", "sources").([]any)
			thumb := ""
			if len(sources) > 0 {
				if lastSource, ok := sources[len(sources)-1].(map[string]any); ok {
					thumb = safeString(lastSource["url"])
				}
			}

			durationText := ""
			overlays, _ := dig(thumbVM, "overlays").([]any)
			for _, overlay := range overlays {
				if overlayMap, ok := overlay.(map[string]any); ok {
					if badgeVM, ok := dig(overlayMap, "thumbnailOverlayBadgeViewModel").(map[string]any); ok {
						badges, _ := dig(badgeVM, "thumbnailBadges").([]any)
						if len(badges) > 0 {
							if badge, ok := badges[0].(map[string]any); ok {
								if badgeData, ok := dig(badge, "thumbnailBadgeViewModel").(map[string]any); ok {
									durationText = safeString(badgeData["text"])
									break
								}
							}
						}
					}
				}
			}

			if durationText == "" || id == "" {
				return
			}

			duration := parseDuration(durationText)
			t := &state.Track{
				URL:       "https://www.youtube.com/watch?v=" + id,
				Title:     title,
				ID:        id,
				Artwork:   thumb,
				Duration:  duration,
				Source:    PlatformYouTube,
				Video:     video,
				Requester: requester,
			}
			*tracks = append(*tracks, t)
			youtubeCache.Set("track:"+t.ID, []*state.Track{t})
		} else if vid, ok := dig(v, "compactVideoRenderer").(map[string]any); ok {
			id := safeString(vid["videoId"])
			title := safeString(dig(vid, "title", "simpleText"))
			if title == "" {
				title = safeString(dig(vid, "title", "runs", 0, "text"))
			}
			thumb := getThumbnailURL(vid)
			durationText := safeString(dig(vid, "lengthText", "simpleText"))

			if durationText == "" || id == "" {
				return
			}

			duration := parseDuration(durationText)
			t := &state.Track{
				URL:       "https://www.youtube.com/watch?v=" + id,
				Title:     title,
				ID:        id,
				Artwork:   thumb,
				Duration:  duration,
				Source:    PlatformYouTube,
				Video:     video,
				Requester: requester,
			}
			*tracks = append(*tracks, t)
			youtubeCache.Set("track:"+t.ID, []*state.Track{t})
		} else {
			for _, child := range v {
				yp.parseNextResults(child, tracks, video, requester)
			}
		}
	}
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
	if len(singleOpt) > 0 && singleOpt[0] {
		single = true
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

	tracks, err = searchYouTube(query)
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

func (yt *YouTubePlatform) normalizeYouTubeURL(
	input string,
) (string, string, error) {
	u, err := url.Parse(strings.TrimSpace(input))
	if err != nil {
		return "", "", err
	}

	host := strings.ToLower(u.Host)
	path := strings.Trim(u.Path, "/")

	if strings.Contains(host, "youtu.be") {
		id := strings.Split(path, "/")[0]
		if len(id) == 11 {
			return "https://www.youtube.com/watch?v=" + id, id, nil
		}
	}

	if strings.Contains(host, "youtube.com") {
		if v := u.Query().Get("v"); len(v) == 11 {
			return "https://www.youtube.com/watch?v=" + v, v, nil
		}

		parts := strings.Split(path, "/")

		if len(parts) >= 2 && parts[0] == "shorts" && len(parts[1]) == 11 {
			return "https://www.youtube.com/watch?v=" + parts[1], parts[1], nil
		}

		if len(parts) >= 3 && parts[0] == "source" && len(parts[1]) == 11 {
			return "https://www.youtube.com/watch?v=" + parts[1], parts[1], nil
		}

		if len(parts) >= 2 && parts[0] == "embed" && len(parts[1]) == 11 {
			return "https://www.youtube.com/watch?v=" + parts[1], parts[1], nil
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

func updateCached(arr []*state.Track, video bool) []*state.Track {
	if len(arr) == 0 {
		return nil
	}
	out := make([]*state.Track, len(arr))
	for i, t := range arr {
		if t == nil {
			continue
		}
		clone := *t
		clone.Video = video
		out[i] = &clone
	}
	return out
}

// The following search functions are adapted from TgMusicBot.
// Copyright (c) 2025 Ashok Shau
// Licensed under GNU GPL v3
// See https://github.com/AshokShau/TgMusicBot
//
// searchYouTube scrapes YouTube results page

func searchYouTube(query string) ([]*state.Track, error) {
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
	parseSearchResults(contents, &tracks)
	return tracks, nil
}

func parseSearchResults(node any, tracks *[]*state.Track) {
	switch v := node.(type) {

	case []any:
		for _, item := range v {
			parseSearchResults(item, tracks)
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
				parseSearchResults(child, tracks)
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

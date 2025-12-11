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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/raitonoberu/ytsearch"
	"resty.dev/v3"

	"main/internal/config"
	"main/internal/core/models"
	"main/internal/utils"
)

type YouTubePlatform struct{}

var (
	videoIDRegex     = regexp.MustCompile(`(?i)(?:v=|\/v\/|\/embed\/|youtu\.be\/)([A-Za-z0-9_-]{11})`)
	playlistRegex    = regexp.MustCompile(`(?i)(?:list=)([A-Za-z0-9_-]+)`)
	youtubeLinkRegex = regexp.MustCompile(`(?i)^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.be|music\.youtube\.com)\/`)
	youtubeCache     = utils.NewCache[string, []*state.Track](1 * time.Hour)
)

const PlatformYouTube state.PlatformName = "YouTube"

func init() {
	addPlatform(90, PlatformYouTube, &YouTubePlatform{})
}

func (*YouTubePlatform) Name() state.PlatformName { return PlatformYouTube }
func (*YouTubePlatform) IsValid(link string) bool { return youtubeLinkRegex.MatchString(link) }

func (yp *YouTubePlatform) GetTracks(input string, video bool) ([]*state.Track, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, errors.New("empty query")
	}

	if youtubeLinkRegex.MatchString(trimmed) {
		// playlist URL
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
				if cached, ok := youtubeCache.Get("track:" + videoID); ok && len(cached) > 0 {
					tracks = append(tracks, cached[0])
					continue
				}

				trackList, err := yp.VideoSearch("https://youtube.com/watch?v="+videoID, true)
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

		// single video URL
		matches := videoIDRegex.FindStringSubmatch(trimmed)
		if len(matches) < 2 {
			return nil, errors.New("unsupported YouTube URL or missing video ID")
		}

		videoID := matches[1]
		if cached, ok := youtubeCache.Get("track:" + videoID); ok && len(cached) > 0 {
			return updateCached(cached, video), nil
		}

		trackList, err := yp.VideoSearch("https://youtube.com/watch?v="+videoID, true)
		if err != nil {
			return nil, err
		}
		if len(trackList) == 0 {
			return nil, errors.New("track not found for the given url")
		}

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

func (*YouTubePlatform) IsDownloadSupported(source state.PlatformName) bool {
	return false
}

func (yt *YouTubePlatform) Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
	return "", errors.New("youtube platform does not support downloading")
}

func (yp *YouTubePlatform) VideoSearch(query string, singleOpt ...bool) ([]*state.Track, error) {
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

	// Try scraping first
	tracks, err = searchYouTube(query)

	// If scraping failed or found no results, fallback to ytsearch
	if err != nil || len(tracks) == 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("Search failed: %v", r)
				}
			}()
			search := ytsearch.VideoSearch(query)
			for i := 0; i < 2; i++ {
				result, searchErr := search.Next()
				if searchErr != nil {
					err = searchErr
					return
				}
				if result == nil {
					continue
				}
				for _, v := range result.Videos {
					if len(v.Thumbnails) == 0 || v.Duration == 0 {
						continue
					}
					thumb := v.Thumbnails[len(v.Thumbnails)-1].URL
					t := &state.Track{
						ID:       v.ID,
						Title:    v.Title,
						Duration: v.Duration,
						Artwork:  thumb,
						URL:      v.URL,
						Source:   PlatformYouTube,
					}
					tracks = append(tracks, t)
					youtubeCache.Set("track:"+t.ID, []*state.Track{t})
					if single {
						break
					}
				}
				if single && len(tracks) > 0 {
					break
				}
			}
		}()
		if err != nil {
			return nil, fmt.Errorf("ytsearch failed: %w", err)
		}
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

func getPlaylist(pUrl string) ([]string, error) {
	if strings.Contains(pUrl, "&") {
		pUrl = strings.Split(pUrl, "&")[0]
	}

	cmd := exec.Command("yt-dlp", "-i", "--compat-options", "no-youtube-unavailable-videos", "--get-id", "--flat-playlist", "--skip-download", "--playlist-end", strconv.Itoa(config.QueueLimit), pUrl)
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
	client := resty.New().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36").
		SetHeader("Accept-Language", "en-US,en;q=0.9").
		SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	defer client.Close()

	encodedQuery := url.QueryEscape(query)
	searchURL := "https://www.youtube.com/results?search_query=" + encodedQuery

	resp, err := client.R().Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	body := resp.String()
	re := regexp.MustCompile(`var ytInitialData = (.*?);\s*</script>`)
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		return nil, fmt.Errorf("ytInitialData not found")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		return nil, err
	}

	contents := dig(data, "contents", "twoColumnSearchResultsRenderer",
		"primaryContents", "sectionListRenderer", "contents")
	if contents == nil {
		return nil, fmt.Errorf("no contents found")
	}

	var tracks []*state.Track
	parseSearchResults(contents, &tracks)
	return tracks, nil
}

func parseSearchResults(node interface{}, tracks *[]*state.Track) {
	switch v := node.(type) {
	case []interface{}:
		for _, item := range v {
			parseSearchResults(item, tracks)
		}
	case map[string]interface{}:
		if vid, ok := dig(v, "videoRenderer").(map[string]interface{}); ok {
			id := safeString(vid["videoId"])
			title := safeString(dig(vid, "title", "runs", 0, "text"))
			thumb := safeString(dig(vid, "thumbnail", "thumbnails", 0, "url"))
			durationText := safeString(dig(vid, "lengthText", "simpleText"))
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

func dig(m interface{}, path ...interface{}) interface{} {
	curr := m
	for _, p := range path {
		switch key := p.(type) {
		case string:
			if mm, ok := curr.(map[string]interface{}); ok {
				curr = mm[key]
			} else {
				return nil
			}
		case int:
			if arr, ok := curr.([]interface{}); ok && len(arr) > key {
				curr = arr[key]
			} else {
				return nil
			}
		}
	}
	return curr
}

func safeString(v interface{}) string {
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

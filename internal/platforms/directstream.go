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
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	td "github.com/AshokShau/gotdbot"

	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/utils"
)

const (
	PlatformDirectStream state.PlatformName = "DirectStream"

	streamUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

type DirectStreamPlatform struct {
	cache *utils.Cache[string, *streamInfo]
}

type streamInfo struct {
	IsVideo  bool
	Duration int
}

var (
	streamExtensions = []string{
		".m3u8", ".m3u", ".mp3", ".mp4", ".webm", ".ogg",
		".wav", ".flac", ".aac", ".opus", ".mkv", ".avi",
		".mov", ".ts", ".mpd",
	}
	streamPathRe    = regexp.MustCompile(`(?i)\.(?:m3u8|mpd)(?:\?|$)|/(?:hls|dash|stream)/`)
	streamMimeTypes = []string{
		"audio/", "video/",
		"application/vnd.apple.mpegurl",
		"application/x-mpegurl",
		"application/dash+xml",
		"application/octet-stream",
	}
	streamHosts = []string{"cdn", "stream", "media", "video", "audio"}
)

func init() {
	Register(&DirectStreamPlatform{
		cache: utils.NewCache[string, *streamInfo](5 * time.Minute),
	})
}

func (d *DirectStreamPlatform) Name() state.PlatformName { return PlatformDirectStream }
func (d *DirectStreamPlatform) Priority() int            { return 65 }

func (d *DirectStreamPlatform) CanGet(query string) bool {
	if _, err := sanitizeMediaURL(query); err != nil {
		return false
	}
	return looksLikeStream(query)
}

func (d *DirectStreamPlatform) Get(query string, video bool) ([]*state.Track, error) {
	safeURL, err := sanitizeMediaURL(query)
	if err != nil {
		return nil, err
	}

	info, err := d.lookup(safeURL)
	if err != nil {
		return nil, fmt.Errorf("stream validation failed: %w", err)
	}

	track := &state.Track{
		ID:       streamID(safeURL),
		Title:    streamTitle(safeURL),
		Duration: info.Duration,
		URL:      safeURL,
		Source:   PlatformDirectStream,
		Video:    video || info.IsVideo,
	}

	utils.Log.Infof("DirectStream: %s (video:%v dur:%d)", track.Title, track.Video, track.Duration)
	return []*state.Track{track}, nil
}

func (d *DirectStreamPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformDirectStream
}

func (d *DirectStreamPlatform) Download(
	_ context.Context,
	track *state.Track,
	_ *td.Message,
) (string, error) {
	utils.Log.Infof("DirectStream: returning URL for streaming: %s", track.URL)
	return track.URL, nil
}

func (d *DirectStreamPlatform) lookup(rawURL string) (*streamInfo, error) {
	if info, ok := d.cache.Get(rawURL); ok && info != nil {
		return info, nil
	}

	info, err := validateStream(rawURL)
	if err != nil {
		return nil, err
	}

	d.cache.Set(rawURL, info)
	return info, nil
}

func looksLikeStream(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if slices.Contains(streamExtensions, strings.ToLower(path.Ext(parsed.Path))) ||
		streamPathRe.MatchString(rawURL) {
		return true
	}

	host := strings.ToLower(parsed.Host)
	return slices.ContainsFunc(streamHosts, func(h string) bool {
		return strings.Contains(host, h)
	})
}

// validateStream probes the URL and checks that it serves audio or video.
func validateStream(rawURL string) (*streamInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Only the headers are read, so a server that ignores Range and streams
	// forever cannot stall the GET fallback.
	probe := func(method string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", streamUserAgent)
		if method == http.MethodGet {
			req.Header.Set("Range", "bytes=0-1024")
		}
		return http.DefaultClient.Do(req)
	}

	resp, err := probe(http.MethodHead)
	if err != nil || resp.StatusCode >= 400 {
		if err == nil {
			resp.Body.Close()
		}
		if resp, err = probe(http.MethodGet); err != nil {
			return nil, fmt.Errorf("failed to validate URL: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("URL returned status %d", resp.StatusCode)
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if !slices.ContainsFunc(streamMimeTypes, func(mt string) bool {
		return strings.HasPrefix(ct, mt)
	}) {
		return nil, fmt.Errorf("not audio/video (content-type: %s)", ct)
	}

	info := &streamInfo{IsVideo: strings.HasPrefix(ct, "video/")}

	// Playlists and manifests are tiny, so their size says nothing about length.
	isFile := info.IsVideo || strings.HasPrefix(ct, "audio/")
	if size := contentSize(resp.Header); isFile && size > 0 {
		bitrate := int64(128000)
		if info.IsVideo {
			bitrate = 500000
		}
		if dur := int(size * 8 / bitrate); dur <= 3600 {
			info.Duration = dur
		}
	}

	return info, nil
}

// contentSize returns the full resource size; a ranged reply carries it in
// Content-Range, not Content-Length.
func contentSize(h http.Header) int64 {
	if _, total, ok := strings.Cut(h.Get("Content-Range"), "/"); ok {
		size, _ := strconv.ParseInt(total, 10, 64)
		return size
	}
	size, _ := strconv.ParseInt(h.Get("Content-Length"), 10, 64)
	return size
}

func streamID(rawURL string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(rawURL))
	return fmt.Sprintf("direct_%x", h.Sum64())
}

func streamTitle(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "Direct Stream"
	}

	name := path.Base(parsed.Path)
	if before, _, found := strings.CutLast(name, "."); found && before != "" {
		name = before
	}
	name = strings.TrimSpace(strings.NewReplacer("_", " ", "-", " ").Replace(name))
	if name == "" || name == "." || name == "index" || name == "stream" {
		name = parsed.Host
	}

	if r, size := utf8.DecodeRuneInString(name); size > 0 {
		name = string(unicode.ToUpper(r)) + name[size:]
	}
	return name
}

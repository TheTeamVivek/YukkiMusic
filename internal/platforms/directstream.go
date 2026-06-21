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
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "yukkimusic/internal/core/models"
)

const PlatformDirectStream state.PlatformName = "DirectStream"

type DirectStreamPlatform struct {
	mu    sync.Mutex
	cache map[string]*cachedStream
}

type cachedStream struct {
	info    *streamInfo
	expires time.Time
}

type streamInfo struct {
	URL         string
	ContentType string
	Size        int64
	IsAudio     bool
	IsVideo     bool
	Duration    int
}

var (
	streamExtensions = []string{
		".m3u8", ".m3u", ".mp3", ".mp4", ".webm", ".ogg",
		".wav", ".flac", ".aac", ".opus", ".mkv", ".avi",
		".mov", ".ts", ".mpd",
	}
	streamPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.m3u8(\?|$)`),
		regexp.MustCompile(`(?i)\.mpd(\?|$)`),
		regexp.MustCompile(`(?i)/hls/`),
		regexp.MustCompile(`(?i)/dash/`),
		regexp.MustCompile(`(?i)/stream/`),
	}
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
		cache: make(map[string]*cachedStream),
	})
}

func (d *DirectStreamPlatform) Name() state.PlatformName { return PlatformDirectStream }
func (d *DirectStreamPlatform) Priority() int            { return 65 }

func (d *DirectStreamPlatform) CanGet(query string) bool {
	if _, err := sanitizeMediaURL(query); err != nil {
		return false
	}
	return d.looksLikeStream(query)
}

func (d *DirectStreamPlatform) Get(query string, video bool) ([]*state.Track, error) {
	safeURL, err := sanitizeMediaURL(query)
	if err != nil {
		return nil, errUnsafeURL
	}

	info, err := d.getStreamInfo(safeURL)
	if err != nil {
		return nil, fmt.Errorf("stream validation failed: %w", err)
	}

	track := &state.Track{
		ID:       d.generateID(safeURL),
		Title:    d.extractTitle(safeURL),
		Duration: info.Duration,
		URL:      safeURL,
		Source:   PlatformDirectStream,
		Video:    video || info.IsVideo,
	}

	if info.IsVideo {
		track.Video = true
	}

	if info.Size > 0 && track.Duration == 0 {
		bitrate := int64(128000)
		if info.IsVideo {
			bitrate = 500000
		}
		dur := int((info.Size * 8) / bitrate)
		if dur <= 3600 {
			track.Duration = dur
		}
	}

	gologging.InfoF("DirectStream: %s (audio:%v video:%v size:%d dur:%d)",
		track.Title, info.IsAudio, info.IsVideo, info.Size, track.Duration)

	return []*state.Track{track}, nil
}

func (d *DirectStreamPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformDirectStream
}

func (d *DirectStreamPlatform) Download(
	_ context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	gologging.InfoF("DirectStream: returning URL for streaming: %s", track.URL)
	return track.URL, nil
}

func (d *DirectStreamPlatform) getStreamInfo(urlStr string) (*streamInfo, error) {
	d.mu.Lock()
	if c, ok := d.cache[urlStr]; ok && time.Now().Before(c.expires) {
		d.mu.Unlock()
		return c.info, nil
	}
	d.mu.Unlock()

	info, err := d.validateStream(urlStr)
	if err != nil {
		return nil, err
	}

	d.mu.Lock()
	d.cache[urlStr] = &cachedStream{info: info, expires: time.Now().Add(5 * time.Minute)}
	d.mu.Unlock()

	return info, nil
}

func (d *DirectStreamPlatform) looksLikeStream(urlStr string) bool {
	parsed, _ := url.Parse(urlStr)
	ext := strings.ToLower(path.Ext(parsed.Path))
	if slices.Contains(streamExtensions, ext) {
		return true
	}
	for _, p := range streamPatterns {
		if p.MatchString(urlStr) {
			return true
		}
	}
	host := strings.ToLower(parsed.Host)
	for _, sh := range streamHosts {
		if strings.Contains(host, sh) {
			return true
		}
	}
	return false
}

func (d *DirectStreamPlatform) validateStream(urlStr string) (*streamInfo, error) {
	client := rc.
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.R().Head(urlStr)
	if err != nil || resp.StatusCode() >= 400 {
		resp, err = client.R().
			SetHeader("Range", "bytes=0-1024").
			Get(urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to validate URL: %w", err)
		}
	}

	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("URL returned status %d", resp.StatusCode())
	}

	ct := strings.ToLower(resp.Header().Get("Content-Type"))
	isAudio, isVideo := false, false

	for _, mt := range streamMimeTypes {
		if strings.HasPrefix(ct, mt) {
			switch {
			case strings.HasPrefix(mt, "audio/"):
				isAudio = true
			case strings.HasPrefix(mt, "video/"):
				isVideo = true
			default:
				isAudio = true
			}
			break
		}
	}

	if !isAudio && !isVideo {
		return nil, fmt.Errorf("not audio/video (content-type: %s)", ct)
	}

	size := int64(0)
	if cl := resp.Header().Get("Content-Length"); cl != "" {
		size, _ = strconv.ParseInt(cl, 10, 64)
	}

	return &streamInfo{
		URL:         urlStr,
		ContentType: ct,
		Size:        size,
		IsAudio:     isAudio,
		IsVideo:     isVideo,
	}, nil
}

func (d *DirectStreamPlatform) generateID(urlStr string) string {
	h := 0
	for _, c := range urlStr {
		h = (h << 5) - h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return fmt.Sprintf("direct_%d", h)
}

func (d *DirectStreamPlatform) extractTitle(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "Direct Stream"
	}
	name := path.Base(parsed.Path)
	if idx := strings.LastIndex(name, "."); idx > 0 {
		name = name[:idx]
	}
	name = strings.NewReplacer("_", " ", "-", " ").Replace(name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "index" || name == "stream" {
		name = parsed.Host
	}
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}

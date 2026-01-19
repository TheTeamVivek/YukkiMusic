/*
 * This file is part of YukkiMusic.
 *
 * DirectStream Platform - Handles direct audio/video URLs and M3U8 streams
 * This platform acts as a fallback when no other platform recognizes the URL
 *
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 */
package platforms

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
)

type DirectStreamPlatform struct {
	name state.PlatformName
}

var (
	// Common streaming file extensions
	streamExtensions = []string{
		".m3u8", ".m3u", ".mp3", ".mp4", ".webm", ".ogg",
		".wav", ".flac", ".aac", ".opus", ".mkv", ".avi",
		".mov", ".ts", ".mpd", // MPEG-DASH
	}

	// URL patterns that indicate streaming
	streamPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.m3u8(\?|$)`),
		regexp.MustCompile(`(?i)\.mpd(\?|$)`),
		regexp.MustCompile(`(?i)/hls/`),
		regexp.MustCompile(`(?i)/dash/`),
		regexp.MustCompile(`(?i)/stream/`),
	}

	// Common audio/video MIME types
	streamMimeTypes = []string{
		"audio/", "video/",
		"application/vnd.apple.mpegurl", // M3U8
		"application/x-mpegurl",         // M3U
		"application/dash+xml",          // MPEG-DASH
		"application/octet-stream",      // Generic binary
	}
	streamHosts = []string{"cdn", "stream", "media", "video", "audio"}
)

const PlatformDirectStream state.PlatformName = "DirectStream"

type streamInfo struct {
	URL         string
	ContentType string
	Size        int64
	IsAudio     bool
	IsVideo     bool
	Duration    int
}

func init() {
	// Lowest priority - acts as fallback
	Register(10, &DirectStreamPlatform{
		name: PlatformDirectStream,
	})
}

func (d *DirectStreamPlatform) Name() state.PlatformName {
	return d.name
}

func (d *DirectStreamPlatform) CanGetTracks(query string) bool {
	query = strings.TrimSpace(query)

	// Must be a valid URL
	parsedURL, err := url.Parse(query)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	if d.looksLikeStream(query) {
		return true
	}

	_, err = d.validateStream(query)
	return err == nil
}

func (d *DirectStreamPlatform) GetTracks(
	query string,
	video bool,
) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	info, err := d.validateStream(query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate stream: %w", err)
	}

	// Extract metadata
	track := &state.Track{
		ID:       d.generateID(query),
		Title:    d.extractTitle(query),
		Duration: info.Duration,
		Artwork:  "",
		URL:      query,
		Source:   PlatformDirectStream,
		Video:    video || info.IsVideo,
	}

	// Try to get more metadata if possible
	d.enrichMetadata(track, info)

	return []*state.Track{track}, nil
}

func (d *DirectStreamPlatform) Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	// For direct streams, we don't download - just return the URL
	// The streaming system will handle it directly
	gologging.InfoF(
		"DirectStream: Returning URL for direct streaming: %s",
		track.URL,
	)
	return track.URL, nil
}

func (d *DirectStreamPlatform) CanDownload(
	source state.PlatformName,
) bool {
	return source == PlatformDirectStream
}

// looksLikeStream does a quick pattern check
func (d *DirectStreamPlatform) looksLikeStream(urlStr string) bool {
	// Check file extension
	parsedURL, _ := url.Parse(urlStr)
	ext := strings.ToLower(path.Ext(parsedURL.Path))

	for _, streamExt := range streamExtensions {
		if ext == streamExt {
			return true
		}
	}

	// Check URL patterns
	for _, pattern := range streamPatterns {
		if pattern.MatchString(urlStr) {
			return true
		}
	}

	// Check for common streaming domains
	host := strings.ToLower(parsedURL.Host)
	for _, sh := range streamHosts {
		if strings.Contains(host, sh) {
			return true
		}
	}

	return false
}

func (*DirectStreamPlatform) CanSearch() bool { return false }

func (*DirectStreamPlatform) Search(
	string,
	bool,
) ([]*state.Track, error) {
	return nil, nil
}

// validateStream makes a HEAD request to validate the URL
func (d *DirectStreamPlatform) validateStream(
	urlStr string,
) (*streamInfo, error) {
	client := rc.
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Try HEAD request first (faster)
	resp, err := client.R().Head(urlStr)
	if err != nil || resp.StatusCode() >= 400 {
		// If HEAD fails, try GET with range request
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

	contentType := resp.Header().Get("Content-Type")
	contentLength := resp.Header().Get("Content-Length")

	// Check if it's audio/video
	isAudio := false
	isVideo := false

	for _, mimeType := range streamMimeTypes {
		if strings.HasPrefix(strings.ToLower(contentType), mimeType) {
			if strings.HasPrefix(mimeType, "audio/") {
				isAudio = true
			} else if strings.HasPrefix(mimeType, "video/") {
				isVideo = true
			} else {
				isAudio = true
			}
			break
		}
	}

	if !isAudio && !isVideo {
		return nil, fmt.Errorf(
			"not a valid audio/video stream (content-type: %s)",
			contentType,
		)
	}

	size := int64(0)
	if contentLength != "" {
		size, _ = strconv.ParseInt(contentLength, 10, 64)
	}

	info := &streamInfo{
		URL:         urlStr,
		ContentType: contentType,
		Size:        size,
		IsAudio:     isAudio,
		IsVideo:     isVideo,
		Duration:    0, // Will be detected during playback
	}

	return info, nil
}

// generateID creates a unique ID from the URL
func (d *DirectStreamPlatform) generateID(urlStr string) string {
	// Use a hash of the URL for ID
	hash := 0
	for _, c := range urlStr {
		hash = (hash << 5) - hash + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return fmt.Sprintf("direct_%d", hash)
}

// extractTitle tries to get a meaningful title from the URL
func (d *DirectStreamPlatform) extractTitle(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "Direct Stream"
	}

	filename := path.Base(parsedURL.Path)

	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}

	filename = strings.ReplaceAll(filename, "_", " ")
	filename = strings.ReplaceAll(filename, "-", " ")
	filename = strings.TrimSpace(filename)

	if filename == "" || filename == "." || filename == "index" ||
		filename == "stream" {
		filename = parsedURL.Host
	}

	if len(filename) > 0 {
		filename = strings.ToUpper(filename[:1]) + filename[1:]
	}

	return filename
}

// enrichMetadata tries to extract more metadata
func (d *DirectStreamPlatform) enrichMetadata(
	track *state.Track,
	info *streamInfo,
) {
	filename := d.extractTitle(track.URL)
	track.Title = strings.TrimSpace(filename)

	if info.IsVideo {
		track.Video = true
	}

	if info.Size > 0 && track.Duration == 0 {
		var bitrate int64 = 128000 // bits per second
		if info.IsVideo {
			bitrate = 500000
		}

		track.Duration = int((info.Size * 8) / bitrate)

		if track.Duration > 3600 {
			track.Duration = 0
		}
	}

	gologging.InfoF(
		"DirectStream metadata: %s (audio: %v, video: %v, size: %d, duration: %d)",
		track.Title,
		info.IsAudio,
		info.IsVideo,
		info.Size,
		track.Duration,
	)
}

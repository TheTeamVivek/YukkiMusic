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
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/utils"
)

const PlatformSoundCloud state.PlatformName = "SoundCloud"

type SoundCloudPlatform struct {
	cache *utils.Cache[string, []*state.Track]
}

var soundcloudLinkRe = regexp.MustCompile(
	`(?i)^(https?://)?(www\.)?(soundcloud\.com|snd\.sc)/`,
)

func init() {
	Register(&SoundCloudPlatform{
		cache: utils.NewCache[string, []*state.Track](1 * time.Hour),
	})
}

func (s *SoundCloudPlatform) Name() state.PlatformName { return PlatformSoundCloud }
func (s *SoundCloudPlatform) Priority() int            { return 85 }

func (s *SoundCloudPlatform) CanGet(query string) bool {
	return soundcloudLinkRe.MatchString(strings.TrimSpace(query))
}

func (s *SoundCloudPlatform) Get(query string, _ bool) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	safeURL, err := sanitizeMediaURL(query)
	if err != nil {
		return nil, errUnsafeURL
	}

	cacheKey := "sc:" + strings.ToLower(query)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached, nil
	}

	info, err := s.extractMetadata(safeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	var tracks []*state.Track
	if len(info.Entries) > 0 {
		for _, entry := range info.Entries {
			tracks = append(tracks, s.toTrack(&entry))
		}
	} else {
		tracks = []*state.Track{s.toTrack(info)}
	}

	if len(tracks) > 0 {
		s.cache.Set(cacheKey, tracks)
	}

	return tracks, nil
}

func (s *SoundCloudPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformSoundCloud
}

func (s *SoundCloudPlatform) Download(
	ctx context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	track.Video = false

	if p := findFile(track); p != "" {
		return p, nil
	}

	safeURL, err := sanitizeMediaURL(track.URL)
	if err != nil {
		return "", errUnsafeURL
	}

	args := []string{
		"-f", "ba[abr>=128]/ba",
		"-x",
		"--concurrent-fragments", "4",
		"--no-playlist",
		"--no-part",
		"--no-warnings",
		"--no-overwrites",
		"--ignore-errors",
		"--no-check-certificate",
		"-q",
		"-o", getPath(track, ".%(ext)s"),
		"--", safeURL,
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		findAndRemove(track)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", fmt.Errorf(
			"yt-dlp failed: %w\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String(),
		)
	}

	p := findFile(track)
	if p == "" {
		return "", errors.New("yt-dlp produced no output file")
	}

	gologging.InfoF("SoundCloud: downloaded %s", track.Title)
	return p, nil
}

func (s *SoundCloudPlatform) extractMetadata(urlStr string) (*ytdlpInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	args := []string{
		"-j", "--flat-playlist",
		"--no-warnings", "--no-check-certificate",
		"--", urlStr,
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("metadata extraction failed: %w\n%s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")

	if len(lines) > 1 {
		var info ytdlpInfo
		for _, line := range lines {
			var entry ytdlpInfo
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				gologging.DebugF("SoundCloud: skip bad entry: %v", err)
				continue
			}
			info.Entries = append(info.Entries, entry)
		}
		if len(info.Entries) == 0 {
			return nil, errors.New("no valid entries in playlist")
		}
		return &info, nil
	}

	var info ytdlpInfo
	if err := json.Unmarshal([]byte(stdout.String()), &info); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}
	return &info, nil
}

func (s *SoundCloudPlatform) toTrack(info *ytdlpInfo) *state.Track {
	return &state.Track{
		ID:       info.ID,
		Title:    info.Title,
		Duration: int(info.Duration),
		Artwork:  info.Thumbnail,
		URL:      firstNonEmpty(info.OriginalURL, info.URL),
		Source:   PlatformSoundCloud,
		Video:    false,
	}
}

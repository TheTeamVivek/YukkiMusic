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
	"cmp"
	"context"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config/cookies"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/utils"
)

const PlatformYtDlp state.PlatformName = "YtDlp"

type YtdlpPlatform struct{}

type ytdlpInfo struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Duration    float64     `json:"duration"`
	Thumbnail   string      `json:"thumbnail"`
	URL         string      `json:"webpage_url"`
	DirectURL   string      `json:"url"`
	OriginalURL string      `json:"original_url"`
	Uploader    string      `json:"uploader"`
	IsLive      bool        `json:"is_live"`
	Extractor   string      `json:"extractor"`
	Entries     []ytdlpInfo `json:"entries"`
}

var (
	bannedExtractors = map[string]bool{
		"alphaporno": true, "beeg": true, "behindkink": true, "bongacams": true,
		"cam4": true, "cammodels": true, "camsoda": true, "chaturbate": true,
		"drtuber": true, "eporner": true, "erocast": true, "eroprofile": true,
		"fourtube": true, "goshgay": true, "hellporno": true, "iwara": true,
		"lovehomeporn": true, "manyvids": true, "motherless": true, "murrtube": true,
		"nonktube": true, "noodlemagazine": true, "nubilesporn": true, "nuvid": true,
		"oftv": true, "peekvids": true, "pornbox": true, "pornflip": true,
		"pornhub": true, "pornotube": true, "pornovoisines": true, "pornoxo": true,
		"redgifs": true, "redtube": true, "rule34video": true, "sauceplus": true,
		"sexu": true, "slutload": true, "spankbang": true, "stripchat": true,
		"sunporno": true, "thisvid": true, "tnaflix": true, "toypics": true,
		"txxx": true, "xhamster": true, "xnxx": true, "xvideos": true,
		"xxxymovies": true, "youjizz": true, "youporn": true, "zenporn": true,
	}
	// audioOnlyExtractors are sites without video streams; a video request is
	// silently downgraded to audio.
	audioOnlyExtractors = map[string]bool{
		"soundcloud": true,
	}
)

func init() {
	Register(&YtdlpPlatform{})
}

func (y *YtdlpPlatform) Name() state.PlatformName { return PlatformYtDlp }
func (y *YtdlpPlatform) Priority() int            { return 60 }

func (y *YtdlpPlatform) CanGet(query string) bool {
	safeURL, err := sanitizeMediaURL(query)
	if err != nil {
		return false
	}
	parsed, err := url.Parse(safeURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	switch host {
	case "t.me", "telegram.me", "telegram.dog":
		return false
	}
	return !strings.HasSuffix(host, ".t.me")
}

func (y *YtdlpPlatform) Get(query string, video bool) ([]*state.Track, error) {
	safeURL, err := sanitizeMediaURL(query)
	if err != nil {
		return nil, err
	}

	info, err := y.extractMetadata(safeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	if info.IsLive {
		return nil, errLiveStream
	}
	if bannedExtractors[strings.ToLower(info.Extractor)] {
		return nil, errors.New("adult content is not allowed")
	}

	if len(info.Entries) == 0 {
		return []*state.Track{y.toTrack(info, video)}, nil
	}

	var tracks []*state.Track
	for _, entry := range info.Entries {
		if entry.IsLive || bannedExtractors[strings.ToLower(entry.Extractor)] {
			continue
		}
		tracks = append(tracks, y.toTrack(&entry, video))
	}
	return tracks, nil
}

func (y *YtdlpPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformYtDlp || source == PlatformYouTube
}

func (y *YtdlpPlatform) Download(
	ctx context.Context,
	track *state.Track,
	_ *td.Message,
) (string, error) {
	if f := findFile(track); f != "" {
		utils.Log.Debug("YtDlp: cache hit " + f)
		return f, nil
	}

	safeURL, err := sanitizeMediaURL(track.URL)
	if err != nil {
		return "", err
	}

	args := []string{
		"--no-playlist",
		"--geo-bypass",
		"--no-warnings",
		"--ignore-errors",
		"--no-check-certificate",
		"-q",
		"-o", trackFile(track, ".%(ext)s"),
	}

	if track.Video {
		args = append(args,
			"-f", "(b[height>=360][height<=1080]/bv*[height>=360][height<=1080]/bv*)+(ba[abr>=180][abr<=360]/ba)/b",
		)
	} else {
		args = append(args,
			"-f", "ba[abr>=180][abr<=360]/ba",
			"-x",
			"--concurrent-fragments", "4",
		)
	}

	args = append(args, cookieArgs(track.URL)...)
	args = append(args, "--", safeURL)

	stdout, stderr, err := runYtDlp(ctx, args...)
	if err != nil {
		removeFiles(track)
		if canceled(ctx, err) {
			return "", context.Canceled
		}
		return "", fmt.Errorf("yt-dlp failed: %w\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	path := findFile(track)
	if path == "" {
		return "", fmt.Errorf("yt-dlp produced no output file\nstderr: %s", stderr)
	}

	utils.Log.Infof("YtDlp: downloaded %s", path)
	return path, nil
}

func (y *YtdlpPlatform) extractMetadata(urlStr string) (*ytdlpInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-j", "--flat-playlist", "--no-warnings", "--no-check-certificate"}
	args = append(args, cookieArgs(urlStr)...)
	args = append(args, "--", urlStr)

	stdout, stderr, err := runYtDlp(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("metadata extraction failed: %w\n%s", err, stderr)
	}
	if stdout == "" {
		return nil, errors.New("yt-dlp produced empty metadata")
	}

	// A playlist prints one JSON object per line, a single item prints one.
	lines := strings.Split(stdout, "\n")
	if len(lines) == 1 {
		var info ytdlpInfo
		if err := jsonv2.Unmarshal([]byte(lines[0]), &info); err != nil {
			return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
		}
		return &info, nil
	}

	var info ytdlpInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry ytdlpInfo
		if err := jsonv2.Unmarshal([]byte(line), &entry); err != nil {
			utils.Log.Debugf("YtDlp: skip bad entry: %v", err)
			continue
		}
		info.Entries = append(info.Entries, entry)
	}
	if len(info.Entries) == 0 {
		return nil, errors.New("no valid entries in playlist")
	}
	return &info, nil
}

func (y *YtdlpPlatform) toTrack(info *ytdlpInfo, video bool) *state.Track {
	return &state.Track{
		ID:       info.ID,
		Title:    cmp.Or(info.Title, "Audio Stream"),
		Duration: int(info.Duration),
		Artwork:  info.Thumbnail,
		URL:      cmp.Or(info.OriginalURL, info.URL, info.DirectURL),
		Source:   PlatformYtDlp,
		Video:    video && !audioOnlyExtractors[strings.ToLower(info.Extractor)],
	}
}

func runYtDlp(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	var out, errOut bytes.Buffer
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	cmd.Stdout, cmd.Stderr = &out, &errOut

	err = cmd.Run()
	return strings.TrimSpace(out.String()), strings.TrimSpace(errOut.String()), err
}

// cookieArgs returns the cookie flags for YouTube links, if cookies are configured.
func cookieArgs(rawURL string) []string {
	if !isYouTubeURL(rawURL) {
		return nil
	}
	if file, err := cookies.GetRandomCookieFile(); err == nil && file != "" {
		return []string{"--cookies", file}
	}
	return nil
}

func isYouTubeURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "youtu.be" || host == "youtube.com" || strings.HasSuffix(host, ".youtube.com")
}

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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/cookies"
	state "main/internal/core/models"
)

const PlatformYtDlp state.PlatformName = "YtDlp"

type YtDlpDownloader struct {
	name state.PlatformName
}

type ytdlpInfo struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Duration    float64     `json:"duration"`
	Thumbnail   string      `json:"thumbnail"`
	URL         string      `json:"webpage_url"`
	OriginalURL string      `json:"original_url"`
	Uploader    string      `json:"uploader"`
	Description string      `json:"description"`
	IsLive      bool        `json:"is_live"`
	Entries     []ytdlpInfo `json:"entries"`
}

// URLs that are likely handled by YouTube
var youtubePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(youtube\.com|youtu\.be|music\.youtube\.com)`),
}

func init() {
	Register(60, &YtDlpDownloader{
		name: PlatformYtDlp,
	})
}

func (y *YtDlpDownloader) Name() state.PlatformName {
	return y.name
}

// IsValid checks if this is a valid URL that yt-dlp might handle
func (y *YtDlpDownloader) IsValid(query string) bool {
	query = strings.TrimSpace(query)

	// Must be a URL
	parsedURL, err := url.Parse(query)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	return true
}

// GetTracks extracts metadata using yt-dlp
func (y *YtDlpDownloader) GetTracks(query string, video bool) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	gologging.InfoF("YtDlp: Extracting metadata for %s", query)

	info, err := y.extractMetadata(query)
	if err != nil {
		gologging.ErrorF("YtDlp: Failed to extract metadata: %v", err)
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Check if it's a live stream
	if info.IsLive {
		gologging.Info("YtDlp: Detected live stream, returning error")
		return nil, errors.New("live streams are not supported by yt-dlp downloader")
	}

	var tracks []*state.Track

	// Handle playlists
	if len(info.Entries) > 0 {
		gologging.InfoF("YtDlp: Found playlist with %d entries", len(info.Entries))
		for _, entry := range info.Entries {
			if entry.IsLive {
				continue // Skip live entries
			}
			track := y.infoToTrack(&entry, video)
			tracks = append(tracks, track)
		}
	} else {
		track := y.infoToTrack(info, video)
		tracks = []*state.Track{track}
	}

	if len(tracks) > 0 {
		gologging.InfoF("YtDlp: Successfully extracted %d track(s)", len(tracks))
	}

	return tracks, nil
}

func (y *YtDlpDownloader) IsDownloadSupported(source state.PlatformName) bool {
	// YtDlp can download from itself (when it extracts info)
	// and from YouTube platform
	return source == y.name || source == PlatformYouTube
}

func (y *YtDlpDownloader) Download(ctx context.Context, track *state.Track, _ *telegram.NewMessage) (string, error) {
	// Check if already downloaded
	if path, err := checkDownloadedFile(track.ID); err == nil {
		gologging.InfoF("YtDlp: Using cached file for %s", track.ID)
		return path, nil
	}

	gologging.InfoF("YtDlp: Downloading %s", track.Title)

	if err := ensureDownloadsDir(); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}
	
	ext := ".mp3"
	
  if track.Video {
    ext = ".mp4"
  }
  
	filePath := filepath.Join("downloads", track.ID+ext)

	args := []string{
		"--no-playlist",
		"-o", filePath,
		"--geo-bypass",
		"--no-warnings",
		"--no-overwrites",
		"--ignore-errors",
		"--no-check-certificate",
		"-q",
	}

	// Format selection based on platform and type
	if y.isYouTubeURL(track.URL) {
		if track.Video {
			// For YouTube videos: 720p max with audio
			args = append(args, "-f", "bestvideo[height<=720]+bestaudio/best[height<=720]")
			args = append(args, "--merge-output-format", "mp4")
		} else {
			// For YouTube audio: extract best audio
			args = append(args, "--extract-audio", "--audio-format", "best")
		}
	} else {
		// For other platforms: use best available
		if track.Video {
			args = append(args, "-f", "bestvideo+bestaudio/best")
		} else {
			args = append(args, "-f", "bestaudio/best")
		}
	}

	// Only add cookies for YouTube URLs
	if y.isYouTubeURL(track.URL) {
		cookieFile, err := cookies.GetRandomCookieFile()
		if err != nil {
			gologging.Debug("YtDlp: No cookie file available: " + err.Error())
		} else if cookieFile != "" {
			args = append(args, "--cookies", cookieFile)
		}
	}

	args = append(args, track.URL)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		outStr := stdout.String()
		errStr := stderr.String()

		gologging.ErrorF(
			"YtDlp: Download failed for %s: %v\nSTDOUT:\n%s\nSTDERR:\n%s",
			track.URL, err, outStr, errStr,
		)

		os.Remove(filePath)

		// Check for context cancellation
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		return "", fmt.Errorf("yt-dlp error: %w\nstdout: %s\nstderr: %s", err, outStr, errStr)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		gologging.ErrorF("YtDlp: Downloaded file not found: %v", err)
		return "", fmt.Errorf("downloaded file not found: %w", err)
	}

	gologging.InfoF("YtDlp: Successfully downloaded %s", track.Title)
	return filePath, nil
}

// extractMetadata uses yt-dlp to extract video/audio metadata
func (y *YtDlpDownloader) extractMetadata(urlStr string) (*ytdlpInfo, error) {
	args := []string{
		"-j",
		"--flat-playlist",
		"--no-warnings",
		"--no-check-certificate",
	}

	// Add cookies only for YouTube
	if y.isYouTubeURL(urlStr) {
		cookieFile, err := cookies.GetRandomCookieFile()
		if err == nil && cookieFile != "" {
			args = append(args, "--cookies", cookieFile)
		}
	}

	args = append(args, urlStr)

	cmd := exec.Command("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		gologging.ErrorF("YtDlp: Metadata extraction failed: %v\n%s", err, errStr)
		return nil, fmt.Errorf("metadata extraction failed: %w", err)
	}

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Handle playlists (multiple JSON objects)
	if len(lines) > 1 {
		var info ytdlpInfo
		info.Entries = make([]ytdlpInfo, 0, len(lines))

		for _, line := range lines {
			var entry ytdlpInfo
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				gologging.ErrorF("YtDlp: Failed to parse entry JSON: %v", err)
				continue
			}
			info.Entries = append(info.Entries, entry)
		}

		if len(info.Entries) == 0 {
			return nil, errors.New("no valid entries found in playlist")
		}

		return &info, nil
	}

	// Single video/audio
	var info ytdlpInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		gologging.ErrorF("YtDlp: Failed to parse JSON: %v", err)
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &info, nil
}

// infoToTrack converts yt-dlp info to Track
func (y *YtDlpDownloader) infoToTrack(info *ytdlpInfo, video bool) *state.Track {
	duration := int(info.Duration)

	// Use original_url if available, otherwise webpage_url
	trackURL := info.URL
	if info.OriginalURL != "" {
		trackURL = info.OriginalURL
	}

	return &state.Track{
		ID:       info.ID,
		Title:    info.Title,
		Duration: duration,
		Artwork:  info.Thumbnail,
		URL:      trackURL,
		Source:   PlatformYtDlp,
		Video:    video,
	}
}

// isYouTubeURL checks if the URL is from YouTube
func (y *YtDlpDownloader) isYouTubeURL(urlStr string) bool {
	for _, pattern := range youtubePatterns {
		if pattern.MatchString(urlStr) {
			return true
		}
	}
	return false
}

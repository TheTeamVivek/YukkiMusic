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
	"context"
	"errors"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
	"main/internal/utils"
)

// PlatformRegistry manages all registered platforms
type PlatformRegistry struct {
	platforms []platformEntry
	mu        sync.RWMutex
}

type platformEntry struct {
	platform state.Platform
	priority int
}

var registry = &PlatformRegistry{
	platforms: make([]platformEntry, 0),
}

// Register adds a platform to the registry with given priority
// Higher priority = checked first for URL validation
func Register(priority int, platform state.Platform) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.platforms = append(registry.platforms, platformEntry{
		platform: platform,
		priority: priority,
	})

	// Sort by priority (highest first)
	sort.Slice(registry.platforms, func(i, j int) bool {
		return registry.platforms[i].priority > registry.platforms[j].priority
	})
}

// GetOrderedPlatforms returns all platforms sorted by priority
func GetOrderedPlatforms() []state.Platform {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	result := make([]state.Platform, len(registry.platforms))
	for i, entry := range registry.platforms {
		result[i] = entry.platform
	}
	return result
}

// FindPlatform returns the first platform that validates the given URL
func FindPlatform(url string) state.Platform {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	for _, entry := range registry.platforms {
		if entry.platform.IsValid(url) {
			return entry.platform
		}
	}
	return nil
}

// GetTracks extracts tracks from the given query
// Automatically detects the appropriate platform
func GetTracks(m *telegram.NewMessage, video bool) ([]*state.Track, error) {
	gologging.Debug("GetTracks called")

	urls, _ := utils.ExtractURLs(m)
	query := m.Args()

	var allTracks []*state.Track
	var errorsL []string

	// Process URLs first
	if len(urls) > 0 {

		for _, url := range urls {
			gologging.Info("Processing URL: " + url)

			platform := FindPlatform(url)
			if platform == nil {
				errMsg := "No platform found for URL: " + url
				gologging.Error(errMsg)
				errorsL = append(errorsL, errMsg)
				continue
			}

			gologging.Debug("Found platform: " + string(platform.Name()))

			tracks, err := platform.GetTracks(url, video)
			// ytdlp didn't support or unable to extract, skip silently
			if err != nil {

				if strings.Contains(
					err.Error(),
					"failed to extract metadata: metadata extraction failed",
				) {
					continue
				}

				errMsg := string(platform.Name()) + ": " + err.Error()
				gologging.Error(errMsg)
				errorsL = append(errorsL, errMsg)
				continue
			}

			gologging.Info("Tracks found: " + strconv.Itoa(len(tracks)))
			allTracks = append(allTracks, tracks...)
		}

		// If we have tracks from URLs, return them
		if len(allTracks) > 0 {
			gologging.Info("Returning tracks from URLs")
			return allTracks, nil
		}

		if len(errorsL) == 0 {
			return nil, errors.New("no supported platform for given URL(s)")
		}
		return nil, formatErrors(errorsL)

	}

	// If no URLs but have query, search YouTube
	if query != "" {
		gologging.Info("No URLs found, searching YouTube with query: " + query)

		yt := &YouTubePlatform{}
		tracks, err := yt.GetTracks(query, video)
		if err != nil {
			gologging.Error("YouTube search failed: " + err.Error())
			return nil, err
		}

		if len(tracks) > 0 {
			gologging.Info("YouTube track found, returning first result")
			return []*state.Track{tracks[0]}, nil
		}
	}

	// Handle reply messages
	if m.IsReply() {
		gologging.Debug("Message is a reply, checking media")

		rmsg, err := m.GetReplyMessage()
		if err != nil {
			gologging.Error("Failed to get replied message: " + err.Error())
			return nil, errors.New(
				"failed to get replied message: " + err.Error(),
			)
		}

		if !(rmsg.IsMedia() &&
			(rmsg.Audio() != nil || rmsg.Video() != nil || rmsg.Voice() != nil || rmsg.Document() != nil)) {
			gologging.Info("Reply does not contain valid media")
			return nil, errors.New("⚠️ Reply with a valid media (audio/video)")
		}

		tg := &TelegramPlatform{}
		isAudio := false
		isVideo := false

		if rmsg.Audio() != nil || rmsg.Voice() != nil {
			isAudio = true
		} else if rmsg.Video() != nil {
			isVideo = true
		} else if rmsg.Document() != nil {
			ext := strings.ToLower(rmsg.File.Ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			mimeType := mime.TypeByExtension(ext)
			isAudio = strings.HasPrefix(mimeType, "audio/")
			isVideo = strings.HasPrefix(mimeType, "video/")
		}

		if !isAudio && !isVideo {
			gologging.Info("Replied media is neither audio nor video")
			return nil, errors.New("⚠️ Reply with a valid media (audio/video)")
		}

		t, err := tg.GetTracksByMessage(rmsg)
		if err != nil {
			errMsg := "Failed to get track from reply: " + err.Error()
			gologging.Error(errMsg)
			errorsL = append(errorsL, errMsg)
		} else {
			t.Video = isVideo

			if isVideo {
				gologging.Debug("Reply media is video, preparing thumbnail")

				if err := os.MkdirAll("cache", os.ModePerm); err != nil {
					gologging.Error("Failed to create cache folder: " + err.Error())
					return []*state.Track{t}, nil
				}

				thumbPath := filepath.Join("cache", "thumb_"+t.ID+".jpg")
				if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
					path, err := rmsg.Download(&telegram.DownloadOptions{
						ThumbOnly: true,
						FileName:  thumbPath,
					})
					if err == nil {
						if _, err := os.Stat(path); err == nil {
							t.Artwork = path
							gologging.Debug("Thumbnail saved: " + path)
						}
					}
				}
			}

			gologging.Info("Returning track from reply message")
			return []*state.Track{t}, nil
		}
	}

	if len(errorsL) > 0 {
		gologging.Error("Returning aggregated errors")
		return nil, formatErrors(errorsL)
	}

	gologging.Info("No tracks found")
	return nil, errors.New("no tracks found")
}

// Download attempts to download a track using available downloaders
func Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	var errs []string

	for _, p := range GetOrderedPlatforms() {
		if !p.IsDownloadSupported(track.Source) {
			continue
		}

		path, err := p.Download(ctx, track, mystic)
		if err == nil {
			// Special case: DirectStream returns the URL itself, not a file path
			// The streaming system will handle it
			return path, nil
		}

		// Don't retry on context cancellation
		if errors.Is(err, context.Canceled) {
			return "", err
		}

		errs = append(errs, string(p.Name())+": "+err.Error())
	}

	if len(errs) > 0 {
		return "", formatErrors(errs)
	}

	return "", errors.New("no downloader available for " + string(track.Source))
}

func formatErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errors.New(errs[0])
	}

	msg := "Multiple errors occurred:\n"
	for _, e := range errs {
		msg += "• " + e + "\n"
	}
	return errors.New(msg)
}

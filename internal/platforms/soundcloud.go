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
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
	"main/internal/utils"
)

type SoundCloudPlatform struct {
	name state.PlatformName
}

var (
	soundcloudLinkRegex = regexp.MustCompile(
		`(?i)^(https?://)?(www\.)?(soundcloud\.com|snd\.sc)/`,
	)
	soundcloudCache = utils.NewCache[string, []*state.Track](1 * time.Hour)
)

const PlatformSoundCloud state.PlatformName = "SoundCloud"

func init() {
	Register(85, &SoundCloudPlatform{
		name: PlatformSoundCloud,
	})
}

func (s *SoundCloudPlatform) Name() state.PlatformName {
	return s.name
}

func (s *SoundCloudPlatform) CanGetTracks(query string) bool {
	return soundcloudLinkRegex.MatchString(strings.TrimSpace(query))
}

func (s *SoundCloudPlatform) GetTracks(
	query string,
	_ bool,
) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	cacheKey := "soundcloud:" + strings.ToLower(query)
	if cached, ok := soundcloudCache.Get(cacheKey); ok {
		gologging.Debug("SoundCloud: Using cached tracks")
		return cached, nil
	}

	gologging.InfoF("SoundCloud: Fetching metadata for %s", query)

	info, err := s.extractMetadata(query)
	if err != nil {
		gologging.ErrorF("SoundCloud: Failed to extract metadata: %v", err)
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	var tracks []*state.Track

	if len(info.Entries) > 0 {
		gologging.InfoF(
			"SoundCloud: Found playlist with %d tracks",
			len(info.Entries),
		)
		for _, entry := range info.Entries {
			track := s.infoToTrack(&entry)
			tracks = append(tracks, track)
		}
	} else {
		track := s.infoToTrack(info)
		tracks = []*state.Track{track}
	}

	if len(tracks) > 0 {
		soundcloudCache.Set(cacheKey, tracks)
		gologging.InfoF(
			"SoundCloud: Successfully extracted %d track(s)",
			len(tracks),
		)
	}

	return tracks, nil
}

func (s *SoundCloudPlatform) CanDownload(
	source state.PlatformName,
) bool {
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

	gologging.InfoF("SoundCloud: Downloading %s", track.Title)

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
			"SoundCloud: yt-dlp download failed for %s: %v\nSTDOUT:\n%s\nSTDERR:\n%s",
			track.URL,
			err,
			outStr,
			errStr,
		)

		findAndRemove(track)

		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		return "", fmt.Errorf(
			"download failed: %w\nstdout: %s\nstderr: %s",
			err,
			outStr,
			errStr,
		)
	}

	path := findFile(track)
	if path == "" {
		return "", errors.New("yt-dlp did not return output file path")
	}

	gologging.InfoF("SoundCloud: Successfully downloaded %s", track.Title)
	return path, nil
}

func (*SoundCloudPlatform) CanSearch() bool { return false } // can but for now not needed
func (*SoundCloudPlatform) Search(
	string,
	bool,
) ([]*state.Track, error) {
	return nil, nil
}

func (s *SoundCloudPlatform) extractMetadata(url string) (*ytdlpInfo, error) {
	args := []string{
		"-j",
		"--flat-playlist",
		"--no-warnings",
		"--no-check-certificate",
		url,
	}

	cmd := exec.Command("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		gologging.ErrorF(
			"SoundCloud: yt-dlp metadata extraction failed: %v\n%s",
			err,
			errStr,
		)
		return nil, fmt.Errorf("metadata extraction failed: %w", err)
	}

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) > 1 {
		var info ytdlpInfo
		info.Entries = make([]ytdlpInfo, 0, len(lines))

		for _, line := range lines {
			var entry ytdlpInfo
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				gologging.ErrorF(
					"SoundCloud: Failed to parse entry JSON: %v",
					err,
				)
				continue
			}
			info.Entries = append(info.Entries, entry)
		}

		if len(info.Entries) == 0 {
			err := errors.New("no valid entries found in playlist")
			gologging.Error("SoundCloud: " + err.Error())
			return nil, err
		}

		return &info, nil
	}

	var info ytdlpInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		gologging.ErrorF("SoundCloud: Failed to parse JSON: %v", err)
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &info, nil
}

func (s *SoundCloudPlatform) infoToTrack(info *ytdlpInfo) *state.Track {
	title := info.Title
	duration := int(info.Duration)

	track := &state.Track{
		ID:       info.ID,
		Title:    title,
		Duration: duration,
		Artwork:  info.Thumbnail,
		URL:      info.URL,
		Source:   PlatformSoundCloud,
		Video:    false,
	}

	return track
}

func (s *SoundCloudPlatform) CanGetRecommendations() bool {
	return false
}

func (s *SoundCloudPlatform) GetRecommendations(
	track *state.Track,
) ([]*state.Track, error) {
	return nil, errors.New("recommendations not supported on soundcloud")
}

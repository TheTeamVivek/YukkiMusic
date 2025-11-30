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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/amarnathcjd/gogram/telegram"
	"main/internal/state"
	"resty.dev/v3"
)

type PlatformName = state.PlatformName
type YoutubifyPlatform struct{}

var (
	apiBase = os.Getenv("YOUTUBIFY_API_URL")
	apiKey  = os.Getenv("YOUTUBIFY_API_KEY")
)

const (
	PlatformYoutubify state.PlatformName = "Youtubeify API"
)

func init() {
	addPlatform(100, PlatformYoutubify, &YoutubifyPlatform{})

	if apiBase == "" {
		apiBase = "https://youtubify.me"
	}
}

func (*YoutubifyPlatform) Name() state.PlatformName {
	return PlatformYoutubify
}

func (*YoutubifyPlatform) IsValid(query string) bool {
	return false
}

func (*YoutubifyPlatform) GetTracks(query string) ([]*state.Track, error) {
	return nil, errors.New("youtubify is a direct download platform")
}

func (*YoutubifyPlatform) IsDownloadSupported(source PlatformName) bool {
	return source == PlatformYouTube && apiKey != ""
}

func (f *YoutubifyPlatform) Download(
	_ context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	return downloadAudio(track.ID)
}

func downloadAudio(videoID string) (string, error) {
	filepath := fmt.Sprintf("downloads/%s.mp3", videoID)

	if _, err := os.Stat(filepath); err == nil {
		return filepath, nil
	}

	if err := os.MkdirAll("downloads", 0755); err != nil {
		return "", err
	}

	client := resty.New()
	url := fmt.Sprintf("%s/download/audio?video_id=%s&mode=download&no_redirect=1&api_key=%s", apiBase, videoID, apiKey)

	resp, err := client.R().SetOutputFileName(filepath).Get(url)
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("API returned %s", resp.Status())
	}

	return filepath, nil
}

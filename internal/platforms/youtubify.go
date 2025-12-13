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
	"resty.dev/v3"

	"main/internal/config"
	state "main/internal/core/models"
)

type YoutubifyPlatform struct {
	name state.PlatformName
}

const PlatformYoutubify state.PlatformName = "Youtubeify API"

func init() {
	Register(100, &YoutubifyPlatform{
		name: PlatformYoutubify,
	})
}

func (y *YoutubifyPlatform) Name() state.PlatformName {
	return y.name
}

func (y *YoutubifyPlatform) IsValid(query string) bool {
	return false
}

func (y *YoutubifyPlatform) GetTracks(_ string, _ bool) ([]*state.Track, error) {
	return nil, errors.New("youtubify is a direct download platform")
}

func (y *YoutubifyPlatform) IsDownloadSupported(source state.PlatformName) bool {
	return source == PlatformYouTube && config.YoutubifyApiKey != ""
}

func (y *YoutubifyPlatform) Download(
	ctx context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	ext := "mp3"
	endpoint := "audio"

	if track.Video {
		ext = "mp4"
		endpoint = "video"
	}

	filepath := "downloads/" + track.ID + "." + ext

	if _, err := os.Stat(filepath); err == nil {
		return filepath, nil
	}

	if err := os.MkdirAll("downloads", 0o755); err != nil {
		return "", err
	}

	client := resty.New()
	defer client.Close()

	url := config.YoutubifyApiURL +
		"/download/" + endpoint +
		"?video_id=" + track.ID +
		"&mode=download&no_redirect=1&api_key=" + config.YoutubifyApiKey

	resp, err := client.R().
		SetContext(ctx).
		SetOutputFileName(filepath).
		Get(url)
	if err != nil {
		os.Remove(filepath)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", err
	}

	if resp.IsError() {
		_ = os.Remove(filepath)
		return "", fmt.Errorf("API returned %s", resp.Status())
	}

	if ctx.Err() != nil {
		_ = os.Remove(filepath)
		return "", ctx.Err()
	}

	return filepath, nil
}

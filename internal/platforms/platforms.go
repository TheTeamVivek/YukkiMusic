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
	"html"
	"mime"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/state"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

func GetTracks(m *telegram.NewMessage) ([]*state.Track, error) {
	var tracks []*state.Track
	query := m.Args()
	errorsCollected := []string{}

	// --- 1. Extract URLs from message and reply ---
	urls, _ := utils.ExtractURLs(m)
	for _, url := range urls {
		found := false
		for _, p := range getOrderedPlatforms() {
			if p.IsValid(url) {
				t, err := p.GetTracks(url)
				if err != nil {
					errorsCollected = append(errorsCollected, fmt.Sprintf("<b>%s error:</b> %v", p.Name(), err))
				} else {
					tracks = append(tracks, t...)
					found = true

					// Check queue limit
					if config.QueueLimit > 0 && len(tracks) >= config.QueueLimit {
						return tracks[:config.QueueLimit], nil
					}
				}
				break
			}
		}

		if !found {
			errorsCollected = append(errorsCollected, fmt.Sprintf("Unsupported or invalid URL: %s", html.EscapeString(url)))
		}
	}

	if len(tracks) > 0 {
		return tracks, nil
	}

	// --- 2. If no URL tracks, try query as normal search ---
	if query != "" {
		yt := &YouTubePlatform{}
		ytTracks, err := yt.VideoSearch(query, true)
		if err != nil {
			return nil, fmt.Errorf("<b>⚠️ YouTube search error</b>\n\n<i>%s</i>", html.EscapeString(err.Error()))
		}

		if len(ytTracks) > 0 {
			if config.QueueLimit > 0 && len(ytTracks) > config.QueueLimit {
				ytTracks = ytTracks[:config.QueueLimit]
			}
			return []*state.Track{ytTracks[0]}, nil
		}
		errorsCollected = append(errorsCollected, fmt.Sprintf("No results found for: %s", html.EscapeString(query)))
	}

	// --- 3. If no URLs or query, check replied media ---
	if m.IsReply() {
		rmsg, err := m.GetReplyMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to get replied message: %v", err)
		}

		if rmsg.IsMedia() && (rmsg.Audio() != nil || rmsg.Video() != nil || rmsg.Voice() != nil || rmsg.Document() != nil) {
			tg := &TelegramPlatform{}
			ext := strings.ToLower(rmsg.File.Ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}

			mimeType := mime.TypeByExtension(ext)
			audio := strings.HasPrefix(mimeType, "audio/")
			video := strings.HasPrefix(mimeType, "video/")

			if audio || video {
				t, err := tg.GetTracksByMessage(rmsg)
				if err != nil {
					errorsCollected = append(errorsCollected, fmt.Sprintf("Failed to get track from reply: %v", err))
				} else {
					if config.QueueLimit > 0 && len(t) > config.QueueLimit {
						t = t[:config.QueueLimit]
					}
					return t, nil
				}
			} else {
				return nil, fmt.Errorf("⚠️ Reply with a valid media (audio/video)")
			}
		} else {
			return nil, fmt.Errorf("⚠️ Reply with a valid media (audio/video)")
		}
	}

	if len(errorsCollected) > 0 {
		return nil, errors.New(strings.Join(errorsCollected, "\n"))
	}
	return nil, errors.New("⚠️ Provide a song to play")
}

func Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
	for _, p := range getOrderedPlatforms() {
		if p.IsDownloadSupported(track.Source) {
			return p.Download(ctx, track, mystic)
		}
	}
	return "", fmt.Errorf("no downloader available for source %q", track.Source)
}
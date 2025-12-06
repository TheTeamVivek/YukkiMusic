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
	"html"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/state"
	"main/internal/utils"
)

func clampTracks(tracks []*state.Track) []*state.Track {
	if config.QueueLimit > 0 && len(tracks) > config.QueueLimit {
		return tracks[:config.QueueLimit]
	}
	return tracks
}

func GetTracks(m *telegram.NewMessage, video bool) ([]*state.Track, error) {
	var tracks []*state.Track
	query := m.Args()
	var errorsCollected []string

	urls, _ := utils.ExtractURLs(m)

	for _, url := range urls {
		supported := false

		for _, p := range getOrderedPlatforms() {
			if !p.IsValid(url) {
				continue
			}

			supported = true
			t, err := p.GetTracks(url, video)
			if err != nil {
				errorsCollected = append(errorsCollected,
					"<b>"+html.EscapeString(string(p.Name()))+" error:</b> "+html.EscapeString(err.Error()))
			} else {
				tracks = clampTracks(append(tracks, t...))
				if len(tracks) > 0 && (config.QueueLimit == 0 || len(tracks) >= config.QueueLimit) {
					return tracks, nil
				}
			}
			break
		}

		if !supported {
			errorsCollected = append(errorsCollected,
				"Unsupported or invalid URL: "+html.EscapeString(url))
		}
	}

	if len(tracks) > 0 {
		return tracks, nil
	}

	if len(errorsCollected) > 0 {
		return nil, formatErrorsHTML(errorsCollected)
	}

	if query != "" {
		yt := &YouTubePlatform{}
		ytTracks, err := yt.GetTracks(query, video)
		if err != nil {
			return nil, errors.New(
				"<b>⚠️ YouTube search error</b>\n\n<i>" +
					html.EscapeString(err.Error()) + "</i>",
			)
		}

		if len(ytTracks) > 0 {
			return []*state.Track{ytTracks[0]}, nil
		}

		errorsCollected = append(errorsCollected,
			"No results found for: "+html.EscapeString(query))
	}

	if m.IsReply() {
		rmsg, err := m.GetReplyMessage()
		if err != nil {
			return nil, errors.New("failed to get replied message: " + err.Error())
		}

		if !(rmsg.IsMedia() &&
			(rmsg.Audio() != nil || rmsg.Video() != nil || rmsg.Voice() != nil || rmsg.Document() != nil)) {
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
			return nil, errors.New("⚠️ Reply with a valid media (audio/video)")
		}

		t, err := tg.GetTracksByMessage(rmsg)
		if err != nil {
			errorsCollected = append(errorsCollected,
				"Failed to get track from reply: "+html.EscapeString(err.Error()))
		} else {
			// for tg medias we allow only Video when replied media is a video
			t.Video = isVideo
			if isVideo {
				thumbPath := filepath.Join("downloads", "thumb_"+t.ID+".jpg")

				if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
					path, err := rmsg.Download(&telegram.DownloadOptions{
						ThumbOnly: true,
						FileName:  thumbPath,
					})
					if err == nil {
						if _, err := os.Stat(path); err == nil {
							t.Artwork = path
						}
					}
				}
			}
			return []*state.Track{t}, nil
		}
	}

	if len(errorsCollected) > 0 {
		return nil, formatErrorsHTML(errorsCollected)
	}

	return nil, errors.New("⚠️ Provide a song to play")
}

func Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
	var errs []string

	for _, p := range getOrderedPlatforms() {
		if p.IsDownloadSupported(track.Source) {
			path, err := p.Download(ctx, track, mystic)
			if err == nil {
				return path, nil
			}
			if errors.Is(err, context.Canceled) {
				return "", err
			}
			errs = append(errs,
				html.EscapeString(string(p.Name()))+": "+html.EscapeString(err.Error()))
		}
	}

	if len(errs) > 0 {
		return "", formatErrorsHTML(errs)
	}

	return "", errors.New("⚠️ No downloader available for source \"" + html.EscapeString(string(track.Source)) + "\"")
}

func formatErrorsHTML(errs []string) error {
	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errors.New(errs[0])
	}

	var b strings.Builder
	b.Grow(64 + len(errs)*32)

	b.WriteString("<blockquote><b>⚠️ Multiple errors occurred:</b>\n\n")

	for _, e := range errs {
		b.WriteString("• ")
		b.WriteString(e)
	}

	b.WriteString("</blockquote>")
	return errors.New(b.String())
}

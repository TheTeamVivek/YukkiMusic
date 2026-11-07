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
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/utils"
)

const PlatformTelegram state.PlatformName = "Telegram"

type TelegramPlatform struct{}

var (
	telegram = &TelegramPlatform{}

	telegramLinkRegex = regexp.MustCompile(
		`^(?:(?:https?://)?t\.me/((c/)?[\w\d_-]+/\d+|[\w\d_-]+)|@[\w\d_-]+)$`,
	)
	telegramExtractRegex = regexp.MustCompile(
		`^(?:https?://)?t\.me/(c/)?([\w\d_-]+)/(\d+)(?:\?.*)?$`,
	)
	telegramProfileRegex = regexp.MustCompile(
		`^(?:(?:https?://)?t\.me/|@)([\w\d_-]{4,})/?$`,
	)
	telegramMsgCache = utils.NewCache[string, *td.Message](1 * time.Hour)
)

func init() {
	Register(telegram)
}

func (t *TelegramPlatform) Name() state.PlatformName { return PlatformTelegram }
func (t *TelegramPlatform) Priority() int            { return 100 }

func (t *TelegramPlatform) CanGet(query string) bool {
	return telegramLinkRegex.MatchString(strings.TrimSpace(query))
}

func (t *TelegramPlatform) Get(query string, _ bool) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	switch {
	case telegramExtractRegex.MatchString(query):
		msg, err := fetchTelegramMessage(query)
		if err != nil {
			return nil, err
		}

		target, isVideo, _ := playableMedia(core.Bot, msg)
		if target == nil {
			return nil, errors.New("message does not contain playable media")
		}

		track, err := t.GetTracksByMessage(core.Bot, target)
		if err != nil {
			return nil, err
		}
		track.Video = isVideo
		return []*state.Track{track}, nil

	case telegramProfileRegex.MatchString(query):
		return nil, errors.New("telegram profile music is not supported")
	}

	return nil, errors.New("invalid Telegram link")
}

func fetchTelegramMessage(query string) (*td.Message, error) {
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		query = "https://" + query
	}

	info, err := core.Bot.GetMessageLinkInfo(query)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve message link: %w", err)
	}
	if info == nil || info.Message == nil {
		return nil, errors.New("could not fetch message from link")
	}

	return info.Message, nil
}

func (t *TelegramPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformTelegram
}

func (t *TelegramPlatform) Download(
	ctx context.Context,
	track *state.Track,
	statusMsg *td.Message,
) (string, error) {
	ext := ".mp3"
	if track.Video {
		ext = ".mp4"
	}
	path := trackFile(track, ext)

	if !fileExists(path) {
		msg, ok := telegramMsgCache.Get(track.ID)
		if !ok {
			var err error
			if msg, err = fetchTelegramMessage(track.URL); err != nil {
				return "", errors.New("failed to locate file")
			}
		}

		var err error
		if path, err = downloadTDLibMessage(ctx, msg, path, statusMsg); err != nil {
			return "", err
		}
	}

	ensureDuration(track, path)
	return path, nil
}

func ensureDuration(track *state.Track, path string) {
	if track.Duration != 0 {
		return
	}

	dur, err := utils.GetDurationByFFProbe(path)
	if err != nil {
		utils.Log.Errorf("GetDurationByFFProbe failed for %q: %v", path, err)
		return
	}
	track.Duration = dur
}

func downloadTDLibMessage(
	ctx context.Context,
	msg *td.Message,
	dstPath string,
	statusMsg *td.Message,
) (string, error) {
	if OnDownloadStart != nil {
		OnDownloadStart(msg.RemoteFileID(), statusMsg)
	}

	file, err := msg.Download(core.Bot, 1, &td.DownloadFileOpts{Synchronous: true})
	if err != nil {
		_ = os.Remove(dstPath)
		if canceled(ctx, err) {
			return "", context.Canceled
		}
		return "", fmt.Errorf("download failed: %w", err)
	}
	if file == nil || file.Local == nil || file.Local.Path == "" {
		_ = os.Remove(dstPath)
		return "", errors.New("downloaded file missing")
	}

	if err := copyFile(file.Local.Path, dstPath); err != nil {
		_ = os.Remove(dstPath)
		return "", fmt.Errorf("copy failed: %w", err)
	}
	_ = os.Remove(file.Local.Path)

	return dstPath, nil
}

// GetTracksByMessage builds a track from a message that carries audio or video.
func (t *TelegramPlatform) GetTracksByMessage(c *td.Client, msg *td.Message) (*state.Track, error) {
	if msg == nil || msg.RemoteFileID() == "" {
		return nil, ErrNotMedia
	}

	title := "Telegram File"
	switch content := msg.Content.(type) {
	case *td.MessageAudio:
		if content.Audio != nil {
			title = cmp.Or(content.Audio.Title, content.Audio.FileName, title)
		}
	case *td.MessageVideo:
		if content.Video != nil {
			title = cmp.Or(content.Video.FileName, title)
		}
	case *td.MessageVoiceNote:
		title = "Voice Note"
	case *td.MessageDocument:
		if content.Document != nil {
			title = cmp.Or(content.Document.FileName, title)
		}
	}

	trackURL := ""
	if link, err := msg.GetLink(c); err == nil && link != nil {
		trackURL = link.Link
	}

	track := &state.Track{
		ID:       msg.RemoteFileID(),
		Title:    title,
		Duration: utils.GetDuration(msg),
		URL:      trackURL,
		Source:   PlatformTelegram,
	}
	telegramMsgCache.Set(track.ID, msg)

	return track, nil
}

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
	Register(&TelegramPlatform{})
}

func (t *TelegramPlatform) Name() state.PlatformName { return PlatformTelegram }
func (t *TelegramPlatform) Priority() int            { return 100 }

func (t *TelegramPlatform) CanGet(query string) bool {
	return telegramLinkRegex.MatchString(strings.TrimSpace(query))
}

func (t *TelegramPlatform) Get(query string, _ bool) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	if telegramExtractRegex.MatchString(query) {
		msg, err := fetchTelegramMessage(query)
		if err != nil {
			return nil, err
		}

		_, isVideo, isAudio := playableMedia(core.Bot, msg)
		if !isVideo && !isAudio {
			return nil, errors.New("message does not contain playable media")
		}

		track, err := t.GetTracksByMessage(core.Bot, msg)
		if err != nil {
			return nil, err
		}
		track.Video = isVideo
		return []*state.Track{track}, nil
	}

	if matches := telegramProfileRegex.FindStringSubmatch(query); len(matches) >= 2 {
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
	path := getPath(track, ext)

	if fileExists(path) {
		t.ensureDuration(track, path)
		return path, nil
	}

	msg, msgOk := telegramMsgCache.Get(track.ID)
	if !msgOk {
		fetched, fetchErr := fetchTelegramMessage(track.URL)
		if fetchErr == nil {
			msg = fetched
			msgOk = true
		}
	}
	if !msgOk {
		return "", errors.New("failed to locate file")
	}

	outPath, err := downloadTDLibMessage(ctx, msg, path, statusMsg)
	if err != nil {
		return "", err
	}

	t.ensureDuration(track, outPath)
	return outPath, nil
}

func (t *TelegramPlatform) ensureDuration(track *state.Track, path string) {
	if track.Duration == 0 {
		if dur, err := utils.GetDurationByFFProbe(path); err == nil {
			track.Duration = dur
		} else {
			utils.Log.Errorf("GetDurationByFFProbe failed for %q: %v", path, err)
		}
	}
}

func downloadTDLibMessage(
	ctx context.Context,
	msg *td.Message,
	dstPath string,
	statusMsg *td.Message,
) (string, error) {
	if msg == nil {
		return "", errors.New("nil message")
	}

	fileID := msg.RemoteFileID()
	if OnDownloadStart != nil {
		OnDownloadStart(fileID, statusMsg)
	}

	file, err := msg.Download(core.Bot, 1, &td.DownloadFileOpts{Synchronous: true})
	if err != nil {
		_ = os.Remove(dstPath)
		if isDownloadCancelled(err) || (ctx != nil && ctx.Err() != nil) {
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

func (t *TelegramPlatform) GetTracksByMessage(c *td.Client, msg *td.Message) (*state.Track, error) {
	if msg == nil {
		return nil, errors.New("nil message")
	}

	if msg.RemoteFileID() == "" && msg.ReplyToMessageID() > 0 {
		r, err := msg.GetRepliedMessage(c)
		if err == nil && r.RemoteFileID() != "" {
			msg = r
		}
	}

	if msg.RemoteFileID() == "" {
		return nil, fmt.Errorf("⚠️ This message doesn't contain any media")
	}

	title := "Telegram File"
	switch content := msg.Content.(type) {
	case *td.MessageAudio:
		if content.Audio != nil {
			title = firstNonEmpty(content.Audio.Title, content.Audio.FileName, title)
		}
	case *td.MessageVideo:
		if content.Video != nil {
			title = firstNonEmpty(content.Video.FileName, title)
		}
	case *td.MessageVoiceNote:
		title = "Voice Note"
	case *td.MessageDocument:
		if content.Document != nil {
			title = firstNonEmpty(content.Document.FileName, title)
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

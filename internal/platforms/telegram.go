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
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/utils"
)

const PlatformTelegram state.PlatformName = "Telegram"

type TelegramPlatform struct{}

var (
	telegramLinkRegex = regexp.MustCompile(
		`^(?:(?:https?://)?t\.me/((c/)?[\w\d_-]+/\d+|[\w\d_-]+)|@[\w\d_-]+)$`,
	)
	telegramExtractRegex = regexp.MustCompile(
		`^(?:https?://)?t\.me/(c/)?([\w\d_-]+)/(\d+)$`,
	)
	telegramProfileRegex = regexp.MustCompile(
		`^(?:(?:https?://)?t\.me/|@)([\w\d_-]{4,})/?$`,
	)
	telegramMsgCache = utils.NewCache[string, *telegram.NewMessage](1 * time.Hour)
	telegramDocCache = utils.NewCache[string, *telegram.DocumentObj](1 * time.Hour)
)

func init() {
	Register(&TelegramPlatform{})
}

func (t *TelegramPlatform) Name() state.PlatformName { return PlatformTelegram }
func (t *TelegramPlatform) Priority() int             { return 100 }

func (t *TelegramPlatform) CanGet(query string) bool {
	return telegramLinkRegex.MatchString(strings.TrimSpace(query))
}

func (t *TelegramPlatform) Get(query string, _ bool) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	if matches := telegramExtractRegex.FindStringSubmatch(query); len(matches) >= 4 {
		username := matches[2]
		msgID, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, errors.New("invalid message ID in Telegram link")
		}

		msg, err := core.Bot.GetMessageByID(username, int32(msgID))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch message: %w", err)
		}

		isVideo, isAudio := playableMedia(msg)
		if !isVideo && !isAudio {
			return nil, errors.New("message does not contain playable media")
		}

		track, err := t.GetTracksByMessage(msg)
		if err != nil {
			return nil, err
		}
		track.Video = isVideo
		return []*state.Track{track}, nil
	}

	if matches := telegramProfileRegex.FindStringSubmatch(query); len(matches) >= 2 {
		doc, err := t.resolveUserProfileMusic(matches[1])
		if err != nil {
			return nil, err
		}
		track, err := t.GetTrackFromDocument(doc)
		if err != nil {
			return nil, err
		}
		track.URL = "https://t.me/" + matches[1]
		telegramDocCache.Set(track.ID, doc)
		return []*state.Track{track}, nil
	}

	return nil, errors.New("invalid Telegram link")
}

func (t *TelegramPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformTelegram
}

func (t *TelegramPlatform) Download(
	ctx context.Context,
	track *state.Track,
	statusMsg *telegram.NewMessage,
) (string, error) {
	ext := ".mp3"
	if track.Video {
		ext = ".mp4"
	}
	path := getPath(track, ext)

	if fileExists(path) {
		if track.Duration == 0 {
			if dur, err := utils.GetDurationByFFProbe(path); err == nil {
				track.Duration = dur
			}
		}
		return path, nil
	}

	dOpts := &telegram.DownloadOptions{FileName: path, Ctx: ctx}
	if statusMsg != nil {
		dOpts.ProgressManager = utils.GetProgress(statusMsg)
	}

	var err error

	msg, msgOk := telegramMsgCache.Get(track.ID)
	doc, docOk := telegramDocCache.Get(track.ID)

	switch {
	case msgOk:
		path, err = msg.Download(dOpts)
	case docOk:
		path, err = core.Bot.DownloadMedia(doc, dOpts)
	default:
		file, ferr := telegram.ResolveBotFileID(track.ID)
		if ferr != nil {
			return "", fmt.Errorf("failed to locate file: %w", ferr)
		}
		path, err = core.Bot.DownloadMedia(file, dOpts)
	}

	if err != nil {
		os.Remove(path)
		if errors.Is(err, context.Canceled) {
			return "", err
		}
		return "", fmt.Errorf("download failed: %w", err)
	}

	if _, statErr := os.Stat(path); statErr != nil {
		return "", fmt.Errorf("downloaded file missing: %w", statErr)
	}

	if track.Duration == 0 {
		if dur, err := utils.GetDurationByFFProbe(path); err == nil {
			track.Duration = dur
		}
	}

	return path, nil
}

func (t *TelegramPlatform) GetTracksByMessage(msg *telegram.NewMessage) (*state.Track, error) {
	if msg == nil {
		return nil, errors.New("nil message")
	}

	target := msg
	if target.File == nil || target.File.FileID == "" {
		if msg.IsReply() {
			rmsg, err := msg.GetReplyMessage()
			if err == nil && rmsg.File != nil && rmsg.File.FileID != "" {
				target = rmsg
			}
		}
	}

	if target.File == nil || target.File.FileID == "" {
		return nil, fmt.Errorf(
			"⚠️ This <a href=\"%s\">message</a> doesn't contain any media",
			msg.Link(),
		)
	}

	file := target.File
	duration := utils.GetDuration(target.Media().(*telegram.MessageMediaDocument))
	telegramMsgCache.Set(file.FileID, target)

	return &state.Track{
		ID:       file.FileID,
		Title:    file.Name,
		Duration: duration,
		URL:      target.Link(),
		Source:   PlatformTelegram,
	}, nil
}

func (t *TelegramPlatform) GetTrackFromDocument(doc *telegram.DocumentObj) (*state.Track, error) {
	if doc == nil {
		return nil, errors.New("nil document")
	}

	track := &state.Track{
		ID:     telegram.PackBotFileID(doc),
		Source: PlatformTelegram,
	}

	var audioTitle, fileName string
	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *telegram.DocumentAttributeAudio:
			audioTitle = a.Title
			track.Duration = int(a.Duration)
		case *telegram.DocumentAttributeVideo:
			track.Video = true
			track.Duration = int(a.Duration)
		case *telegram.DocumentAttributeFilename:
			fileName = a.FileName
		}
	}

	track.Title = firstNonEmpty(audioTitle, fileName, "Telegram File")
	return track, nil
}

func (t *TelegramPlatform) resolveUserProfileMusic(username string) (*telegram.DocumentObj, error) {
	peer, err := core.Bot.ResolvePeer(username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}

	inputUser, ok := peer.(*telegram.InputPeerUser)
	if !ok {
		return nil, errors.New("resolved peer is not a user")
	}

	full, err := core.Bot.UsersGetFullUser(&telegram.InputUserObj{
		UserID:     inputUser.UserID,
		AccessHash: inputUser.AccessHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch full user: %w", err)
	}

	if full.FullUser.SavedMusic == nil {
		return nil, errors.New("user has no saved music on profile")
	}

	doc, ok := full.FullUser.SavedMusic.(*telegram.DocumentObj)
	if !ok {
		return nil, errors.New("invalid saved music document type")
	}

	return doc, nil
}
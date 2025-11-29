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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/state"
	"main/internal/utils"
)

type TelegramPlatform struct{}

var (
	telegramLinkRegex    = regexp.MustCompile(`^(https?://)?t\.me/(c/)?[\w\d_-]+/\d+$`)
	telegramExtractRegex = regexp.MustCompile(`https?://t\.me/(c/)?([\w\d_-]+)/(\d+)`)
	telegramMsgCache     = make(map[string]*telegram.NewMessage)
)

const PlatformTelegram state.PlatformName = "Telegram"

func init() {
	addPlatform(100, PlatformTelegram, &TelegramPlatform{})
}

func (t *TelegramPlatform) Name() state.PlatformName {
	return PlatformTelegram
}

func (t *TelegramPlatform) IsValid(query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return false
	}
	return telegramLinkRegex.MatchString(query)
}

func (t *TelegramPlatform) GetTracks(query string) ([]*state.Track, error) {
	if !telegramLinkRegex.MatchString(query) {
		return nil, fmt.Errorf("Provide a valid Telegram link (e.g., https://t.me/channel/12345).")
	}

	matches := telegramExtractRegex.FindStringSubmatch(query)

	if len(matches) < 4 {
		return nil, fmt.Errorf("provide a valid Telegram link (e.g., https://t.me/channel/12345)")
	}

	username := matches[2]
	messageID, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid Telegram link: bad message ID")
	}

	msg, err := core.Bot.GetMessageByID(username, int32(messageID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Telegram message: %w", err)
	}

	// process fetched message with GetTrackByMessage
	return t.GetTracksByMessage(msg)
}

func (t *TelegramPlatform) GetTracksByMessage(rmsg *telegram.NewMessage) ([]*state.Track, error) {
	file := rmsg.File
	if file == nil || file.FileID == "" {
		return nil, fmt.Errorf("⚠️ Oops! This <a href=\"%s\">message</> doesn't contain any media.", rmsg.Link())
	}

	duration := utils.GetDuration(rmsg.Media().(*telegram.MessageMediaDocument))

	telegramMsgCache[file.FileID] = rmsg

	track := &state.Track{
		ID:       file.FileID,
		Title:    file.Name,
		Duration: duration,
		URL:      rmsg.Link(),
		Source:   PlatformTelegram,
	}

	return []*state.Track{track}, nil
}

func (t *TelegramPlatform) Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
	downloadsDir := "downloads"
	if err := os.MkdirAll(downloadsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("can't create downloads folder: %v", err)
	}

	ext := ".webm"
	if ext2 := filepath.Ext(track.Title); ext2 != "" {
		ext = ext2
	}
	rawFile := filepath.Join(downloadsDir, fmt.Sprintf("%s%s", track.ID, ext))

	if path, err := findDownloadedFile(track.ID); err == nil && path != "" {
		if track.Duration == 0 {
			if dur, err := utils.GetDurationByFFProbe(path); err == nil {
				track.Duration = dur
			}
		}
		return path, nil
	}

	dOpts := &telegram.DownloadOptions{
		FileName: rawFile,
		Ctx:      ctx,
	}
	if mystic != nil {
		dOpts.ProgressManager = utils.GetProgress(mystic)
	}

	var path string
	var err error

	if msg, ok := telegramMsgCache[track.ID]; ok {
		path, err = msg.Download(dOpts)
	} else {
		file, ferr := telegram.ResolveBotFileID(track.ID)
		if ferr != nil {
			return "", fmt.Errorf("failed to locate file: %v", ferr)
		}
		path, err = core.Bot.DownloadMedia(file, dOpts)
	}

	if err != nil {
os.Remove(rawFile)
		
		if errors.Is(err, context.Canceled) {
			return "", err
		}
		return "", fmt.Errorf("download failed: %v", err)
	}

	if _, statErr := os.Stat(rawFile); statErr != nil {
		return "", fmt.Errorf("unable to get downloaded file: %v", statErr)
	}

	if track.Duration == 0 {
		if dur, err := utils.GetDurationByFFProbe(path); err == nil {
			track.Duration = dur
		}
	}

	return path, nil
}

func findDownloadedFile(id string) (string, error) {
	matches, err := filepath.Glob(filepath.Join("./downloads", id+".*"))
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		return matches[0], nil
	}
	return "", errors.New("no file found")
}

func (t *TelegramPlatform) IsDownloadSupported(source state.PlatformName) bool {
	return source == PlatformTelegram
}

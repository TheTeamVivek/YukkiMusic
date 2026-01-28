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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/utils"
)

type TelegramPlatform struct {
	name state.PlatformName
}

var (
	telegramLinkRegex = regexp.MustCompile(
		`^(https?://)?t\.me/(c/)?[\w\d_-]+/\d+$`,
	)
	telegramExtractRegex = regexp.MustCompile(
		`https?://t\.me/(c/)?([\w\d_-]+)/(\d+)`,
	)
	telegramMsgCache = make(map[string]*telegram.NewMessage)
)

const PlatformTelegram state.PlatformName = "Telegram"

func init() {
	Register(100, &TelegramPlatform{
		name: PlatformTelegram,
	})
}

func (t *TelegramPlatform) Name() state.PlatformName {
	return t.name
}

func (t *TelegramPlatform) CanGetTracks(query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return false
	}
	return telegramLinkRegex.MatchString(query)
}

func (t *TelegramPlatform) CanDownload(source state.PlatformName) bool {
	return source == t.name
}

func (t *TelegramPlatform) GetTracks(
	query string,
	_ bool,
) ([]*state.Track, error) {
	if !telegramLinkRegex.MatchString(query) {
		return nil, fmt.Errorf(
			"provide a valid Telegram link (e.g., https://t.me/channel/12345)",
		)
	}

	matches := telegramExtractRegex.FindStringSubmatch(query)
	if len(matches) < 4 {
		return nil, fmt.Errorf(
			"provide a valid Telegram link (e.g., https://t.me/channel/12345)",
		)
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

	isVideo, isAudio := playableMedia(msg)
	if !isVideo && !isAudio {
		return nil, fmt.Errorf(
			"telegram message does not contain playable media",
		)
	}

	track, err := t.GetTracksByMessage(msg)
	if err != nil {
		return nil, err
	}

	track.Video = isVideo

	if isVideo {
		if err := os.MkdirAll("cache", os.ModePerm); err != nil {
			gologging.Error("Failed to create cache folder: " + err.Error())
			return []*state.Track{track}, nil
		}

		thumbPath := filepath.Join("cache", "thumb_"+track.ID+".jpg")
		if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
			path, err := msg.Download(&telegram.DownloadOptions{
				ThumbOnly: true,
				FileName:  thumbPath,
			})
			if err == nil {
				if _, err := os.Stat(path); err == nil {
					track.Artwork = path
				}
			}
		}
	}

	return []*state.Track{track}, nil
}

func (*TelegramPlatform) CanSearch() bool { return false }

func (*TelegramPlatform) Search(
	string,
	bool,
) ([]*state.Track, error) {
	return nil, nil
}

func (t *TelegramPlatform) GetTracksByMessage(
	msg *telegram.NewMessage,
) (*state.Track, error) {
	if msg == nil {
		return nil, fmt.Errorf("invalid telegram message")
	}

	var target *telegram.NewMessage

	// First: check current message itself
	if msg.File != nil && msg.File.FileID != "" {
		target = msg
	} else {
		// Second: if no media in current message, check reply
		if msg.IsReply() {
			rmsg, err := msg.GetReplyMessage()
			if err == nil {
				if rmsg.File != nil && rmsg.File.FileID != "" {
					target = rmsg
				}
			}
		}
	}

	// Still nothing? Then fail.
	if target == nil {
		return nil, fmt.Errorf(
			"⚠️ Oops! This <a href=\"%s\">message</a> doesn't contain any media",
			msg.Link(),
		)
	}

	file := target.File

	duration := utils.GetDuration(
		target.Media().(*telegram.MessageMediaDocument),
	)

	telegramMsgCache[file.FileID] = target

	track := &state.Track{
		ID:       file.FileID,
		Title:    file.Name,
		Duration: duration,
		URL:      target.Link(),
		Source:   PlatformTelegram,
	}

	return track, nil
}

func (t *TelegramPlatform) Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	path := getPath(track, ".mp3")
	if track.Video {
		path = getPath(track, ".mp4")
	}

	if fileExists(path) {
		if track.Duration == 0 {
			if dur, err := utils.GetDurationByFFProbe(path); err == nil {
				track.Duration = dur
			}
		}
		return path, nil
	}

	dOpts := &telegram.DownloadOptions{
		FileName: path,
		Ctx:      ctx,
	}
	if mystic != nil {
		dOpts.ProgressManager = utils.GetProgress(mystic)
	}
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
		os.Remove(path)

		if errors.Is(err, context.Canceled) {
			return "", err
		}
		return "", fmt.Errorf("download failed: %v", err)
	}

	if _, statErr := os.Stat(path); statErr != nil {
		return "", fmt.Errorf("unable to get downloaded file: %v", statErr)
	}

	if track.Duration == 0 {
		if dur, err := utils.GetDurationByFFProbe(path); err == nil {
			track.Duration = dur
		}
	}

	return path, nil
}

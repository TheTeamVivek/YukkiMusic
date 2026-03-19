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
	"time"

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
		`^(https?://)?t\.me/((c/)?[\w\d_-]+/\d+|[\w\d_-]+)$`,
	)
	telegramExtractRegex = regexp.MustCompile(
		`^(?:https?://)?t\.me/(c/)?([\w\d_-]+)/(\d+)$`,
	)
	telegramProfileRegex = regexp.MustCompile(
		`^(?:https?://)?t\.me/([\w\d_-]{4,})/?$`,
	)
	telegramMsgCache = utils.NewCache[string, *telegram.NewMessage](1 * time.Hour)
	telegramDocCache = utils.NewCache[string, *telegram.DocumentObj](1 * time.Hour)
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
	query = strings.TrimSpace(query)
	if !telegramLinkRegex.MatchString(query) {
		return nil, fmt.Errorf(
			"provide a valid Telegram link (e.g., https://t.me/channel/12345 or https://t.me/username)",
		)
	}

	if matches := telegramExtractRegex.FindStringSubmatch(query); len(matches) >= 4 {
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

	if matches := telegramProfileRegex.FindStringSubmatch(query); len(matches) >= 2 {
		username := matches[1]
		doc, err := t.resolveUserProfileMusic(username)
		if err != nil {
			return nil, err
		}

		track, err := t.GetTrackFromDocument(doc)
		if err != nil {
			return nil, err
		}
track.URL = "https://t.me/" + username
		telegramDocCache.Set(track.ID, doc)

		return []*state.Track{track}, nil
	}

	return nil, fmt.Errorf("invalid Telegram link")
}

func (t *TelegramPlatform) GetTrackFromDocument(doc *telegram.DocumentObj) (*state.Track, error) {
	d := doc
	if d == nil {
		return nil, fmt.Errorf("invalid document")
	}

	track := &state.Track{
		ID:     telegram.PackBotFileID(d),
		Source: PlatformTelegram,
	}

	var audioTitle, fileName string
	for _, attr := range d.Attributes {
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

	if audioTitle != "" {
		track.Title = audioTitle
	} else if fileName != "" {
		track.Title = fileName
	}

	if track.Title == "" {
		track.Title = "Telegram File"
	}

	return track, nil
}

func (t *TelegramPlatform) resolveUserProfileMusic(username string) (*telegram.DocumentObj, error) {
	peer, err := core.Bot.ResolvePeer(username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}

	inputUser, ok := peer.(*telegram.InputPeerUser)
	if !ok {
		return nil, fmt.Errorf("resolved peer is not a user")
	}

	fullUser, err := core.Bot.UsersGetFullUser(&telegram.InputUserObj{
		UserID:     inputUser.UserID,
		AccessHash: inputUser.AccessHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch full user info: %w", err)
	}

	if fullUser.FullUser.SavedMusic == nil {
		return nil, fmt.Errorf("user does not have any saved music on profile")
	}

	doc, ok := fullUser.FullUser.SavedMusic.(*telegram.DocumentObj)
	if !ok {
		return nil, fmt.Errorf("invalid saved music document type")
	}

	return doc, nil
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

	telegramMsgCache.Set(file.FileID, target)

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
	statusMsg *telegram.NewMessage,
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
	if statusMsg != nil {
		dOpts.ProgressManager = utils.GetProgress(statusMsg)
	}
	var err error

	msg, msgOk := telegramMsgCache.Get(track.ID)
	doc, docOk := telegramDocCache.Get(track.ID)

	if msgOk {
		path, err = msg.Download(dOpts)
	} else if docOk {
		path, err = core.Bot.DownloadMedia(doc, dOpts)
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


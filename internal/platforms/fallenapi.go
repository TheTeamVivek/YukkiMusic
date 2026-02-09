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
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/utils"
)

var telegramDLRegex = regexp.MustCompile(
	`https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)`,
)

const PlatformFallenApi state.PlatformName = "FallenApi"

type apiResponse struct {
	CdnUrl string `json:"cdnurl"`
}

type FallenApiPlatform struct {
	name state.PlatformName
}

func init() {
	Register(80, &FallenApiPlatform{
		name: PlatformFallenApi,
	})
}

func (f *FallenApiPlatform) Name() state.PlatformName {
	return f.name
}

func (f *FallenApiPlatform) CanGetTracks(query string) bool {
	return false
}

func (f *FallenApiPlatform) GetTracks(
	_ string,
	_ bool,
) ([]*state.Track, error) {
	return nil, errors.New("fallenapi is a download-only platform")
}

func (f *FallenApiPlatform) CanDownload(
	source state.PlatformName,
) bool {
	if config.FallenAPIURL == "" || config.FallenAPIKey == "" {
		return false
	}
	return source == PlatformYouTube
}

func (f *FallenApiPlatform) Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	// fallen api didn't support video downloads so disable it
	track.Video = false

	if f := findFile(track); f != "" {
		gologging.Debug("FallenApi: Download -> Cached File -> " + f)
		return f, nil
	}

	var pm *telegram.ProgressManager
	if mystic != nil {
		pm = utils.GetProgress(mystic)
	}

	dlURL, err := f.getDownloadURL(ctx, track.URL)
	if err != nil {
		return "", err
	}

	path := getPath(track, ".mp3")

	var downloadErr error
	if telegramDLRegex.MatchString(dlURL) {
		path, downloadErr = f.downloadFromTelegram(ctx, dlURL, path, pm)
	} else {
		downloadErr = f.downloadFromURL(ctx, dlURL, path)
	}

	if downloadErr != nil {
		return "", downloadErr
	}
	if !fileExists(path) {
		return "", errors.New("empty file returned by API")
	}
	return path, nil
}

func (f *FallenApiPlatform) CanGetRecommendations() bool {
	return false
}

func (f *FallenApiPlatform) GetRecommendations(
	track *state.Track,
) ([]*state.Track, error) {
	return nil, errors.New("recommendations not supported on fallenapi")
}

func (*FallenApiPlatform) CanSearch() bool { return false }

func (*FallenApiPlatform) Search(
	string,
	bool,
) ([]*state.Track, error) {
	return nil, nil
}

func (f *FallenApiPlatform) getDownloadURL(
	ctx context.Context,
	mediaURL string,
) (string, error) {
	apiReqURL := fmt.Sprintf(
		"%s/api/track?api_key=%s&url=%s",
		config.FallenAPIURL,
		config.FallenAPIKey,
		url.QueryEscape(mediaURL),
	)

	var apiResp apiResponse

	resp, err := rc.R().
		SetContext(ctx).
		SetResult(&apiResp).
		Get(apiReqURL)
	if err != nil {
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		return "", fmt.Errorf(
			"failed to download %s, api request failed: %w",mediaURL,
			sanitizeAPIError(err, config.FallenAPIKey),
		)
	}

	if resp.IsError() {
		err = fmt.Errorf(
			"failed to download %s, api request failed with status: %d body: %s", mediaURL,
			resp.StatusCode(),
			resp.String(),
		)
		gologging.Error(err.Error())
		return "", err
	}

	if apiResp.CdnUrl == "" {
		err = fmt.Errorf("failed to download %s, empty API response body: %s", mediaURL, resp.String())
		gologging.Error(err.Error())
		return "", err
	}

	return apiResp.CdnUrl, nil
}

func (f *FallenApiPlatform) downloadFromURL(
	ctx context.Context,
	dlURL, path string,
) error {
	resp, err := rc.R().
		SetContext(ctx).
		SetOutputFileName(path).
		Get(dlURL)
	if err != nil {
		os.Remove(path)
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("http download failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode())
	}

	return nil
}

func (f *FallenApiPlatform) downloadFromTelegram(
	ctx context.Context,
	dlURL, path string,
	pm *telegram.ProgressManager,
) (string, error) {
	matches := telegramDLRegex.FindStringSubmatch(dlURL)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid telegram download url: %s", dlURL)
	}

	username := matches[1]
	messageID, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid message ID: %v", err)
	}

	msg, err := core.Bot.GetMessageByID(username, int32(messageID))
	if err != nil {
		return "", fmt.Errorf("failed to fetch Telegram message: %w", err)
	}

	dOpts := &telegram.DownloadOptions{
		FileName: path,
		Ctx:      ctx,
	}
	if pm != nil {
		dOpts.ProgressManager = pm
	}
	_, err = msg.Download(dOpts)
	if err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

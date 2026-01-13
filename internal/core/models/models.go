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

package state

import (
	"context"
	"os"
	"path/filepath"

	"github.com/amarnathcjd/gogram/telegram"
)

type (
	Track struct {
		ID        string       // track unique id
		Title     string       // title
		Duration  int          // track duration in seconds
		Artwork   string       // thumbnail url of the track
		URL       string       // track url
		Requester string       // html mention or @username who requested this track
		Video     bool         // whether this track will be played as video
		Source    PlatformName // unique PlatformName
	}
	PlatformName string

	Platform interface {
		Close() // cleanup
		Name() PlatformName
		CanGetTracks(query string) bool
		GetTracks(query string, video bool) ([]*Track, error)
		Download(
			ctx context.Context,
			track *Track,
			mystic *telegram.NewMessage,
		) (string, error)
		CanDownload(source PlatformName) bool
	}
)

// returns filepath where song should be Downloaded
func (t *Track) FilePath() string {
	if t == nil {
		return ""
	}

	_ = os.MkdirAll("downloads", os.ModePerm)

	if t.Video {
		return filepath.Join("downloads", "video_"+t.ID+".mp4")
	}
	return filepath.Join("downloads", "audio_"+t.ID+".m4a")
}

// returns true if the track is downloaded
func (t *Track) IsExists() bool {
	if t == nil {
		return false
	}

	info, err := os.Stat(t.FilePath())
	return err == nil && info.Size() > 0
}

// remove the track if downloaded
func (t *Track) Remove() (r bool) {
	if t != nil {
		err := os.Remove(t.FilePath())
		r = err == nil
	}
	return r
}

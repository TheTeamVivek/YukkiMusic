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

	"github.com/amarnathcjd/gogram/telegram"
)

type (
  Track struct {
	  ID         string        // track unique id
	  Title      string        // title
	  Duration   int           // track duration in seconds
  	Artwork    string        // thumbnail url of the track
  	URL        string        // track url
	  Requester  string        // html mention or @username who requested this track
  	Video      bool          // whether this track will be played as video
  	Source     PlatformName  // unique PlatformName
  }
	PlatformName string

	Platform interface {
		Name() PlatformName
		IsValid(query string) bool
		GetTracks(query string) ([]*Track, error)
		Download(ctx context.Context, track *Track, mystic *telegram.NewMessage) (string, error)
		IsDownloadSupported(source PlatformName) bool
	}
	
 PlayOpts struct {
	Force bool
	CPlay bool
	Video bool
}
)

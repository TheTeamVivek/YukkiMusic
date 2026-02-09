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
	// Platform defines a common contract for all supported platforms
	// (e.g. YouTube, SoundCloud, Spotify, etc.).
	//
	// Each platform is responsible for determining whether it can
	// search, resolve, or download tracks from a given query or source.
	Platform interface {
		// Name returns the unique identifier of the platform.
		Name() PlatformName

		// CanSearch reports whether this platform supports search.
		CanSearch() bool

		// Search searches the platform for tracks matching the query.
		//
		// query: the search string
		// video:
		//   - If the platform supports both audio and video, propagate this
		//     value into Track.Video
		//   - If the platform is audio-only, always set Track.Video = false
		//   - If the platform is video-only, always set Track.Video = true
		//
		// This method is primarily used for video playback workflows.
		Search(query string, video bool) ([]*Track, error)

		// CanDownload reports whether this platform can download tracks
		// originating from the given source platform.
		CanDownload(source PlatformName) bool

		// Download downloads the given track and returns the local file path.
		//
		// ctx is used for cancellation and timeouts.
		// track is the track to download.
		// mystic used to send progress updates (if not nil).
		// if your platform support video playback so return local path of video when track.Video is true
		Download(
			ctx context.Context,
			track *Track,
			mystic *telegram.NewMessage,
		) (string, error)

		// CanGetTracks reports whether this platform can resolve
		// tracks from the given query search term.
		CanGetTracks(query string) bool

		// GetTracks fetches track metadata for the given query.
		//
		// video indicates whether video playback is requested.
		// Platforms that do not support video should still return tracks,
		// but must set Track.Video = false.
		GetTracks(query string, video bool) ([]*Track, error)

		// CanGetRecommendations reports whether this platform can
		// provide track recommendations based on a given track.
		CanGetRecommendations() bool

		// GetRecommendations fetches recommended tracks for the given track.
		GetRecommendations(track *Track) ([]*Track, error)
	}
)

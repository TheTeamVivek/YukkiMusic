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

package models

import (
	"context"

	"github.com/amarnathcjd/gogram/telegram"
)

type (
	PlatformName string

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

	// Platform defines a common contract for all supported platforms
	// (e.g. YouTube, SoundCloud, Spotify, etc.).
	//
	// Each platform is responsible for determining whether it can
	// search, resolve, or download tracks from a given query or source.
	Platform interface {
		// Name returns the unique identifier of the platform.
		Name() PlatformName

		// Priority returns the priority of this platform when multiple
		// platforms can handle the same query. Higher values take precedence.
		Priority() int

		// CanGet reports whether this platform can resolve
		// tracks from the given query or search term.
		CanGet(query string) bool

		// Get fetches track metadata for the given query.
		//
		// video indicates whether video playback is requested.
		// Platforms that do not support video should still return tracks,
		// but must set Track.Video = false.
		Get(query string, video bool) ([]*Track, error)

		// CanDownload reports whether this platform can download tracks
		// originating from the given source platform.
		CanDownload(source PlatformName) bool

		// Download downloads the given track and returns the local file path.
		//
		// ctx is used for cancellation and timeouts.
		// track is the track to download.
		// msg is used to send progress updates (if not nil).
		// If the platform supports video playback, return the local path
		// of the video file when track.Video is true.
		Download(ctx context.Context, track *Track, msg *telegram.NewMessage) (string, error)
	}
)
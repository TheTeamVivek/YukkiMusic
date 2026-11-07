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
	"slices"
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
	"resty.dev/v3"

	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/utils"
)

var (
	// ErrNoTracks means the search ran fine but found nothing playable.
	ErrNoTracks = errors.New("no tracks found")
	// ErrNotMedia means the replied message holds no audio or video.
	ErrNotMedia = errors.New("replied message is not playable media")

	// rc serves API calls; dl serves file downloads, which can outlast any
	// fixed timeout and are stopped through their context instead.
	rc = resty.New().SetTimeout(20 * time.Second)
	dl = resty.New()

	mu         sync.RWMutex
	registered []state.Platform // highest priority first
)

func Register(p state.Platform) {
	mu.Lock()
	defer mu.Unlock()

	registered = slices.DeleteFunc(registered, func(q state.Platform) bool {
		return q.Name() == p.Name()
	})
	registered = append(registered, p)
	slices.SortStableFunc(registered, func(a, b state.Platform) int {
		return b.Priority() - a.Priority()
	})
}

func GetPlatform(name state.PlatformName) (state.Platform, bool) {
	for _, p := range ordered() {
		if p.Name() == name {
			return p, true
		}
	}
	return nil, false
}

func ordered() []state.Platform {
	mu.RLock()
	defer mu.RUnlock()
	return slices.Clone(registered)
}

func GetTracks(c *td.Client, m *td.Message, video bool) ([]*state.Track, error) {
	urls, _ := utils.ExtractURLs(c, m)
	if len(urls) > 0 {
		tracks, errs := fetchFromURLs(urls, video)
		if len(tracks) > 0 {
			return tracks, nil
		}
		if !hasPlayableReply(c, m) {
			return nil, combineErrs("no supported platform for given URL(s)", errs)
		}
	}

	// A query that is just a failed link would only turn into a junk text search.
	var searchErr error
	if q := m.Args(); q != "" && !slices.ContainsFunc(urls, func(u string) bool {
		return strings.Contains(q, u)
	}) {
		tracks, err := searchQuery(q, video)
		if err == nil && len(tracks) > 0 {
			return tracks, nil
		}
		if err != nil && !errors.Is(err, ErrNoTracks) {
			searchErr = err
		}
	}

	if m.ReplyToMessageID() > 0 {
		return fromReply(c, m)
	}

	if searchErr != nil {
		return nil, fmt.Errorf("search failed: %w", searchErr)
	}
	return nil, ErrNoTracks
}

func Download(
	ctx context.Context,
	track *state.Track,
	msg *td.Message,
) (string, error) {
	var errs []string

	for _, p := range ordered() {
		if !p.CanDownload(track.Source) {
			continue
		}

		path, err := p.Download(ctx, track, msg)
		if err == nil {
			utils.Log.Info("Download ok via " + string(p.Name()) + " -> " + path)
			return path, nil
		}
		if errors.Is(err, context.Canceled) {
			return "", err
		}
		errs = append(errs, string(p.Name())+": "+err.Error())
	}

	if len(errs) > 0 {
		return "", combineErrs("download failed", errs)
	}
	return "", errors.New("no downloader for source: " + string(track.Source))
}

func AutoplayTracks(last *state.Track, limit int) ([]*state.Track, error) {
	if last == nil || last.ID == "" || last.Source != PlatformYouTube {
		return nil, errors.New("autoplay requires a YouTube track")
	}

	tracks, err := youtube.fetchMixPlaylist("RD"+last.ID, limit)
	if err != nil {
		return nil, err
	}
	return withVideo(tracks, last.Video), nil
}

// getFromURL asks every platform that accepts the link, best first, until one
// resolves it. errs stays empty when no platform accepted the link.
func getFromURL(link string, video bool) (tracks []*state.Track, errs []string) {
	for _, p := range ordered() {
		if !p.CanGet(link) {
			continue
		}

		got, err := p.Get(link, video)
		if err != nil {
			errs = append(errs, string(p.Name())+": "+err.Error())
			continue
		}
		if len(got) > 0 {
			return got, nil
		}
	}
	return nil, errs
}

func fetchFromURLs(urls []string, video bool) (tracks []*state.Track, errs []string) {
	for _, u := range urls {
		got, e := getFromURL(u, video)
		if len(got) == 0 && len(e) == 0 {
			e = []string{"no platform for: " + u}
		}
		tracks = append(tracks, got...)
		errs = append(errs, e...)
	}
	return tracks, errs
}

// searchQuery resolves a link typed as the query, or runs a YouTube text
// search and keeps the first hit.
func searchQuery(q string, video bool) ([]*state.Track, error) {
	tracks, errs := getFromURL(q, video)
	if len(tracks) > 0 {
		return tracks, nil
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}

	tracks, err := youtube.Get(q, video)
	if err != nil {
		return nil, err
	}
	return tracks[:1], nil
}

func fromReply(c *td.Client, m *td.Message) ([]*state.Track, error) {
	target, isVideo, _ := playableMedia(c, m)
	if target == nil {
		return nil, ErrNotMedia
	}

	track, err := telegram.GetTracksByMessage(c, target)
	if err != nil {
		return nil, err
	}
	track.Video = isVideo
	return []*state.Track{track}, nil
}

func hasPlayableReply(c *td.Client, m *td.Message) bool {
	if m.ReplyToMessageID() <= 0 {
		return false
	}
	msg, _, _ := playableMedia(c, m)
	return msg != nil
}

func combineErrs(prefix string, errs []string) error {
	if len(errs) == 0 {
		return errors.New(prefix)
	}
	return errors.New(prefix + "\n• " + strings.Join(errs, "\n• "))
}

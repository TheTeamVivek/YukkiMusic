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
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
	"resty.dev/v3"
	"yukkimusic/internal/logger"

	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"
)

type RedispatchError struct {
	Track *state.Track
}

func (e *RedispatchError) Error() string {
	return "redispatch:" + string(e.Track.Source)
}

type reg struct {
	mu     sync.RWMutex
	sorted []state.Platform
	byName map[state.PlatformName]state.Platform
}

var (
	global = &reg{
		byName: make(map[state.PlatformName]state.Platform),
	}
	rc = resty.New().SetTimeout(20 * time.Second)
)

func Register(p state.Platform) {
	global.mu.Lock()
	defer global.mu.Unlock()

	global.byName[p.Name()] = p
	global.sorted = append(global.sorted, p)
	sort.Slice(global.sorted, func(i, j int) bool {
		return global.sorted[i].Priority() > global.sorted[j].Priority()
	})
}

func GetPlatform(name state.PlatformName) (state.Platform, bool) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	p, ok := global.byName[name]
	return p, ok
}

func ordered() []state.Platform {
	global.mu.RLock()
	defer global.mu.RUnlock()
	out := make([]state.Platform, len(global.sorted))
	copy(out, global.sorted)
	return out
}

func findFor(query string) state.Platform {
	for _, p := range ordered() {
		if p.CanGet(query) {
			return p
		}
	}
	return nil
}

func GetTracks(c *td.Client, m *td.Message, video bool) ([]*state.Track, error) {
	logger.Debug("GetTracks | video:" + strconv.FormatBool(video))

	if urls, _ := utils.ExtractURLs(c, m); len(urls) > 0 {
		tracks, errs := fetchFromURLs(urls, video)
		if len(tracks) > 0 {
			return tracks, nil
		}
		if !hasPlayableReply(c, m) {
			return nil, combineErrs("no supported platform for given URL(s)", errs)
		}
	}

	if q := m.Args(); q != "" {
		tracks, err := searchQuery(q, video)
		if err == nil && len(tracks) > 0 {
			return tracks, nil
		}
	}

	if m.ReplyToMessageID() > 0 {
		return fromReply(c, m)
	}

	return nil, errors.New("no tracks found")
}

func Download(
	ctx context.Context,
	track *state.Track,
	msg *td.Message,
) (string, error) {
	return download(ctx, track, msg, false)
}

func download(
	ctx context.Context,
	track *state.Track,
	msg *td.Message,
	redispatched bool,
) (string, error) {
	var errs []string

	for _, p := range ordered() {
		if !p.CanDownload(track.Source) {
			continue
		}

		logger.Debug("Download attempt: " + string(p.Name()))
		path, err := p.Download(ctx, track, msg)
		if err == nil {
			logger.Info("Download ok via " + string(p.Name()) + " -> " + path)
			return path, nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		var rd *RedispatchError
		if !redispatched && errors.As(err, &rd) {
			logger.Debug("Redispatch to source: " + string(rd.Track.Source))
			return download(ctx, rd.Track, msg, true)
		}

		errs = append(errs, string(p.Name())+": "+err.Error())
	}

	if len(errs) > 0 {
		return "", combineErrs("download failed", errs)
	}
	return "", errors.New("no downloader for source: " + string(track.Source))
}

func fetchFromURLs(urls []string, video bool) ([]*state.Track, []string) {
	var tracks []*state.Track
	var errs []string

	for _, u := range urls {
		p := findFor(u)
		if p == nil {
			errs = append(errs, "no platform for: "+u)
			continue
		}

		logger.Debug("URL matched " + string(p.Name()) + ": " + u)
		got, err := p.Get(u, video)
		if err != nil {
			if strings.Contains(err.Error(), "failed to extract metadata") {
				continue
			}
			errs = append(errs, string(p.Name())+": "+err.Error())
			continue
		}
		tracks = append(tracks, got...)
	}

	return tracks, errs
}

func searchQuery(q string, video bool) ([]*state.Track, error) {
	if p := findFor(q); p != nil && p.Name() != PlatformYouTube {
		got, err := p.Get(q, video)
		if err == nil && len(got) > 0 {
			return got, nil
		}
	}

	yt, ok := GetPlatform(PlatformYouTube)
	if !ok {
		return nil, errors.New("youtube platform not registered")
	}

	tracks, err := yt.Get(q, video)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, nil
	}
	return []*state.Track{tracks[0]}, nil
}

func fromReply(c *td.Client, m *td.Message) ([]*state.Track, error) {
	target, isVideo, err := mediaInReply(c, m)
	if err != nil {
		return nil, err
	}

	tgp, ok := GetPlatform(PlatformTelegram)
	if !ok {
		return nil, errors.New("telegram platform not registered")
	}

	track, err := tgp.(*TelegramPlatform).GetTracksByMessage(c, target)
	if err != nil {
		return nil, err
	}

	track.Video = isVideo

	if isVideo {
		noThumb, err := database.ThumbnailsDisabled(m.ChatID())
		if err != nil || !noThumb {
			downloadThumbnail(c, target, track)
		}
	}

	return []*state.Track{track}, nil
}

func mediaInReply(c *td.Client, m *td.Message) (*td.Message, bool, error) {
	curr, err := m.GetRepliedMessage(c)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get reply: %w", err)
	}

	for range 2 {
		if v, a := playableMedia(c, curr); v || a {
			return curr, v, nil
		}
		if curr.ReplyToMessageID() <= 0 {
			break
		}
		next, err := curr.GetRepliedMessage(c)
		if err != nil {
			break
		}
		curr = next
	}

	return nil, false, errors.New("⚠️ Reply with a valid media (audio/video)")
}

func downloadThumbnail(c *td.Client, m *td.Message, t *state.Track) {
	if err := os.MkdirAll("cache", os.ModePerm); err != nil {
		return
	}
	dest := filepath.Join("cache", "thumb_"+t.ID+".jpg")
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		video, ok := m.Content.(*td.MessageVideo)
		if !ok || video.Video == nil || video.Video.Thumbnail == nil ||
			video.Video.Thumbnail.File == nil {
			return
		}
		thumbFile, err := video.Video.Thumbnail.File.Download(
			c, 0, 0, 1, &td.DownloadFileOpts{Synchronous: true},
		)
		if err != nil || thumbFile == nil || thumbFile.Local == nil || thumbFile.Local.Path == "" {
			return
		}
		if err := copyFile(thumbFile.Local.Path, dest); err != nil {
			return
		}
		t.Artwork = dest
	} else {
		t.Artwork = dest
	}
}

func hasPlayableReply(c *td.Client, m *td.Message) bool {
	if m.ReplyToMessageID() <= 0 {
		return false
	}
	rmsg, err := m.GetRepliedMessage(c)
	if err != nil {
		return false
	}
	v, a := playableMedia(c, rmsg)
	return v || a
}

func combineErrs(prefix string, errs []string) error {
	if len(errs) == 0 {
		return errors.New(prefix)
	}
	return errors.New(prefix + "\n• " + strings.Join(errs, "\n• "))
}

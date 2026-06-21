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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"main/config"
	state "main/internal/core/models"
	"main/internal/utils"
)

const PlatformSpotify state.PlatformName = "Spotify"

type SpotifyPlatform struct {
	client     *spotify.Client
	clientOnce sync.Once
	initErr    error
}

var (
	spotifyTrackRe    = regexp.MustCompile(`(?i)spotify\.com/track/([a-zA-Z0-9]+)`)
	spotifyPlaylistRe = regexp.MustCompile(`(?i)spotify\.com/playlist/([a-zA-Z0-9]+)`)
	spotifyAlbumRe    = regexp.MustCompile(`(?i)spotify\.com/album/([a-zA-Z0-9]+)`)
	spotifyArtistRe   = regexp.MustCompile(`(?i)spotify\.com/artist/([a-zA-Z0-9]+)`)
	spotifyLinkRe     = regexp.MustCompile(`(?i)^(https?:\/\/)?(open\.)?spotify\.com\/`)
	nameCleanupRe     = regexp.MustCompile(`[\(\[].*?[\)\]]`)
	spotifyCache      = utils.NewCache[string, []*state.Track](1 * time.Hour)
)

func init() {
	Register(&SpotifyPlatform{})
}

func (s *SpotifyPlatform) Name() state.PlatformName { return PlatformSpotify }
func (s *SpotifyPlatform) Priority() int            { return 95 }

func (s *SpotifyPlatform) CanGet(query string) bool {
	return spotifyLinkRe.MatchString(query)
}

func (s *SpotifyPlatform) Get(query string, video bool) ([]*state.Track, error) {
	if config.SpotifyClientID == "" || config.SpotifyClientSecret == "" {
		return nil, errors.New("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not configured")
	}

	cacheKey := "spotify:" + strings.ToLower(query)
	if cached, ok := spotifyCache.Get(cacheKey); ok {
		return withVideo(cached, video), nil
	}

	if err := s.ensureClient(); err != nil {
		return nil, fmt.Errorf("spotify client init failed: %w", err)
	}

	ctx := context.Background()
	var (
		tracks []*state.Track
		err    error
	)

	switch {
	case spotifyTrackRe.MatchString(query):
		tracks, err = s.getTrack(ctx, spotify.ID(spotifyTrackRe.FindStringSubmatch(query)[1]))
	case spotifyPlaylistRe.MatchString(query):
		tracks, err = s.getPlaylist(ctx, spotify.ID(spotifyPlaylistRe.FindStringSubmatch(query)[1]))
	case spotifyAlbumRe.MatchString(query):
		tracks, err = s.getAlbum(ctx, spotify.ID(spotifyAlbumRe.FindStringSubmatch(query)[1]))
	case spotifyArtistRe.MatchString(query):
		tracks, err = s.getArtistTop(ctx, spotify.ID(spotifyArtistRe.FindStringSubmatch(query)[1]))
	default:
		return nil, errors.New("unsupported Spotify URL (track/playlist/album/artist only)")
	}

	if err != nil {
		return nil, err
	}

	if len(tracks) > 0 {
		spotifyCache.Set(cacheKey, tracks)
	}

	return withVideo(tracks, video), nil
}

func (s *SpotifyPlatform) CanDownload(source state.PlatformName) bool {
	return source == PlatformSpotify
}

// Download resolves the Spotify track to a YouTube URL and redispatches.
// The registry's Download loop will then pick FallenApi or YtDlp.
func (s *SpotifyPlatform) Download(
	_ context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	ytTrack, err := s.resolveToYouTube(track)
	if err != nil {
		return "", err
	}
	return "", &RedispatchError{Track: ytTrack}
}

func (s *SpotifyPlatform) resolveToYouTube(track *state.Track) (*state.Track, error) {
	yt, ok := GetPlatform(PlatformYouTube)
	if !ok {
		return nil, errors.New("youtube platform not registered")
	}

	ytp := yt.(*YouTubePlatform)

	queries := buildSearchQueries(track.Title)
	for i, q := range queries {
		gologging.DebugF("[Spotify→YouTube] attempt %d: %q", i+1, q)
		results, err := ytp.VideoSearch(q, true)
		if err != nil || len(results) == 0 {
			continue
		}
		t := results[0]
		t.Video = track.Video
		gologging.DebugF("[Spotify→YouTube] matched %q → %s", q, t.URL)
		return t, nil
	}

	return nil, errors.New("could not find track on YouTube")
}

func buildSearchQueries(title string) []string {
	clean := cleanTitle(title)
	trimmed := trimTitle(clean, 25, 40)

	seen := map[string]bool{}
	var out []string
	for _, q := range []string{clean, trimmed, trimTitle(title, 25, 40)} {
		if q != "" && !seen[q] {
			seen[q] = true
			out = append(out, q)
		}
	}
	return out
}

func (s *SpotifyPlatform) ensureClient() error {
	s.clientOnce.Do(func() {
		cfg := &clientcredentials.Config{
			ClientID:     config.SpotifyClientID,
			ClientSecret: config.SpotifyClientSecret,
			TokenURL:     spotifyauth.TokenURL,
		}
		token, err := cfg.Token(context.Background())
		if err != nil {
			s.initErr = fmt.Errorf("spotify token: %w", err)
			return
		}
		s.client = spotify.New(spotifyauth.New().Client(context.Background(), token))
	})
	return s.initErr
}

func (s *SpotifyPlatform) getTrack(ctx context.Context, id spotify.ID) ([]*state.Track, error) {
	ft, err := s.client.GetTrack(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("spotify GetTrack: %w", err)
	}
	return []*state.Track{s.convert(&ft.SimpleTrack, ft.Album.Images)}, nil
}

func (s *SpotifyPlatform) getPlaylist(ctx context.Context, id spotify.ID) ([]*state.Track, error) {
	var tracks []*state.Track
	offset := 0
	for {
		page, err := s.client.GetPlaylistItems(ctx, id, spotify.Limit(100), spotify.Offset(offset))
		if err != nil {
			return nil, fmt.Errorf("spotify GetPlaylist: %w", err)
		}
		for _, item := range page.Items {
			if item.Track.Track == nil {
				continue
			}
			tracks = append(tracks, s.convert(&item.Track.Track.SimpleTrack, item.Track.Track.Album.Images))
		}
		if page.Next == "" {
			break
		}
		offset += 100
	}
	return tracks, nil
}

func (s *SpotifyPlatform) getAlbum(ctx context.Context, id spotify.ID) ([]*state.Track, error) {
	album, err := s.client.GetAlbum(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("spotify GetAlbum: %w", err)
	}
	tracks := make([]*state.Track, 0, len(album.Tracks.Tracks))
	for _, st := range album.Tracks.Tracks {
		tracks = append(tracks, s.convert(&st, album.Images))
	}
	return tracks, nil
}

func (s *SpotifyPlatform) getArtistTop(ctx context.Context, id spotify.ID) ([]*state.Track, error) {
	top, err := s.client.GetArtistsTopTracks(ctx, id, "US")
	if err != nil {
		return nil, fmt.Errorf("spotify GetArtistTopTracks: %w", err)
	}
	if len(top) == 0 {
		return nil, errors.New("no tracks found for artist")
	}
	tracks := make([]*state.Track, 0, len(top))
	for _, ft := range top {
		tracks = append(tracks, s.convert(&ft.SimpleTrack, ft.Album.Images))
	}
	return tracks, nil
}

func (s *SpotifyPlatform) convert(st *spotify.SimpleTrack, images []spotify.Image) *state.Track {
	thumb := ""
	if len(images) > 0 {
		thumb = images[0].URL
	}
	return &state.Track{
		ID:       string(st.ID),
		Title:    st.Name,
		Duration: int(st.Duration) / 1000,
		Artwork:  thumb,
		URL:      st.ExternalURLs["spotify"],
		Source:   PlatformSpotify,
	}
}

func withVideo(tracks []*state.Track, video bool) []*state.Track {
	out := make([]*state.Track, 0, len(tracks))
	for _, t := range tracks {
		if t == nil {
			continue
		}
		clone := *t
		clone.Video = video
		out = append(out, &clone)
	}
	return out
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	title = nameCleanupRe.ReplaceAllString(title, "")
	return strings.Join(strings.Fields(title), " ")
}

func trimTitle(title string, min, max int) string {
	title = strings.TrimSpace(title)
	runes := []rune(title)
	if len(runes) <= max {
		return title
	}
	cut := max
	for i := max - 1; i >= min; i-- {
		if runes[i] == ' ' {
			cut = i
			break
		}
	}
	return strings.TrimSpace(string(runes[:cut]))
}

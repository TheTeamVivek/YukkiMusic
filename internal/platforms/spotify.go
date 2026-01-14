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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"main/internal/config"
	state "main/internal/core/models"
	"main/internal/utils"
)

type SpotifyPlatform struct {
	name       state.PlatformName
	client     *spotify.Client
	clientOnce sync.Once
	initErr    error
}

var (
	spotifyTrackRegex = regexp.MustCompile(
		`(?i)spotify\.com/track/([a-zA-Z0-9]+)`,
	)
	spotifyPlaylistRegex = regexp.MustCompile(
		`(?i)spotify\.com/playlist/([a-zA-Z0-9]+)`,
	)
	spotifyAlbumRegex = regexp.MustCompile(
		`(?i)spotify\.com/album/([a-zA-Z0-9]+)`,
	)
	spotifyArtistRegex = regexp.MustCompile(
		`(?i)spotify\.com/artist/([a-zA-Z0-9]+)`,
	)
	spotifyLinkRegex = regexp.MustCompile(
		`(?i)^(https?:\/\/)?(open\.)?spotify\.com\/`,
	)
	nameCleanupRe = regexp.MustCompile(`[\(\[].*?[\)\]]`)
	spotifyCache  = utils.NewCache[string, []*state.Track](1 * time.Hour)
)

const PlatformSpotify state.PlatformName = "Spotify"

func init() {
	Register(95, &SpotifyPlatform{
		name: PlatformSpotify,
	})
}

func (s *SpotifyPlatform) Name() state.PlatformName {
	return s.name
}

func (s *SpotifyPlatform) CanGetTracks(query string) bool {
	return spotifyLinkRegex.MatchString(query)
}

func (s *SpotifyPlatform) GetTracks(
	query string,
	video bool,
) ([]*state.Track, error) {
	if config.SpotifyClientID == "" || config.SpotifyClientSecret == "" {
		return nil, errors.New(
			"spotify client credentials not configured. Set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET",
		)
	}

	// Check cache first
	cacheKey := "spotify:" + strings.ToLower(query)
	if cached, ok := spotifyCache.Get(cacheKey); ok {
		return updateVideoFlag(cached, video), nil
	}

	// Initialize client if needed
	if err := s.ensureClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize Spotify client: %w", err)
	}

	var tracks []*state.Track
	var err error

	ctx := context.Background()

	// Handle different Spotify URL types
	if matches := spotifyTrackRegex.FindStringSubmatch(query); len(
		matches,
	) > 1 {
		trackID := spotify.ID(matches[1])
		tracks, err = s.getTrack(ctx, trackID)
	} else if matches := spotifyPlaylistRegex.FindStringSubmatch(query); len(matches) > 1 {
		playlistID := spotify.ID(matches[1])
		tracks, err = s.getPlaylist(ctx, playlistID)
	} else if matches := spotifyAlbumRegex.FindStringSubmatch(query); len(matches) > 1 {
		albumID := spotify.ID(matches[1])
		tracks, err = s.getAlbum(ctx, albumID)
	} else if matches := spotifyArtistRegex.FindStringSubmatch(query); len(matches) > 1 {
		artistID := spotify.ID(matches[1])
		tracks, err = s.getArtistTopTracks(ctx, artistID)
	} else {
		return nil, errors.New("invalid Spotify URL format. Supported: tracks, playlists, albums, artists")
	}

	if err != nil {
		return nil, err
	}

	if len(tracks) > 0 {
		spotifyCache.Set(cacheKey, tracks)
	}

	return updateVideoFlag(tracks, video), nil
}

func (s *SpotifyPlatform) CanDownload(source state.PlatformName) bool {
	return source == s.name
}

func (s *SpotifyPlatform) Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	clean := cleanTitle(track.Title)
	trimmed := trimTitleLen(clean, 25, 40)

	var queries []string

	if clean != "" {
		queries = append(queries, clean)
	}

	if clean != "" && trimmed != "" && trimmed != clean {
		queries = append(queries, clean+" "+trimmed)
	}

	if trimTitleLen(track.Title, 25, 40) != "" {
		queries = append(queries, trimTitleLen(track.Title, 25, 40))
	}

	var ytTrack *state.Track

	for i, q := range queries {
		gologging.DebugF(
			"[Spotify→YouTube] Search attempt %d: %q",
			i+1,
			q,
		)

		ytTracks, err := yt.VideoSearch(q, true)
		if err != nil || len(ytTracks) == 0 {
			gologging.DebugF(
				"[Spotify→YouTube] No result for %q (err=%v)",
				q,
				err,
			)
			continue
		}

		ytTrack = ytTracks[0]
		ytTrack.Video = track.Video

		gologging.DebugF(
			"[Spotify→YouTube] Match found using %q → %s",
			q,
			ytTrack.URL,
		)
		break
	}

	if ytTrack == nil {
		return "", errors.New("failed to find track on YouTube")
	}

	for _, p := range GetOrderedPlatforms() {
		if p.CanDownload(PlatformYouTube) {
			path, err := p.Download(ctx, ytTrack, mystic)
			if err == nil {
				gologging.InfoF(
					"Downloaded Spotify track '%s' from YouTube: %s",
					track.Title,
					ytTrack.URL,
				)
				return path, nil
			}

			gologging.DebugF(
				"[Spotify→YouTube] Downloader %T failed: %v",
				p,
				err,
			)
		}
	}

	return "", errors.New("no YouTube downloader available")
}

func (*SpotifyPlatform) CanSearch() bool { return false }

func (*SpotifyPlatform) Search(
	string,
	bool,
) ([]*Track, error) {
	return nil, nil
}

// ensureClient initializes the Spotify client (once)
func (s *SpotifyPlatform) ensureClient() error {
	s.clientOnce.Do(func() {
		config := &clientcredentials.Config{
			ClientID:     config.SpotifyClientID,
			ClientSecret: config.SpotifyClientSecret,
			TokenURL:     spotifyauth.TokenURL,
		}

		token, err := config.Token(context.Background())
		if err != nil {
			s.initErr = fmt.Errorf("failed to get Spotify token: %w", err)
			return
		}

		httpClient := spotifyauth.New().Client(context.Background(), token)
		s.client = spotify.New(httpClient)

		gologging.Info("Spotify client initialized successfully")
	})

	return s.initErr
}

// getTrack fetches a single track by ID
func (s *SpotifyPlatform) getTrack(
	ctx context.Context,
	trackID spotify.ID,
) ([]*state.Track, error) {
	fullTrack, err := s.client.GetTrack(ctx, trackID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spotify track: %w", err)
	}

	track := s.convertSpotifyTrack(
		&fullTrack.SimpleTrack,
		fullTrack.Album.Images,
	)
	return []*state.Track{track}, nil
}

// getPlaylist fetches all tracks from a playlist
func (s *SpotifyPlatform) getPlaylist(
	ctx context.Context,
	playlistID spotify.ID,
) ([]*state.Track, error) {
	var tracks []*state.Track
	offset := 0
	limit := 100

	for {
		playlistPage, err := s.client.GetPlaylistItems(
			ctx,
			playlistID,
			spotify.Limit(limit),
			spotify.Offset(offset),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Spotify playlist: %w", err)
		}

		for _, item := range playlistPage.Items {
			if item.Track.Track == nil {
				continue
			}

			track := s.convertSpotifyTrack(
				&item.Track.Track.SimpleTrack,
				item.Track.Track.Album.Images,
			)
			tracks = append(tracks, track)
		}

		if playlistPage.Next == "" {
			break
		}

		offset += limit

	}

	return tracks, nil
}

// getAlbum fetches all tracks from an album
func (s *SpotifyPlatform) getAlbum(
	ctx context.Context,
	albumID spotify.ID,
) ([]*state.Track, error) {
	album, err := s.client.GetAlbum(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spotify album: %w", err)
	}

	var tracks []*state.Track

	for _, simpleTrack := range album.Tracks.Tracks {
		track := s.convertSpotifyTrack(&simpleTrack, album.Images)
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// getArtistTopTracks fetches an artist's top tracks
func (s *SpotifyPlatform) getArtistTopTracks(
	ctx context.Context,
	artistID spotify.ID,
) ([]*state.Track, error) {
	topTracks, err := s.client.GetArtistsTopTracks(ctx, artistID, "US")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist's top tracks: %w", err)
	}

	if len(topTracks) == 0 {
		return nil, errors.New("no tracks found for this artist")
	}

	var tracks []*state.Track

	for _, fullTrack := range topTracks {
		track := s.convertSpotifyTrack(
			&fullTrack.SimpleTrack,
			fullTrack.Album.Images,
		)
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *SpotifyPlatform) convertSpotifyTrack(
	simpleTrack *spotify.SimpleTrack,
	images []spotify.Image,
) *state.Track {
	var artists []string
	for _, artist := range simpleTrack.Artists {
		artists = append(artists, artist.Name)
	}
	artistStr := strings.Join(artists, ", ")
	_ = artistStr

	thumbnail := ""
	if len(images) > 0 {
		thumbnail = images[0].URL
	}

	title := simpleTrack.Name
	duration := int(simpleTrack.Duration) / 1000

	track := &state.Track{
		ID:       string(simpleTrack.ID),
		Title:    title,
		Duration: duration,
		Artwork:  thumbnail,
		URL:      simpleTrack.ExternalURLs["spotify"],
		Source:   PlatformSpotify,
	}

	return track
}

func updateVideoFlag(tracks []*state.Track, video bool) []*state.Track {
	if len(tracks) == 0 {
		return nil
	}

	result := make([]*state.Track, len(tracks))
	for i, t := range tracks {
		if t == nil {
			continue
		}
		clone := *t
		clone.Video = video
		result[i] = &clone
	}

	return result
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	title = nameCleanupRe.ReplaceAllString(title, "")

	title = strings.Join(strings.Fields(title), " ")
	return title
}

func trimTitleLen(title string, min, max int) string {
	title = strings.TrimSpace(title)

	runes := []rune(title)
	length := len(runes)

	if length <= max {
		return title
	}

	trimmed := runes[:max]

	// avoid cutting mid-word (space based)
	cut := max
	for i := max - 1; i >= min; i-- {
		if trimmed[i] == ' ' {
			cut = i
			break
		}
	}

	return strings.TrimSpace(string(trimmed[:cut]))
}

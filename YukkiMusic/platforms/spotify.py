#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import re

import spotipy
from async_lru import alru_cache

import config
from ..core.enum import SourceType
from ..core.track import Track
from .base import PlatformBase


class Spotify(PlatformBase):
    def __init__(self):
        self.regex = r"^(https:\/\/open.spotify.com\/)(.*)$"
        self.client_id = config.SPOTIFY_CLIENT_ID
        self.client_secret = config.SPOTIFY_CLIENT_SECRET
        if config.SPOTIFY_CLIENT_ID and config.SPOTIFY_CLIENT_SECRET:
            self.client_credentials_manager = spotipy.oauth2.SpotifyClientCredentials(
                self.client_id, self.client_secret
            )
            self.spotify = spotipy.Spotify(
                client_credentials_manager=self.client_credentials_manager
            )
        else:
            self.spotify = None

    async def valid(self, link: str):
        return bool(re.search(self.regex, link))

    @alru_cache(maxsize=None)
    async def track(self, link: str) -> Track:
        track = self.spotify.track(link)
        info = track["name"]
        for artist in track["artists"]:
            fetched = f' {artist["name"]}'
            if "Various Artists" not in fetched:
                info += fetched
        from .youtube import YouTube        
        t = await YouTube.track(info)
        t.link = link
        t.streamtype = SourceType.SPOTIFY
        return t

    @alru_cache(maxsize=None)
    async def playlist(self, url: str) -> tuple:
        playlist = self.spotify.playlist(url)
        playlist_id = playlist["id"]
        results = []
        for item in playlist["tracks"]["items"]:
            music_track = item["track"]
            info = music_track["name"]
            for artist in music_track["artists"]:
                fetched = f' {artist["name"]}'
                if "Various Artists" not in fetched:
                    info += fetched
            results.append(info)
            
        if len(results) > 0:
            from .youtube import YouTube
            t = await YouTube.track(results.pop(0))
            t.link = url
            t.streamtype = SourceType.SPOTIFY
            results.insert(0, t)
        return results, playlist_id

    @alru_cache(maxsize=None)
    async def album(self, url: str) -> tuple:
        album = self.spotify.album(url)
        album_id = album["id"]
        results = []
        for item in album["tracks"]["items"]:
            info = item["name"]
            for artist in item["artists"]:
                fetched = f' {artist["name"]}'
                if "Various Artists" not in fetched:
                    info += fetched
            results.append(info)
            
        if len(results) > 0:
            from .youtube import YouTube
            t = await YouTube.track(results.pop(0))
            t.link = url
            t.streamtype = SourceType.SPOTIFY
            results.insert(0, t)    
        return results, album_id

    @alru_cache(maxsize=None)
    async def artist(self, url: str) -> tuple:
        artist_info = self.spotify.artist(url)
        artist_id = artist_info["id"]
        results = []
        artist_top_tracks = self.spotify.artist_top_tracks(url)
        for item in artist_top_tracks["tracks"]:
            info = item["name"]
            for artist in item["artists"]:
                fetched = f' {artist["name"]}'
                if "Various Artists" not in fetched:
                    info += fetched
            info = await search(info)
            info.link = url
            info.streamtype = SourceType.SPOTIFY
            results.append(info)
            
        if len(results) > 0:
            from .youtube import YouTube
            t = await YouTube.track(results.pop(0))
            t.link = url
            t.streamtype = SourceType.SPOTIFY
            results.insert(0, t)
        return results, artist_id

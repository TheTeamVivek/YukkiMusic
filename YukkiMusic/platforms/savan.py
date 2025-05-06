#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from async_lru import alru_cache
from yt_dlp import YoutubeDL

from config import PLAYLIST_FETCH_LIMIT

from ..core.enum import SourceType
from ..core.request import Request
from ..core.track import Track
from .base import PlatformBase


class Saavn(PlatformBase):
    async def valid(self, link: str) -> bool:
        return "jiosaavn.com" in link

    async def track(self, url):
        is_song = lambda url: "song" in url and not any(
            x in url for x in ["/featured/", "/album/"]
        )
        is_playlist = lambda url: "/featured/" in url or "/album" in url
        handlers = {
            is_song: self.__track,
            is_playlist: self.playlist,
        }
        for condition, func in handlers.items():
            if condition(url):
                return await func(url)
        return None

    @alru_cache(maxsize=None)
    async def playlist(self, url, limit: int = PLAYLIST_FETCH_LIMIT):
        x = await self._playlist(url, limit)
        return x

    # @asyncify
    async def _playlist(self, url, limit: int = PLAYLIST_FETCH_LIMIT):
        ydl_opts = {
            "extract_flat": True,
            "force_generic_extractor": True,
            "quiet": True,
        }
        tracks = []
        with YoutubeDL(ydl_opts) as ydl:
            try:
                playlist_info = ydl.extract_info(url, download=False)
                for entry in playlist_info["entries"]:
                    if len(tracks) == limit:
                        break
                    duration_sec = entry.get("duration", 0)
                    track = Track(
                        title=entry["title"],
                        duration=duration_sec,
                        thumb=entry.get("thumbnail", None),
                        link=entry["webpage_url"],
                        video=False,
                        streamtype=SourceType.SAAVN,
                        is_live=False,
                    )
                    tracks.append(track)
            except Exception:
                pass
        return tracks

    @alru_cache(maxsize=None)
    # @asyncify
    async def __track(self, url):
        ydl_opts = {
            "format": "bestaudio/best",
            "geo_bypass": True,
            "nocheckcertificate": True,
            "quiet": True,
            "no_warnings": True,
        }

        with YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)
            return Track(
                title=info["title"],
                link=info["webpage_url"],
                duration=info.get("duration", 0),
                thumb=info.get("thumbnail", None),
                video=False,
                streamtype=SourceType.SAAVN,
                is_live=False,
            )


@alru_cache(maxsize=None)
async def search(self, query: str) -> Track:
    url = "https://saavn.dev/api/search/songs"
    result = await Request.get_json(url, params={"query": query, "limit": 1})
    if result.get("success"):
        info = result["data"]["topQuery"]["results"][0]
        return await self.track(info["url"])

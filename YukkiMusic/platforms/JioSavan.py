#
# Copyright (C) 2024-2025-2025-2025-2025-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os

import yt_dlp
from async_lru import alru_cache

from config import seconds_to_time
from YukkiMusic.utils.decorators import asyncify

from .base import PlatformBase
from ..core.youtube import Track

class Saavn(PlatformBase):
    async def valid(self, link: str) -> bool:
        return "jiosaavn.com" in link

    async def is_song(self, url: str) -> bool: #TODO remove this function
        return "song" in url and not "/featured/" in url and "/album/" not in url

    async def is_playlist(self, url: str) -> bool:
        return "/featured/" in url or "/album" in url #TODO Remove this function

    def clean_url(self, url: str) -> str: 
        if "#" in url:
            url = url.split("#")[0]
        return url

    @alru_cache(maxsize=None)
    @asyncify
    def playlist(self, url, limit):
        clean_url = self.clean_url(url)
        ydl_opts = {
            "extract_flat": True,
            "force_generic_extractor": True,
            "quiet": True,
        }
        song_info = []
        count = 0
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            try:
                playlist_info = ydl.extract_info(clean_url, download=False)
                for entry in playlist_info["entries"]:
                    if count == limit:
                        break
                    duration_sec = entry.get("duration", 0)
                    info = {
                        "title": entry["title"],
                        "duration_sec": duration_sec,
                        "duration_min": seconds_to_time(duration_sec),
                        "thumb": entry.get("thumbnail", ""),
                        "url": self.clean_url(entry["url"]),
                    }
                    song_info.append(info)
                    count += 1
            except Exception:
                pass
        return song_info

    @alru_cache(maxsize=None)
    @asyncify
    def track(self, url):
        url = self.clean_url(url)
        ydl_opts = {
            "format": "bestaudio/best",
            "geo_bypass": True,
            "nocheckcertificate": True,
            "quiet": True,
            "no_warnings": True,
        }

        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(clean_url, download=False)
            return Track(
                title = info["title"],
                link = self.clean_url(info["url"])
                duration_sec = info.get("duration", 0),
                thumb = info.get("thumbnail", None),
            )

    @asyncify
    def download(self, url):
        clean_url = self.clean_url(url)
        ydl_opts = {
            "format": "bestaudio/best",
            "outtmpl": "downloads/%(id)s.%(ext)s",
            "geo_bypass": True,
            "nocheckcertificate": True,
            "quiet": True,
            "no_warnings": True,
            "retries": 3,
            "nooverwrites": False,
            "continuedl": True,
        }

        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(clean_url, download=False)
            file_path = os.path.join("downloads", f"{info['id']}.{info['ext']}")

            if os.path.exists(file_path):
                return file_path, {
                    "title": info["title"],
                    "duration_sec": info.get("duration", 0),
                    "duration_min": seconds_to_time(info.get("duration", 0)),
                    "thumb": info.get("thumbnail", None),
                    "url": self.clean_url(info["url"]),
                    "filepath": file_path,
                }

            ydl.download([clean_url])
            return file_path, {
                "title": info["title"],
                "duration_sec": info.get("duration", 0),
                "duration_min": seconds_to_time(info.get("duration", 0)),
                "thumb": info.get("thumbnail", None),
                "url": self.clean_url(info["url"]),
                "filepath": file_path,
            }

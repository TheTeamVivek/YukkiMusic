#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os

import aiohttp
import yt_dlp

from config import seconds_to_time
from YukkiMusic.utils.decorators import asyncify


class Saavn:

    @staticmethod
    async def valid(url: str) -> bool:
        return "jiosaavn.com" in url

    @staticmethod
    async def is_song(url: str) -> bool:
        return "song" in url and not "/featured/" in url and "/album/" not in url

    @staticmethod
    async def is_playlist(url: str) -> bool:
        return "/featured/" in url or "/album" in url

    def clean_url(self, url: str) -> str:
        if "#" in url:
            url = url.split("#")[0]
        return url

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
                        "url": self.clean_url(entry["webpage_url"]),
                    }
                    song_info.append(info)
                    count += 1
            except Exception:
                pass
        return song_info

    async def info(self, url):
        url = self.clean_url(url)

        async with aiohttp.ClientSession() as session:
            if "jiosaavn.com" in url:
                api_url = "https://saavn.dev/api/songs"
                params = {"link": url, "limit": 1}
            else:
                api_url = "https://saavn.dev/api/search/songs"
                params = {"query": url, "limit": 1}

            async with session.get(api_url, params=params) as response:
                data = await response.json()

                if "jiosaavn.com" in url:
                    info = data["data"][0]  # For Saavn URLs
                else:
                    info = data["data"]["results"][0]  # For search queries

                return {
                    "title": info["name"],
                    "duration_sec": info.get("duration", 0),
                    "duration_min": seconds_to_time(info.get("duration", 0)),
                    "thumb": info["image"][-1]["url"],
                    "url": self.clean_url(info["url"]),
                    "_download_url": info["downloadUrl"][-1]["url"],
                    "_id": info["id"],
                }

    async def download(self, url):
        details = await self.info(url)
        file_path = os.path.join("downloads", f"Saavn_{details['_id']}.mp3")

        if not os.path.exists(file_path):
            async with aiohttp.ClientSession() as session:
                async with session.get(details["_download_url"]) as resp:
                    if resp.status == 200:
                        with open(file_path, "wb") as f:
                            while chunk := await resp.content.read(1024):
                                f.write(chunk)
                        print(f"Downloaded: {file_path}")
                    else:
                        raise ValueError(
                            f"Failed to download {details['_download_url']}. HTTP Status: {resp.status}"
                        )

        details["filepath"] = file_path
        return file_path, details

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os
from io import BytesIO

import aiofiles
import aiohttp
import yt_dlp
from PIL import Image

from config import seconds_to_time
from yukkimusic.utils.decorators import asyncify


class Saavn:
    @staticmethod
    async def valid(url: str) -> bool:
        return "jiosaavn.com" in url

    @staticmethod
    async def is_song(url: str) -> bool:
        return "song" in url and "/featured/" not in url and "/album/" not in url

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

                thumb_url = info["image"][-1]["url"]
                thumb_path = await self._resize_thumb(thumb_url, info["id"])

                return {
                    "title": info["name"],
                    "duration_sec": info.get("duration", 0),
                    "duration_min": seconds_to_time(info.get("duration", 0)),
                    "thumb": thumb_path,
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
                        async with aiofiles.open(file_path, "wb") as f:
                            while chunk := await resp.content.read(1024):
                                await f.write(chunk)
                        print(f"Downloaded: {file_path}")
                    else:
                        raise ValueError(
                            f"Failed to download {details['_download_url']}. HTTP Status: {resp.status}"
                        )

        details["filepath"] = file_path
        return file_path, details

    async def _resize_thumb(self, thumb_url, _id, size=(1280, 720)):
        thumb_path = os.path.join("cache", f"Thumb_{_id}.jpg")

        if os.path.exists(thumb_path):
            return thumb_path

        async with aiohttp.ClientSession() as session:
            async with session.get(thumb_url) as response:
                img_data = await response.read()

        img = Image.open(BytesIO(img_data))
        scale_factor = size[1] / img.height
        new_width = int(img.width * scale_factor)
        new_height = size[1]

        resized_img = img.resize((new_width, new_height), Image.LANCZOS)
        new_img = Image.new("RGB", size, (0, 0, 0))
        new_img.paste(resized_img, ((size[0] - new_width) // 2, 0))

        new_img.save(thumb_path, format="JPEG")
        return thumb_path

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

from async_lru import alru_cache
from bs4 import BeautifulSoup

from ..core.request import Request
from ..core.youtube import Track, YouTube
from .base import PlatformBase


class Apple(PlatformBase):
    def __init__(self):
        self.regex = r"^(https:\/\/music.apple.com\/)(.*)$"
        self.base = "https://music.apple.com/in/playlist/"

    async def valid(self, link: str):
        return bool(re.search(self.regex, link))

    @alru_cache(maxsize=None)
    async def track(self, url: str) -> Track | bool:
        html = await Request.get_text(url)
        soup = BeautifulSoup(html, "html.parser")
        song_name = None
        for tag in soup.find_all("meta"):
            if tag.get("property", None) == "og:title":
                song_name = tag.get("content", None)
        if song_name is None:
            return False
        return await YouTube.search(song_name)

    @alru_cache(maxsize=None)
    async def playlist(self, url, playid: bool | str = None) -> list[Track]:
        if playid:
            url = self.base + url
        playlist_id = url.split("playlist/")[1]

        html = await Request.get_text(url)
        soup = BeautifulSoup(html, "html.parser")
        applelinks = soup.find_all("meta", attrs={"property": "music:song"})
        results = []
        for item in applelinks:
            try:
                xx = (((item["content"]).split("album/")[1]).split("/")[0]).replace(
                    "-", " "
                )
            except Exception:
                xx = ((item["content"]).split("album/")[1]).split("/")[0]
            results.append(await YouTube.search(xx))
        return results, playlist_id

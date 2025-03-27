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

import config

from ..core.enum import SourceType
from ..core.request import Request
from ..core.track import Track
from .base import PlatformBase


class Apple(PlatformBase):
    def __init__(self):
        self.regex = re.compile(r"^(https:\/\/music.apple.com\/)(.*)$")
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
        from .youtube import YouTube

        t = await YouTube.track(song_name)
        t.streamtype = SourceType.APPLE
        t.link = url
        return t

    @alru_cache(maxsize=None)
    async def playlist(
        self, url: str, limit: int = config.PLAYLIST_FETCH_LIMIT
    ) -> list[Track]:

        playlist_id = url.split("playlist/")[1]

        html = await Request.get_text(url)
        soup = BeautifulSoup(html, "html.parser")
        applelinks = soup.find_all("meta", attrs={"property": "music:song"})
        results = []
        from .youtube import YouTube

        for item in applelinks:
            if len(results) == limit:
                break
            try:
                xx = (((item["content"]).split("album/")[1]).split("/")[0]).replace(
                    "-", " "
                )
            except Exception:
                ((item["content"]).split("album/")[1]).split("/")[0]
            results.append(t)

        if len(results) > 0:
            t = await YouTube.track(results.pop(0))
            t.streamtype = SourceType.APPLE
            t.link = url
            results.insert(0, t)
        return results, playlist_id

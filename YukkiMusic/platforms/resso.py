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
from ..core.youtube import Track, search, SourceType
from .base import PlatformBase


class Resso(PlatformBase):
    def __init__(self):
        self.regex = r"^(https:\/\/m.resso.com\/)(.*)$"
        self.base = "https://m.resso.com/"

    async def valid(self, link: str):
        return bool(re.search(self.regex, link))

    @alru_cache(maxsize=None)
    async def track(self, url, playid: bool | str | None = None) -> Track:
        if playid:
            url = self.base + url
        html = await Request.get_text(url)
        soup = BeautifulSoup(html, "html.parser")
        for tag in soup.find_all("meta"):
            if tag.get("property", None) == "og:title":
                title = tag.get("content", None)
            if tag.get("property", None) == "og:description":
                des = tag.get("content", None)
                try:
                    des = des.split("·")[0]
                except Exception:
                    pass
        if des == "":
            return
        track = await search(title)
        track.link = url
        track.streamtype = SourceType.RESSO
        return track

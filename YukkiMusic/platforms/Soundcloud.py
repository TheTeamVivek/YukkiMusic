#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, <https://github.com/TheTeamVivek>.
#
# This file is part of <https://github.com/TheTeamVivek/YukkiMusic> project,
# and is released under the MIT License.
# Please see <https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE>
#
# All rights reserved.
#

from yt_dlp import YoutubeDL

from YukkiMusic.utils.decorators import asyncify
from ..core.youtube import Track
from .base import PlatformBase


class SoundCloud(PlatformBase):
    async def valid(self, link: str) -> bool:
        return "soundcloud" in link.lower()

    @asyncify
    def track(self, url: str) -> Track | bool:
        options = {
            "format": "bestaudio/best",
            "retries": 3,
            "quiet": True,
            "noplaylist": True,
            "extract_flat": False,
        }

        with YoutubeDL(options) as ydl:
            try:
                info = ydl.extract_info(url, download=False)
            except Exception:
                return False

            return Track(
                title=info["title"],
                duration_sec=["duration"],
                link=url,
                thumb=info["thumbnails"][0]["url"] if info.get("thumbnails") else None,
            )

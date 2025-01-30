from dataclasses import dataclass, field
import asyncio

from config import cookies
from yt_dlp import YoutubeDL
from youtubesearchpython.__future__ import VideosSearch
from YukkiMusic.utils.formatters import seconds_to_min, time_to_seconds
from YukkiMusic.decorators.asyncify import asyncify

@dataclass
class Track:
    title: str
    vidid: str
    link: str
    thumb: str
    duration_min: int | None = field(default=None)
    duration_sec: int | None = field(default=None)

    def __post_init__(self):
        if self.duration_min is not None and self.duration_sec is None:
            self.duration_sec = time_to_seconds(self.duration_min)
        elif self.duration_sec is not None and self.duration_min is None:
            self.duration_min = seconds_to_min(self.duration_sec)

class YouTube:
    def __init__(self, query):
        self.query = query

    async def search(self) -> Track:
        try:
            results = VideosSearch(self.query, limit=1)
            for result in (await results.next())["result"]:
                return Track(
                    title=result["title"],
                    vidid=result["id"],
                    link=result["link"],
                    duration_min=int(result["duration"])
                    if result["duration"]
                    else None,
                    duration_sec=None,
                    thumb=result["thumbnails"][0]["url"].split("?")[0],
                )
        except Exception:
            return await self._search_yt_dlp()

    @asyncify
    def _search_yt_dlp(self) -> Track:
        options = {
            "format": "best",
            "noplaylist": True,
            "quiet": True,
            "extract_flat": "in_playlist",
            "cookiefile": cookies(),
        }

        with YoutubeDL(options) as ydl:
            info_dict = ydl.extract_info(f"ytsearch: {self.query}", download=False)
            details = info_dict.get("entries", [None])[0]
            if not details:
                raise ValueError("No results found.")

            return Track(
                title=details["title"],
                vidid=details["id"],
                link=details["url"],
                duration_min=None,
                duration_sec=details["duration"] if details["duration"] != 0 else None,
                thumb=details["thumbnails"][0]["url"],
            )

from dataclasses import dataclass, field

from async_lru import alru_cache
from youtubesearchpython.__future__ import VideosSearch
from yt_dlp import YoutubeDL

from config import cookies
from YukkiMusic.decorators.asyncify import asyncify
from YukkiMusic.utils.formatters import time_to_seconds

from .enum import SongType


@dataclass
class Track:
    title: str
    link: str
    thumb: str
    duration: int  # duration in Seconds
    download_url: str | None = field(default=None)
    file_path: str | None = field(default=None)

    @property
    def is_exists(self):
        return bool(os.path.exists(self.file_path))

    @asyncify
    def download(
        self,
        type: SongType = SongType.AUDIO,
        options: dict | None = None,
    ):
        ytdl_opts = {
            "format": (
                "bestaudio/best"
                if type == SongType.AUDIO
                else "(bestvideo[height<=?720][width<=?1280][ext=mp4])+(bestaudio[ext=m4a])"
            ),
            "continuedl": True,
            "outtmpl": "downloads/%(id)s.%(ext)s",
            "geo_bypass": True,
            "noplaylist": True,
            "nocheckcertificate": True,
            "quiet": True,
            "retries": 3,
            "no_warnings": True,
            #  "cookiefile": cookies(),
        }
        if options is not None:
            if isinstance(options, dict):
                ytdl_opts.update(options)
            else:
                raise Exception(
                    f"Expected 'options' to be a dict but got {type(ytdl_opts).__name__}"
                )
        url = self.download_url if self.download_url else self.link
        with YoutubeDL(ytdl_opts) as ydl:
            info = ydl.extract_info(url, False)
            file_path = os.path.join("downloads", f"{info['id']}.{info['ext']}")
            if os.path.exists(file_path):
                self.file_path = file_path
                return file_path
            ydl.download([url])
            self.file_path = file_path
            return file_path


@alru_cache(maxsize=None)
async def search(query):
    try:
        results = VideosSearch(query, limit=1)
        for result in (await results.next())["result"]:
            return {
                "title": result["title"],
                "link": result["link"],
                "download_url": result["link"],
                "duration": (
                    time_to_seconds(result["duration"]) if result["duration"] else 0
                ),
                "thumb": result["thumbnails"][0]["url"].split("?")[0],
            }
    except Exception:
        return await search_from_ytdlp(query)


@alru_cache(maxsize=None)
@asyncify
def search_from_ytdlp(query):
    options = {
        "format": "best",
        "noplaylist": True,
        "quiet": True,
        "extract_flat": "in_playlist",
        "cookiefile": cookies(),
    }

    with YoutubeDL(options) as ydl:
        info_dict = ydl.extract_info(f"ytsearch: {query}", download=False)
        details = info_dict.get("entries", [None])[0]
        if not details:
            raise ValueError("No results found.")

        return Track(
            title=details["title"],
            link=(details["url"].split(&)[0] if "&" in details["url"] else details["url"]),
            download_url= details["url"],
            duration= details["duration"],
            thumb= details["thumbnails"][0]["url"],
        }

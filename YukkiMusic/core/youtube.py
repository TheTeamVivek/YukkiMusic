import asyncio
import os
import re
from dataclasses import dataclass

from async_lru import alru_cache
from youtubesearchpython.__future__ import VideosSearch
from yt_dlp import YoutubeDL

from config import YTDOWNLOADER, cookies
from YukkiMusic.decorators.asyncify import asyncify
from YukkiMusic.utils.database import is_on_off
from YukkiMusic.utils.formatters import time_to_seconds


from .enum import SourceType
@dataclass
class Track:
    title: str
    link: str
    thumb: str
    duration: int  # duration in seconds
    streamtype: SourceType
    by: str | None = None  # None but required
    vidid: str | None = None
    download_url: str | None = None
    file_path: str | None = None
    streamable_url: str | None = None

    def __post_init__(self):
        if self.is_youtube:
            pattern = r"(?:v=|\/)([0-9A-Za-z_-]{11})"
            url = self.download_url if self.download_url else self.link
            match = re.search(pattern, url)
            self.vidid = match.group(1) if match else None
        else:
            self.vidid = self.streamtype.value

    @property
    def is_exists(self):
        return bool(self.file_path and os.path.exists(self.file_path))

    @property
    def is_youtube(self) -> bool:
        url = self.download_url if self.download_url else self.link
        return "youtube.com" in url or "youtu.be" in url

    async def download(self, audio: bool = True, options: dict | None = None):
        url = self.download_url if self.download_url else self.link

        if await is_on_off(YTDOWNLOADER):
            ytdl_opts = {
                "format": (
                    "bestaudio/best"
                    if audio
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
            }

            if self.is_youtube:
                ytdl_opts["cookiefile"] = "cookies/cookies.txt"

            if options:
                if isinstance(options, dict):
                    ytdl_opts.update(options)
                else:
                    raise Exception(
                        f"Expected 'options' to be a dict but got {type(options).__name__}"
                    )

            @asyncify
            def _download():
                with YoutubeDL(ytdl_opts) as ydl:
                    info = ydl.extract_info(url, False)
                    self.file_path = os.path.join(
                        "downloads", f"{info['id']}.{info['ext']}"
                    )

                    if not os.path.exists(self.file_path):
                        ydl.download([url])

                    return self.file_path

            return await _download()

        else:
            format_code = "bestaudio/best" if audio else "b"  # Keep "b" not "best"
            command = f'yt-dlp -g -f "{format_code}" {"--cookies cookies/cookies.txt" if self.is_youtube else ""} "{url}"'

            process = await asyncio.create_subprocess_shell(
                command,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )
            stdout, stderr = await process.communicate()

            if stdout:
                self.streamable_url = stdout.decode().strip()
                return self.streamable_url
            else:
                raise Exception(
                    f"Failed to get streamable URL: {stderr.decode().strip()}"
                )

    async def __call__(self, audio: bool = True):
        return self.file_path or self.streamable_url or await self.download(audio)


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
            link=(
                details["webpage_url"].split("&")[0]
                if "&" in details["webpage_url"]
                else details["webpage_url"]
            ),
            download_url=details["webpage_url"],
            duration=details["duration"],
            thumb=details["thumbnails"][0]["url"],
        )

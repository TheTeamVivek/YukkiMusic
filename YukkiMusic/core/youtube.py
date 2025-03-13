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
    video: bool  # The song is audio or video

    download_url: str | None = (
        None  # If provided directly used to download instead self.link
    )
    is_live: bool | None = None
    vidid: str | None = None
    file_path: str | None = None

    def __post_init__(self):
        self.download_url = self.download_url if self.download_url else self.link
        if self.is_youtube and self.vidid is None:
            pattern = r"(?:v=|\/)([0-9A-Za-z_-]{11})"
            p_match = re.search(pattern, self.download_url)
            self.vidid = p_match.group(1)
        else:
            self.vidid = ""

        self.title = self.title.title() if self.title is not None else None
        if (
            not self.duration and self.is_live is None
        ):  # WHEN is_live is not None it means the track is live or not live means no need to check it
            if self.streamtype in [
                SourceType.APPLE,
                SourceType.RESSO,
                SourceType.SPOTIFY,
                SourceType.YOUTUBE,
            ]:
                self.is_live = True

    async def is_exists(self):
        exists = False

        if self.file_path:
            if await is_on_off(YTDOWNLOADER):
                exists = os.path.exists(self.file_path)
            else:
                exists = (
                    len(self.file_path) > 30
                )  # FOR m3u8 URLS for m3u8 download mode

        return exists

    @property
    def is_youtube(self) -> bool:
        return "youtube.com" in self.download_url or "youtu.be" in self.download_url

    @property
    def is_m3u8(self) -> bool:
        return self.streamtype == SourceType.M3U8

    async def download(self):
        if (
            self.file_path is not None and await self.is_exists()
        ):  # THIS CONDITION FOR TELEGRAM FILES BECAUSE THESE FILES ARE ALREADY DOWNLOADED
            return self.file_path

        if await is_on_off(YTDOWNLOADER) and not (self.is_live or self.is_m3u8):
            ytdl_opts = {
                "format": (
                    "(bestvideo[height<=?720][width<=?1280][ext=mp4])+(bestaudio[ext=m4a])"
                    if self.video
                    else "bestaudio/best"
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
                ytdl_opts["cookiefile"] = cookies()

            @asyncify
            def _download():
                with YoutubeDL(ytdl_opts) as ydl:
                    info = ydl.extract_info(self.download_url, False)
                    self.file_path = os.path.join(
                        "downloads", f"{info['id']}.{info['ext']}"
                    )

                    if not os.path.exists(self.file_path):
                        ydl.download([self.download_url])

                    return self.file_path

            return await _download()

        else:
            if self.is_m3u8:
                return self.link or self.download_url

            format_code = "b" if self.video else "bestaudio/best"  # Keep "b" not "best"
            command = f'yt-dlp -g -f "{format_code}" {"--cookies " + cookies() if self.is_youtube else ""} "{self.download_url}"'
            process = await asyncio.create_subprocess_shell(
                command,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )
            stdout, stderr = await process.communicate()

            if stdout:
                self.file_path = stdout.decode("utf-8").strip().split()[0]
                return self.file_path
            else:
                raise Exception(
                    f"Failed to get file path: {stderr.decode('utf-8').strip()}"
                )

    async def __call__(self):
        return self.file_path or await self.download()


@alru_cache(maxsize=None)
async def search(query: str, video: bool = False):
    try:
        results = VideosSearch(query, limit=1)
        for result in (await results.next())["result"]:
            duration = result.get("duration")
            return Track(
                title=result["title"],
                link=result["link"],
                download_url=result["link"],
                duration=(
                    time_to_seconds(duration) if str(duration) != "None" else 0
                ),  # TODO: CHECK THAT THE YOUTBE SEARCH PYTHON RETUNS DURATION IS None or "None"
                thumb=result["thumbnails"][0]["url"].split("?")[0],
                streamtype=SourceType.YOUTUBE,
                video=video,
            )
    except Exception:
        return await search_from_ytdlp(query)


@alru_cache(maxsize=None)
@asyncify
def search_from_ytdlp(query: str, video: bool = False):
    options = {
        "format": "best",
        "noplaylist": True,
        "quiet": True,
        "extract_flat": "in_playlist",
        "cookiefile": cookies(),
    }

    with YoutubeDL(options) as ydl:
        info_dict = ydl.extract_info(
            f"ytsearch: {query}", download=False
        )  # TODO: THIS CAN RETURN SEARCH RESULT OF A CHANNEL FIX IT
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
            streamtype=SourceType.YOUTUBE,  # KEEP HERE YOUTUBE LATER WE CAN CHANGE IT TO CORRECT SourceType
            video=video,
        )

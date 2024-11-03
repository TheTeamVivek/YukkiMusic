import asyncio
import os

import yt_dlp

from config import seconds_to_time


class Saavn:
    def __init__(self):
        pass

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

    async def playlist(self, url, limit):
        loop = asyncio.get_running_loop()
        clean_url = self.clean_url(url)

        def play_list():
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
                            "url": self.clean_url(entry["url"]),
                        }
                        song_info.append(info)
                        count += 1
                except Exception:
                    pass
            return song_info

        return await loop.run_in_executor(None, play_list)

    async def info(self, url):
        loop = asyncio.get_running_loop()
        clean_url = self.clean_url(url)

        def get_info():
            ydl_opts = {
                "format": "bestaudio/best",
                "geo_bypass": True,
                "nocheckcertificate": True,
                "quiet": True,
                "no_warnings": True,
            }

            with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                info = ydl.extract_info(clean_url, download=False)
                return {
                    "title": info["title"],
                    "duration_sec": info.get("duration", 0),
                    "duration_min": seconds_to_time(info.get("duration", 0)),
                    "thumb": info.get("thumbnail", None),
                    "url": self.clean_url(info["url"]),
                }

        return await loop.run_in_executor(None, get_info)

    async def download(self, url):
        loop = asyncio.get_running_loop()
        clean_url = self.clean_url(url)

        def down_load():
            ydl_opts = {
                "format": "bestaudio/best",
                "outtmpl": "downloads/%(id)s.%(ext)s",
                "geo_bypass": True,
                "nocheckcertificate": True,
                "quiet": True,
                "no_warnings": True,
                "retries": 3,
                "nooverwrites": False,
                "continuedl": True,
            }

            with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                info = ydl.extract_info(clean_url, download=False)
                file_path = os.path.join("downloads", f"{info['id']}.{info['ext']}")

                if os.path.exists(file_path):
                    return file_path, {
                        "title": info["title"],
                        "duration_sec": info.get("duration", 0),
                        "duration_min": seconds_to_time(info.get("duration", 0)),
                        "thumb": info.get("thumbnail", None),
                        "url": self.clean_url(info["url"]),
                        "filepath": file_path,
                    }

                ydl.download([clean_url])
                return file_path, {
                    "title": info["title"],
                    "duration_sec": info.get("duration", 0),
                    "duration_min": seconds_to_time(info.get("duration", 0)),
                    "thumb": info.get("thumbnail", None),
                    "url": self.clean_url(info["url"]),
                    "filepath": file_path,
                }

        return await loop.run_in_executor(None, down_load)

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import subprocess
import time


class RtmpStream:
    _streams = {}

    def __new__(cls, chat_id, *args, **kwargs):
        if chat_id in cls._streams:
            return cls._streams[chat_id]
        instance = super().__new__(cls)
        cls._streams[chat_id] = instance
        return instance

    def __init__(
        self,
        chat_id: int,
        rtmp_url: str = None,
        file_path: str = None,
        video: bool = False,
        piped_image: str | None = None,
    ):
        if hasattr(self, "_initialized"):
            return
        self._initialized = True
        self.chat_id = chat_id
        self.rtmp_url = rtmp_url
        self.file_path = file_path
        self.video = video
        self.piped_image = piped_image
        self._task = None
        self._process = None
        self._start_time = None
        self._total_paused_duration = 0
        self._last_pause_time = None
        self._mute_start = None
        self._muted_played = None

    async def _play(self, offset=0):
        cmd = ["ffmpeg"]
        if offset > 0:
            cmd += ["-ss", str(offset)]
        if self.video:
            cmd += ["-i", self.file_path]
        else:
            cmd += [
                "-loop",
                "1",
                "-framerate",
                "2",
                "-i",
                self.piped_image or "black.jpg",
                "-i",
                self.file_path,
                "-shortest",
            ]
        cmd += [
            "-c:v",
            "libx264",
            "-preset",
            "superfast",
            "-b:v",
            "2000k",
            "-maxrate",
            "2000k",
            "-bufsize",
            "4000k",
            "-pix_fmt",
            "yuv420p",
            "-g",
            "30",
            "-threads",
            "0",
            "-c:a",
            "aac",
            "-b:a",
            "96k",
            "-ac",
            "2",
            "-ar",
            "44100",
            "-f",
            "flv",
            self.rtmp_url,
        ]
        self._start_time = time.time() - offset
        self._last_pause_time = None
        self._total_paused_duration = 0
        self._task = asyncio.create_task(self._run_ffmpeg(cmd))

    async def _run_ffmpeg(self, cmd):
        self._process = await asyncio.create_subprocess_exec(
            *cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        await self._process.wait()
        self._task = None

    @property
    def played_seconds(self) -> int:
        if not self._start_time:
            return 0
        paused = time.time() - self._last_pause_time if self._last_pause_time else 0
        return int(
            time.time() - self._start_time - self._total_paused_duration - paused
        )

    async def destroy(self):
        await self.stop()
        RtmpStream._streams.pop(self.chat_id, None)

    async def stop(self):
        if self._process:
            self._process.terminate()
            await self._process.wait()
        self._task = None
        self._process = None

    async def pause(self):
        self._last_pause_time = time.time()
        await self.stop()

    async def resume(self):
        if self._last_pause_time is None:
            return
        self._total_paused_duration += time.time() - self._last_pause_time
        await self._play(offset=self.played_seconds)

    async def mute(self):
        self._mute_start = time.time()
        self._muted_played = self.played_seconds
        await self.stop()

    async def unmute(self):
        if self._mute_start is None:
            return
        resume_at = self._muted_played + int(time.time() - self._mute_start)
        self._mute_start = None
        self._muted_played = None
        await self._play(offset=resume_at)

    async def seek(self, seconds: int):
        await self._play(offset=self.played_seconds + seconds)

    async def seekback(self, seconds: int):
        await self._play(offset=max(0, self.played_seconds - seconds))

    async def replay(self):
        await self._play(offset=0)

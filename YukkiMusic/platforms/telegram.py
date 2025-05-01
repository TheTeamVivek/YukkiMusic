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
import os
import time
from datetime import datetime, timedelta

import aiohttp
from telethon.tl import types

from config import lyrical
from YukkiMusic.utils.inline import downlod_markup

from ..utils.formatters import convert_bytes, get_readable_time

downloader = {}


class Telegram:
    def __init__(self):
        self.sleep = 5

    async def get_url_from_message(self, event) -> str | None:
        messages = [event.message]
        if event.is_reply:
            messages.append(await event.get_reply_message())
        text = ""
        offset = None
        length = None
        for message in messages:
            if offset:
                break
            if message.entities:
                for entity in message.entities:
                    if isinstance(entity, types.MessageEntityUrl):
                        text = message.text
                        offset, length = entity.offset, entity.length
                        break
                    elif isinstance(entity, types.MessageEntityTextUrl):
                        return entity.url
        if offset is None:
            return None
        return text[offset : offset + length]

    async def __get_filepath(self, file, video: bool = False):
        ext = file.name.split(".")[-1] if file.name else ("mp4" if video else "ogg")
        file_name = f"{file.id}.{ext}"
        return os.path.join(os.path.realpath("downloads"), file_name)

    async def is_streamable_url(self, url: str) -> bool:
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(url, timeout=5) as response:
                    if response.status == 200:
                        content_type = response.headers.get("Content-Type", "")
                        if (
                            "application/vnd.apple.mpegurl" in content_type
                            or "application/x-mpegURL" in content_type
                        ):
                            return True
                        if any(
                            keyword in content_type
                            for keyword in [
                                "audio",
                                "video",
                                "mp4",
                                "mpegurl",
                                "m3u8",
                                "mpeg",
                            ]
                        ):
                            return True
                        if url.endswith((".m3u8", ".index", ".mp4", ".mpeg", ".mpd")):
                            return True
        except aiohttp.ClientError:
            pass
        return False

    async def download(self, _, message, mystic, video=False):
        fname = await self.__get_filepath(message.file, video=video)
        left_time = {}
        speed_counter = {}
        if os.path.exists(fname):
            return True

        async def down_load():
            async def progress(current, total):
                if current == total:
                    return
                current_time = time.time()
                start_time = speed_counter.get(message.id)
                check_time = current_time - start_time
                upl = downlod_markup(_)
                if datetime.now() > left_time.get(message.id):
                    percentage = current * 100 / total
                    percentage = str(round(percentage, 2))
                    speed = current / check_time
                    eta = int((total - current) / speed)
                    downloader[message.id] = eta
                    eta = get_readable_time(eta)
                    if not eta:
                        eta = "0 sec"
                    total_size = convert_bytes(total)
                    completed_size = convert_bytes(current)
                    speed = convert_bytes(speed)
                    text = _["tg_3"].format(
                        tbot.mention,
                        total_size,
                        completed_size,
                        percentage[:5],
                        speed,
                        eta,
                    )
                    try:
                        await mystic.edit(text, buttons=upl)
                    except Exception:
                        pass
                    left_time[message.id] = datetime.now() + timedelta(
                        seconds=self.sleep
                    )

            speed_counter[message.id] = time.time()
            left_time[message.id] = datetime.now()

            try:
                await tbot.download_media(
                    message,
                    file=fname,
                    progress_callback=progress,
                )
                await mystic.edit(_["tg_4"])
                downloader.pop(message.id, None)
            except Exception:
                await mystic.edit(_["tg_2"])

        if len(downloader) > 10:
            timers = list(downloader.values())
            try:
                low = min(timers)
                eta = get_readable_time(low)
            except Exception:
                eta = "Unknown"
            await mystic.edit(_["tg_1"].format(eta))
            return False

        task = asyncio.create_task(down_load(), name=f"download_{event.chat_id}")
        lyrical[mystic.id] = task
        await task
        downloaded = downloader.get(message.id)
        if downloaded:
            downloader.pop(message.id)
            return False
        verify = lyrical.get(mystic.id)
        if not verify:
            return False
        lyrical.pop(mystic.id)
        return True

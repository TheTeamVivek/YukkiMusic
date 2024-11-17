#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import time
from datetime import datetime, timedelta

from pyrogram.types import InlineKeyboardMarkup

from strings import get_string
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import db
from YukkiMusic.utils.database import (
    get_active_chats,
    get_assistant,
    get_lang,
    is_music_playing,
    set_loop,
)
from YukkiMusic.utils.formatters import seconds_to_min
from YukkiMusic.utils.inline import stream_markup_timer, telegram_markup_timer

from ..admins.callback import wrong
from .autoleave import autoend

checker = {}
muted = {}


async def timer():
    while not await asyncio.sleep(1):
        active_chats = await get_active_chats()
        for chat_id in active_chats:
            if not await is_music_playing(chat_id):
                continue
            playing = db.get(chat_id)
            if not playing:
                continue
            file_path = playing[0]["file"]
            if "index_" in file_path or "live_" in file_path:
                continue
            duration = int(playing[0]["seconds"])
            if duration == 0:
                continue
            db[chat_id][0]["played"] += 1


async def leave_if_muted():
    while True:
        await asyncio.sleep(2)
        for chat_id, details in list(muted.items()):
            if time.time() - details["timestamp"] >= 60:
                _ = details["_"]
                try:
                    userbot = await get_assistant(chat_id)
                    members = []
                    try:
                        async for member in userbot.get_call_members(
                            chat_id
                        ):
                            if member is None:
                                continue
                            members.append(member)
                    except ValueError:
                        try:
                            await Yukki.stop_stream(chat_id)
                        except Exception:
                            pass
                        continue

                    m = next(
                        (m for m in members if m.chat.id == userbot.id), None
                    )
                    if m is None:
                        continue
                    is_muted = bool(m.is_muted and not m.can_self_unmute)

                    if is_muted:
                        await Yukki.stop_stream(chat_id)
                        await set_loop(chat_id, 0)

                    del muted[chat_id]
                except Exception:
                    del muted[chat_id]


async def markup_timer():
    while True:
        await asyncio.sleep(2)
        active_chats = await get_active_chats()
        for chat_id in active_chats:
            if not await is_music_playing(chat_id):
                continue

            playing = db.get(chat_id)
            if not playing:
                continue

            duration_seconds = int(playing[0]["seconds"])

            try:
                language = await get_lang(chat_id)
                _ = get_string(language)
            except Exception:
                _ = get_string("en")

            is_muted = False
            try:
                userbot = await get_assistant(chat_id)
                members = []
                try:
                    async for member in userbot.get_call_members(chat_id):
                        if member is None:
                            continue
                        members.append(member)
                except ValueError:
                    try:
                        await Yukki.stop_stream(chat_id)
                    except Exception:
                        pass
                    continue

                if not members:
                    await Yukki.stop_stream(chat_id)
                    await set_loop(chat_id, 0)
                    continue

                if len(members) <= 1 and chat_id not in autoend:
                    autoend[chat_id] = datetime.now() + timedelta(seconds=30)

                m = next((m for m in members if m.chat.id == userbot.id), None)
                if m is None:
                    continue

                is_muted = bool(m.is_muted and not m.can_self_unmute)
                if is_muted:

                    if chat_id not in muted:
                        muted[chat_id] = {
                            "timestamp": time.time(),
                            "_": _,
                        }

            except Exception:
                pass

            if duration_seconds == 0:
                continue

            try:
                mystic = playing[0]["mystic"]
                markup = playing[0]["markup"]
            except Exception:
                continue

            try:
                check = wrong[chat_id][mystic.id]
                if check is False:
                    continue
            except Exception:
                pass

            try:
                buttons = (
                    stream_markup_timer(
                        _,
                        playing[0]["vidid"],
                        chat_id,
                        seconds_to_min(playing[0]["played"]),
                        playing[0]["dur"],
                    )
                    if markup == "stream"
                    else telegram_markup_timer(
                        _,
                        chat_id,
                        seconds_to_min(playing[0]["played"]),
                        playing[0]["dur"],
                    )
                )

                await mystic.edit_reply_markup(
                    reply_markup=InlineKeyboardMarkup(buttons)
                )

            except Exception:
                continue


asyncio.create_task(timer(), name="timer")
asyncio.create_task(markup_timer(), name="markup_timer")
asyncio.create_task(leave_if_muted(), name="leave_if_muted")

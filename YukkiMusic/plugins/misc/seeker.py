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

from pyrogram.enums import ChatType
from pyrogram.types import InlineKeyboardMarkup

from config import MUTE_WARNING_TIME
from strings import get_string
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import db
from YukkiMusic.utils.database import (
    get_active_chats,
    get_assistant,
    get_cmode,
    get_lang,
    is_music_playing,
    set_loop,
)
from YukkiMusic.utils.formatters import seconds_to_min
from YukkiMusic.utils.inline import stream_markup_timer, telegram_markup_timer

from ..admins.callback import wrong

autoend = {}
checker = {}
mute_warnings = {}


async def send_warning(chat_id, text):
    chat = await app.get_chat(chat_id)

    if chat.type == ChatType.CHANNEL:
        cmode_chat_id = await get_cmode(chat_id)
        if cmode_chat_id:
            chat = await app.get_chat(cmode_chat_id)
            if chat.type != ChatType.CHANNEL:
                await app.send_message(cmode_chat_id, text)
    else:
        await app.send_message(chat_id, text)


if MUTE_WARNING_TIME < 60:
    t = f"{MUTE_WARNING_TIME} seconds"
else:
    t = time.strftime("%M:%S minutes", time.gmtime(MUTE_WARNING_TIME))


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


async def process_mute_warnings():
    while True:
        await asyncio.sleep(2)
        for chat_id, details in list(mute_warnings.items()):
            if time.time() - details["timestamp"] >= MUTE_WARNING_TIME:
                _ = details["_"]
                try:
                    userbot = await get_assistant(chat_id)
                    members = []
                    async for member in userbot.get_call_members(chat_id):
                        if member is None:
                            continue
                        members.append(member)

                    autoend[chat_id] = len(members)
                    m = next((m for m in members if m.chat.id == userbot.id), None)
                    if m is None:
                        continue
                    is_muted = bool(m.is_muted and not m.can_self_unmute)

                    if is_muted:
                        await Yukki.stop_stream(chat_id)
                        await set_loop(chat_id, 0)
                        await send_warning(chat_id, _["admin_35"].format(t))

                    mute_warnings.pop(chat_id, None)
                except:
                    mute_warnings.pop(chat_id, None)


async def markup_timer():
    while not await asyncio.sleep(2):
        active_chats = await get_active_chats()
        for chat_id in active_chats:
            if chat_id in mute_warnings:
                continue

            try:
                if not await is_music_playing(chat_id):
                    continue
                playing = db.get(chat_id)
                if not playing:
                    continue
                duration_seconds = int(playing[0]["seconds"])
                if duration_seconds == 0:
                    continue
                try:
                    mystic = playing[0]["mystic"]
                    markup = playing[0]["markup"]
                except:
                    continue
                try:
                    check = wrong[chat_id][mystic.message_id]
                    if check is False:
                        continue
                except:
                    pass
                try:
                    language = await get_lang(chat_id)
                    _ = get_string(language)
                except:
                    _ = get_string("en")

                try:
                    userbot = await get_assistant(chat_id)
                    members = []
                    async for member in userbot.get_call_members(chat_id):
                        if member is None:
                            continue
                        members.append(member)

                    if not members:
                        await Yukki.stop_stream(chat_id)
                        await set_loop(chat_id, 0)
                        continue

                    autoend[chat_id] = len(members)
                    m = next((m for m in members if m.chat.id == userbot.id), None)
                    if m is None:
                        continue
                    is_muted = bool(m.is_muted and not m.can_self_unmute)

                    if is_muted:
                        mute_warnings[chat_id] = {
                            "timestamp": time.time(),
                            "_": _,
                        }

                except:
                    continue

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
                except:
                    continue

            except:
                continue


asyncio.create_task(timer())
asyncio.create_task(markup_timer())
asyncio.create_task(process_mute_warnings())

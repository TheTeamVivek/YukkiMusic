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

from telethon.errors import FloodWaitError

from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.misc import db
from YukkiMusic.utils import (
    get_channeplay_cb,
    parse_flags,
    paste,
    seconds_to_min,
)
from YukkiMusic.utils.database import (
    get_cmode,
    is_active_chat,
    is_music_playing,
)
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.inline.queue import queue_back_markup, queue_markup

basic = {}


def get_duration(playing):
    track = playing[0]["track"]
    if track.is_live or track.is_m3u8 or not track.duration:
        return "UNKNOWM"
    return "INLINE"


@tbot.on_message(flt.command("QUEUE_COMMAND", True) & flt.group & ~BANNED_USERS)
@language
async def ping_com(event, _):
    _, _, is_cplay = parse_flags(event.text, "queue")
    if is_cplay:
        chat_id = await get_cmode(event.chat_id)
        if chat_id is None:
            return await event.reply(_["setting_12"])
        try:
            await tbot.get_entity(chat_id)
        except Exception:
            return await event.reply(_["cplay_4"])
        cplay = True
    else:
        chat_id = event.chat_id
        cplay = False
    if not await is_active_chat(chat_id):
        return await event.reply(_["general_6"])
    got = db.get(chat_id)
    if not got:
        return await event.reply(_["queue_2"])
    track = got[0]["track"]
    videoid = track["vidid"]
    user = got[0]["by"]
    title = track["title"]
    streamtype = track.streamtype.value
    DUR = get_duration(got)
    image = track.thumb or None  # TODO: REPLACE NONE WITH A IMAGE URL OR PATH

    send = (
        "**⌛️ Duration:** Unknown duration limit\n\nClick on below button to get whole queued list"
        if DUR == "UNKNOWN"
        else "\nClick on below button to get whole queued list."
    )
    cap = f"""**{tbot.mention} Player**

🎥**Playing:** {title}

🔗**Stream Type:** {streamtype}
🙍‍♂️**Played By:** {user}
{send}"""
    upl = (
        queue_markup(_, DUR, "c" if cplay else "g", videoid)
        if DUR == "UNKNOWN"
        else queue_markup(
            _,
            DUR,
            "c" if cplay else "g",
            videoid,
            seconds_to_min(got[0]["played"]),
            seconds_to_min(track["duration"]),
        )
    )
    basic[videoid] = True
    mystic = await event.reply(file=image, message=cap, buttons=upl)
    if DUR != "UNKNOWN":
        try:
            while (
                db[chat_id][0]["track"]["vidid"] == videoid
                and basic[videoid]
                and await is_active_chat(chat_id)
            ):
                await asyncio.sleep(5)
                if await is_music_playing(chat_id):
                    try:
                        buttons = queue_markup(
                            _,
                            DUR,
                            "c" if cplay else "g",
                            videoid,
                            seconds_to_min(db[chat_id][0]["played"]),
                            seconds_to_min(db[chat_id][0]["track"]["duration"]),
                        )
                        await mystic.edit(buttons=buttons)
                    except FloodWaitError:
                        pass
                else:
                    pass
        except Exception:
            return


@tbot.on(events.CallbackQuery("GetTimer", func=~BANNED_USERS))
async def quite_timer(event):
    try:
        await event.answer()
    except Exception:
        pass


@tbot.on(events.CallbackQuery("GetQueued", func=~BANNED_USERS))
@language
async def queued_tracks(event, _):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    what, videoid = callback_request.split("|")
    try:
        chat_id, channel = await get_channeplay_cb(_, what, event)
    except Exception:
        return
    if not await is_active_chat(chat_id):
        return await event.answer(_["general_6"], alert=True)
    got = db.get(chat_id)
    if not got:
        return await event.answer(_["queue_2"], alert=True)
    if len(got) == 1:
        return await event.answer(_["queue_5"], alert=True)
    await event.answer()
    basic[videoid] = False
    buttons = queue_back_markup(_, what)
    await event.edit(
        file="https://telegra.ph//file/6f7d35131f69951c74ee5.jpg", text=_["queue_1"]
    )
    msg = ""
    for j, x in enumerate(got):
        if j == 0:
            msg += f'Current playing:\n\n🏷Title: {x["title"]}\nDuration: {x["dur"]}\nBy: {x["by"]}\n\n'
        elif j == 1:
            msg += f'Queued:\n\n🏷Title: {x["title"]}\nDuratiom: {x["dur"]}\nby: {x["by"]}\n\n'
        else:
            msg += f'🏷Title: {x["title"]}\nDuration: {x["dur"]}\nBy: {x["by"]}\n\n'

    await asyncio.sleep(1)
    if len(msg) > 700:
        # msg = msg.replace("🏷", "")
        link = await paste(msg)
        return await event.edit(_["queue_3"].format(link), buttons=buttons)

    return await event.edit(msg, buttons=buttons)


@tbot.on(events.CallbackQuery("queue_back_timer", func=~BANNED_USERS))
@language
async def queue_back(event, _):
    callback_data = event.data.decode("utf-8").strip()
    cplay = callback_data.split(None, 1)[1]
    try:
        chat_id, channel = await get_channeplay_cb(_, cplay, event)
    except Exception:
        return
    if not await is_active_chat(chat_id):
        return await event.answer(_["general_6"], alert=True)
    got = db.get(chat_id)
    if not got:
        return await event.answer(_["queue_2"], alert=True)
    await event.answer(_["set_cb_8"], alert=True)
    file = got[0]["track"]
    videoid = file.vidid
    user = await tbot.create_mention(got[0]["by"])
    title = file.title
    streamtype = file.streamtype.value
    DUR = get_duration(got)
    image = file.thumb
    send = (
        "**⌛️ Duration:** Unknown duration limit\n\nClick on below button to get whole queued list"
        if DUR == "UNKNOWN"
        else "\nClick on below button to get whole queued list."
    )
    cap = f"""**{tbot.mention} Player**

🎥**Playing:** {title}

🔗**Stream Type:** {streamtype}
🙍‍♂️**Played By:** {user}
{send}"""
    upl = (
        queue_markup(_, DUR, cplay, videoid)
        if DUR == "UNKNOWN"
        else queue_markup(
            _,
            DUR,
            cplay,
            videoid,
            seconds_to_min(got[0]["played"]),
            seconds_to_min(got[0]["track"]["duration"]),
        )
    )
    basic[videoid] = True

    mystic = await event.edit(file=image, text=cap, buttons=upl)
    if DUR != "UNKNOWN":
        try:
            while (
                db[chat_id][0]["track"]["vidid"] == videoid
                and basic[videoid]
                and await is_active_chat(chat_id)
            ):
                await asyncio.sleep(5)
                if await is_music_playing(chat_id):
                    try:
                        buttons = queue_markup(
                            _,
                            DUR,
                            cplay,
                            videoid,
                            seconds_to_min(db[chat_id][0]["played"]),
                            seconds_to_min(db[chat_id][0]["track"]["duration"]),
                        )
                        await mystic.edit(buttons=buttons)
                    except FloodWaitError:
                        pass
                else:
                    pass
        except Exception:
            return

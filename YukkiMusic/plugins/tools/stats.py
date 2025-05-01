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
import platform
from sys import version as pyver

import psutil
from pyrogram import __version__ as pyrover
from pytgcalls.__version__ import __version__ as pytgver
from telethon import events
from telethon.errors import MessageIdInvalidError

import config
from strings import get_command
from YukkiMusic import tbot
from YukkiMusic.core import filters
from YukkiMusic.core.mongo import mongodb
from YukkiMusic.core.userbot import assistants
from YukkiMusic.misc import BANNED_USERS, SUDOERS
from YukkiMusic.platforms import youtube
from YukkiMusic.utils.database import (
    get_global_tops,
    get_particulars,
    get_queries,
    get_served_chats,
    get_served_users,
    get_sudoers,
    get_top_chats,
    get_topp_users,
)
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.inline.stats import (
    back_stats_buttons,
    back_stats_markup,
    get_stats_markup,
    overallback_stats_markup,
    stats_buttons,
    top_ten_stats_markup,
)

loop = asyncio.get_running_loop()

STATS_COMMAND = get_command("STATS_COMMAND")
GSTATS_COMMAND = get_command("GSTATS_COMMAND")


@tbot.on_message(filters.command(STATS_COMMAND) & ~BANNED_USERS)
@language
async def stats_global(event, _):
    upl = stats_buttons(_, event.sender_id in SUDOERS)
    await event.reply(
        file=config.STATS_IMG_URL,
        message=_["gstats_11"].format(tbot.mention),
        buttons=upl,
    )


@tbot.on_message(filters.command(GSTATS_COMMAND) & ~BANNED_USERS)
@language
async def gstats_global(event, _):
    mystic = await event.reply(_["gstats_1"])
    stats = await get_global_tops()
    if not stats:
        await asyncio.sleep(1)
        return await mystic.edit(_["gstats_2"])

    def get_stats():
        results = {}
        for i in stats:
            top_list = stats[i]["spot"]
            results[str(i)] = top_list
            list_arranged = dict(
                sorted(
                    results.items(),
                    key=lambda item: item[1],
                    reverse=True,
                )
            )
        if not results:
            return mystic.edit(_["gstats_2"])
        videoid = None
        co = None
        for vidid, count in list_arranged.items():
            if vidid == "telegram":
                continue
            else:
                videoid = vidid
                co = count
            break
        return videoid, co

    try:
        videoid, co = await loop.run_in_executor(None, get_stats)
    except Exception:
        return
    track = await youtube.track(videoid)
    title = track.title.title()
    final = f"Top played Tracks on  {tbot.mention}\n\n**Title:** {title}\n\nPlayed** {co} **times"
    upl = get_stats_markup(_, event.sender_id in SUDOERS)
    await event.respond(
        file=thumbnail,
        message=final,
        buttons=upl,
    )
    await mystic.delete()


@tbot.on(events.CallbackQuery(pattern="GetStatsNow", func=~BANNED_USERS))
@language
async def top_users_ten(event, _):
    chat_id = event.chat_id
    callback_data = event.data.decode("utf-8").strip()
    what = callback_data.split(None, 1)[1]
    upl = back_stats_markup(_)
    try:
        await event.answer()
    except Exception:
        pass
    chat = await event.get_chat()
    mystic = await event.edit(
        _["gstats_3"].format(f"ᴏғ {chat.title}" if what == "Here" else what)
    )
    if what == "Tracks":
        stats = await get_global_tops()
    elif what == "Chats":
        stats = await get_top_chats()
    elif what == "Users":
        stats = await get_topp_users()
    elif what == "Here":
        stats = await get_particulars(chat_id)
    if not stats:
        await asyncio.sleep(1)
        return await mystic.edit(_["gstats_2"], buttons=upl)
    queries = await get_queries()

    def get_stats():
        results = {}
        for i in stats:
            top_list = stats[i] if what in ["Chats", "Users"] else stats[i]["spot"]
            results[str(i)] = top_list
            list_arranged = dict(
                sorted(
                    results.items(),
                    key=lambda item: item[1],
                    reverse=True,
                )
            )
        if not results:
            return mystic.edit(_["gstats_2"], buttons=upl)
        msg = ""
        limit = 0
        total_count = 0
        if what in ["Tracks", "Here"]:
            for items, count in list_arranged.items():
                total_count += count
                if limit == 10:
                    continue
                limit += 1
                details = stats.get(items)
                title = (details["title"][:35]).title()
                if items == "telegram":
                    msg += f"🔗[TelegramVideos and media's](https://t.me/telegram) ** Played {count} Times**\n\n"
                else:
                    msg += f"🔗 [{title}](https://www.youtube.com/watch?v={items}) ** Played {count} Times**\n\n"

            temp = (
                _["gstats_4"].format(
                    queries,
                    tbot.mention,
                    len(stats),
                    total_count,
                    limit,
                )
                if what == "Tracks"
                else _["gstats_7"].format(len(stats), total_count, limit)
            )
            msg = temp + msg
        return msg, list_arranged

    try:
        msg, list_arranged = await loop.run_in_executor(None, get_stats)
    except Exception as e:
        print(e)
        return
    limit = 0
    if what in ["Users", "Chats"]:
        for items, count in list_arranged.items():
            if limit == 10:
                break
            try:
                x = await tbot.get_entity(items)
                extract = x.first_name if what == "Users" else x.title
                if extract is None:
                    continue
                await asyncio.sleep(0.5)
            except Exception:
                continue
            limit += 1
            msg += f"🔗`{extract}` Played {count} Times on bot.\n\n"
        temp = (
            _["gstats_5"].format(limit, tbot.mention)
            if what == "Chats"
            else _["gstats_6"].format(limit, tbot.mention)
        )
        msg = temp + msg
    try:
        await event.edit(file=config.GLOBAL_IMG_URL, text=msg, buttons=upl)
    except MessageIdInvalidError:
        await event.respond(file=config.GLOBAL_IMG_URL, message=msg, buttons=upl)


@tbot.on(events.CallbackQuery(pattern="TopOverall", func=~BANNED_USERS))
@language
async def overall_stats(event, _):
    callback_data = event.data.decode("utf-8").strip()
    what = callback_data.split(None, 1)[1]
    if what != "s":
        upl = overallback_stats_markup(_)
    else:
        upl = back_stats_buttons(_)
    try:
        await event.answer()
    except Exception:
        pass
    await event.edit(_["gstats_8"])
    served_chats = len(await get_served_chats())
    served_users = len(await get_served_users())
    total_queries = await get_queries()
    blocked = len(BANNED_USERS)
    sudoers = len(SUDOERS)
    mod = tbot.loaded_plug_counts
    assistant = len(assistants)
    playlist_limit = config.SERVER_PLAYLIST_LIMIT
    fetch_playlist = config.PLAYLIST_FETCH_LIMIT
    song = config.SONG_DOWNLOAD_DURATION
    play_duration = config.DURATION_LIMIT_MIN
    if config.AUTO_LEAVING_ASSISTANT == str(True):
        ass = "Yes"
    else:
        ass = "No"
    text = f"""**Bot's Stats and information:**

**Imported Modules:** {mod}
**Served chats:** {served_chats}
**Served Users:** {served_users}
**Blocked Users:** {blocked}
**Sudo Users:** {sudoers}

**Total Queries:** {total_queries}
**Total Assistant:** {assistant}
**Auto Leaving Assistsant:** {ass}

**Play Duration Limit:** {play_duration} ᴍɪɴs
**Song Download Limit:** {song} ᴍɪɴs
**Bot's Server Playlist Limit:** {playlist_limit}
**Playlist Play Limit:** {fetch_playlist}"""
    try:
        await event.edit(
            file=config.STATS_IMG_URL, text=text, parse_mode="md", buttons=upl
        )
    except MessageIdInvalidError:
        await event.respond(
            file=config.STATS_IMG_URL, message=text, buttons=upl, parse_mode="md"
        )


@tbot.on(events.CallbackQuery(pattern="bot_stats_sudo"))
@language
async def overall_stats(event, _):
    if event.sender_id not in SUDOERS:
        return await event.answer("ᴏɴʟʏ ғᴏʀ sᴜᴅᴏ ᴜsᴇʀ's", alert=True)
    callback_data = event.data.decode("utf-8").strip()
    what = callback_data.split(None, 1)[1]
    if what != "s":
        upl = overallback_stats_markup(_)
    else:
        upl = back_stats_buttons(_)
    try:
        await event.answer()
    except Exception:
        pass
    await event.edit(_["gstats_8"])
    sc = platform.system()
    p_core = psutil.cpu_count(logical=False)
    t_core = psutil.cpu_count(logical=True)
    ram = str(round(psutil.virtual_memory().total / (1024.0**3))) + " GB"
    try:
        cpu_freq = psutil.cpu_freq().current
        if cpu_freq >= 1000:
            cpu_freq = f"{round(cpu_freq / 1000, 2)}GHz"
        else:
            cpu_freq = f"{round(cpu_freq, 2)}MHz"
    except Exception:
        cpu_freq = "Unable to Fetch"
    hdd = psutil.disk_usage("/")
    total = hdd.total / (1024.0**3)
    total = str(total)
    used = hdd.used / (1024.0**3)
    used = str(used)
    free = hdd.free / (1024.0**3)
    free = str(free)
    mod = int(tbot.loaded_plug_counts)
    call = await mongodb.command("dbstats")
    datasize = call["dataSize"] / 1024
    datasize = str(datasize)
    storage = call["storageSize"] / 1024
    objects = call["objects"]
    collections = call["collections"]

    served_chats = len(await get_served_chats())
    served_users = len(await get_served_users())
    total_queries = await get_queries()
    blocked = len(BANNED_USERS)
    sudoers = len(await get_sudoers())
    text = f""" **Bot Stats and information:**

**Imported modules:** {mod}
**Platform:** {sc}
**Ram:** {ram}
**Physical Cores:** {p_core}
**Total Cores:** {t_core}
**Cpu frequency:** {cpu_freq}

**Python Version:** {pyver.split()[0]}
**Pyrogram Version:** {pyrover}
**Py-tgcalls Version:** {pytgver}
**Total Storage:** {total[:4]} ɢiʙ
**Storage Used:** {used[:4]} ɢiʙ
**Storage Left:** {free[:4]} ɢiʙ

**Served chats:** {served_chats}
**Served users:** {served_users}
**Blocked users:** {blocked}
**Sudo users:** {sudoers}

**Total DB Storage:** {storage} ᴍʙ
**Total DB Collection:** {collections}
**Total DB Keys:** {objects}
**Total Bot Queries:** `{total_queries} `
    """
    try:
        await event.edit(
            file=config.STATS_IMG_URL, text=text, parse_mode="md", buttons=upl
        )
    except MessageIdInvalidError:
        await event.respond(
            file=config.STATS_IMG_URL, message=text, parse_mode="md", buttons=upl
        )


@tbot.on(
    events.CallbackQuery(
        pattern=r"^(TOPMARKUPGET|GETSTATS|GlobalStats)$",
        func=~BANNED_USERS,
    )
)
@language
async def back_buttons(event, _):
    try:
        await event.answer()
    except Exception:
        pass
    command = event.pattern_match.group(1).decode("utf-8")
    if command == "TOPMARKUPGET":
        upl = top_ten_stats_markup(_)

        try:
            await event.edit(
                file=config.GLOBAL_IMG_URL, message=_["gstats_9"], buttons=upl
            )
        except MessageIdInvalidError:
            await event.respond(
                file=config.GLOBAL_IMG_URL,
                message=_["gstats_9"],
                buttons=upl,
            )
    if command == "GlobalStats":
        upl = get_stats_markup(
            _,
            event.sender_id in SUDOERS,
        )

        try:
            await event.edit(
                file=config.GLOBAL_IMG_URL,
                text=_["gstats_10"].format(tbot.mention),
                buttons=upl,
            )

        except MessageIdInvalidError:
            await event.respond(
                file=config.GLOBAL_IMG_URL,
                message=_["gstats_10"].format(tbot.mention),
                buttons=upl,
            )

    if command == "GETSTATS":
        upl = stats_buttons(
            _,
            event.sender_id in SUDOERS,
        )

        try:
            await event.edit(
                file=config.GLOBAL_IMG_URL,
                text=_["gstats_11"].format(tbot.mention),
                buttons=upl,
            )

        except MessageIdInvalidError:
            await event.respond(
                file=config.GLOBAL_IMG_URL,
                message=_["gstats_11"].format(tbot.mention),
                buttons=upl,
            )

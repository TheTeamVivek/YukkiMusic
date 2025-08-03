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
from pyrogram import filters
from pyrogram.errors import MessageIdInvalid
from pyrogram.types import InputMediaPhoto, Message
from pytgcalls.__version__ import __version__ as pytgver

import config
from config import BANNED_USERS
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.core.mongo import mongodb
from yukkimusic.core.userbot import assistants
from yukkimusic.misc import SUDOERS
from yukkimusic.platforms import youtube
from yukkimusic.utils.database import (
    get_global_tops,
    get_particulars,
    get_queries,
    get_served_chats,
    get_served_users,
    get_sudoers,
    get_top_chats,
    get_topp_users,
)
from yukkimusic.utils.decorators import asyncify, language
from yukkimusic.utils.inline.stats import (
    back_stats_buttons,
    back_stats_markup,
    get_stats_markup,
    overallback_stats_markup,
    stats_buttons,
    top_ten_stats_markup,
)

from . import mhelp


@app.on_message(command("STATS_COMMAND") & ~BANNED_USERS)
@language
async def stats_global(client, message: Message, _):
    upl = stats_buttons(_, True if message.from_user.id in SUDOERS else False)
    await message.reply_photo(
        photo=config.STATS_IMG_URL,
        caption=_["gstats_11"].format(app.mention),
        reply_markup=upl,
    )


@app.on_message(command("GSTATS_COMMAND") & ~BANNED_USERS)
@language
async def gstats_global(client, message: Message, _):
    mystic = await message.reply_text(_["gstats_1"])
    stats = await get_global_tops()
    if not stats:
        await asyncio.sleep(1)
        return await mystic.edit(_["gstats_2"])

    @asyncify
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
        videoid, co = await get_stats()
    except Exception as e:
        print(e)
        return
    (
        title,
        duration_min,
        duration_sec,
        thumbnail,
        vidid,
    ) = await youtube.details(videoid, True)
    title = title.title()
    final = f"Top played Tracks on  {app.mention}\n\n**Title:** {title}\n\nPlayed** {co} **times"
    upl = get_stats_markup(_, True if message.from_user.id in SUDOERS else False)
    await app.send_photo(
        message.chat.id,
        photo=thumbnail,
        caption=final,
        reply_markup=upl,
    )
    await mystic.delete()


@app.on_callback_query(filters.regex("GetStatsNow") & ~BANNED_USERS)
@language
async def top_users_ten(client, query, _):
    chat_id = query.message.chat.id
    callback_data = query.data.strip()
    what = callback_data.split(None, 1)[1]
    upl = back_stats_markup(_)
    try:
        await query.answer()
    except Exception:
        pass
    mystic = await query.edit_message_text(
        _["gstats_3"].format(
            f"ᴏғ {query.message.chat.title}" if what == "Here" else what
        )
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
        return await mystic.edit(_["gstats_2"], reply_markup=upl)
    queries = await get_queries()

    @asyncify
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
            return mystic.edit(_["gstats_2"], reply_markup=upl)
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
                    app.mention,
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
        msg, list_arranged = await get_stats()
    except Exception as e:
        print(e)
        return
    limit = 0
    if what in ["Users", "Chats"]:
        for items, count in list_arranged.items():
            if limit == 10:
                break
            try:
                extract = (
                    (await app.get_users(items)).first_name
                    if what == "Users"
                    else (await app.get_chat(items)).title
                )
                if extract is None:
                    continue
                await asyncio.sleep(0.5)
            except Exception:
                continue
            limit += 1
            msg += f"🔗`{extract}` Played {count} Times on bot.\n\n"
        temp = (
            _["gstats_5"].format(limit, app.mention)
            if what == "Chats"
            else _["gstats_6"].format(limit, app.mention)
        )
        msg = temp + msg
    med = InputMediaPhoto(media=config.GLOBAL_IMG_URL, caption=msg)
    try:
        await query.edit_message_media(media=med, reply_markup=upl)
    except MessageIdInvalid:
        await query.message.reply_photo(
            photo=config.GLOBAL_IMG_URL, caption=msg, reply_markup=upl
        )


@app.on_callback_query(filters.regex("TopOverall") & ~BANNED_USERS)
@language
async def top_overall_stats(client, query, _):
    callback_data = query.data.strip()
    what = callback_data.split(None, 1)[1]
    if what != "s":
        upl = overallback_stats_markup(_)
    else:
        upl = back_stats_buttons(_)
    try:
        await query.answer()
    except Exception:
        pass
    await query.edit_message_text(_["gstats_8"])
    served_chats = len(await get_served_chats())
    served_users = len(await get_served_users())
    total_queries = await get_queries()
    blocked = len(BANNED_USERS)
    sudoers = len(SUDOERS)
    mod = int(app.loaded_plug_counts)
    assistant = len(assistants)
    playlist_limit = config.SERVER_PLAYLIST_LIMIT
    fetch_playlist = config.PLAYLIST_FETCH_LIMIT
    song = config.SONG_DOWNLOAD_DURATION
    play_duration = config.DURATION_LIMIT_MIN
    if config.AUTO_LEAVING_ASSISTANT:
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
    med = InputMediaPhoto(media=config.STATS_IMG_URL, caption=text)
    try:
        await query.edit_message_media(media=med, reply_markup=upl)
    except MessageIdInvalid:
        await query.message.reply_photo(
            photo=config.STATS_IMG_URL, caption=text, reply_markup=upl
        )


@app.on_callback_query(filters.regex("bot_stats_sudo"))
@language
async def bot_stats(client, query, _):
    if query.from_user.id not in SUDOERS:
        return await query.answer("ᴏɴʟʏ ғᴏʀ sᴜᴅᴏ ᴜsᴇʀ's", show_alert=True)
    callback_data = query.data.strip()
    what = callback_data.split(None, 1)[1]
    if what != "s":
        upl = overallback_stats_markup(_)
    else:
        upl = back_stats_buttons(_)
    try:
        await query.answer()
    except Exception:
        pass
    await query.edit_message_text(_["gstats_8"])
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
    mod = int(app.loaded_plug_counts)
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
    med = InputMediaPhoto(media=config.STATS_IMG_URL, caption=text)
    try:
        await query.edit_message_media(media=med, reply_markup=upl)
    except MessageIdInvalid:
        await query.message.reply_photo(
            photo=config.STATS_IMG_URL, caption=text, reply_markup=upl
        )


@app.on_callback_query(
    filters.regex(pattern=r"^(TOPMARKUPGET|GETSTATS|GlobalStats)$") & ~BANNED_USERS
)
@language
async def back_buttons(client, query, _):
    try:
        await query.answer()
    except Exception:
        pass
    command = query.matches[0].group(1)
    if command == "TOPMARKUPGET":
        upl = top_ten_stats_markup(_)
        med = InputMediaPhoto(
            media=config.GLOBAL_IMG_URL,
            caption=_["gstats_9"],
        )
        try:
            await query.edit_message_media(media=med, reply_markup=upl)
        except MessageIdInvalid:
            await query.message.reply_photo(
                photo=config.GLOBAL_IMG_URL,
                caption=_["gstats_9"],
                reply_markup=upl,
            )
    if command == "GlobalStats":
        upl = get_stats_markup(
            _,
            True if query.from_user.id in SUDOERS else False,
        )
        med = InputMediaPhoto(
            media=config.GLOBAL_IMG_URL,
            caption=_["gstats_10"].format(app.mention),
        )
        try:
            await query.edit_message_media(media=med, reply_markup=upl)
        except MessageIdInvalid:
            await query.message.reply_photo(
                photo=config.GLOBAL_IMG_URL,
                caption=_["gstats_10"].format(app.mention),
                reply_markup=upl,
            )
    if command == "GETSTATS":
        upl = stats_buttons(
            _,
            True if query.from_user.id in SUDOERS else False,
        )
        med = InputMediaPhoto(
            media=config.STATS_IMG_URL,
            caption=_["gstats_11"].format(app.mention),
        )
        try:
            await query.edit_message_media(media=med, reply_markup=upl)
        except MessageIdInvalid:
            await query.message.reply_photo(
                photo=config.STATS_IMG_URL,
                caption=_["gstats_11"].format(app.mention),
                reply_markup=upl,
            )


(
    mhelp.add(
        "en",
        f"<b>★ {pick_commands('GSTATS_COMMAND')}</b> - Get Top 10 Tracks Global Stats, Top 10 Users of Bot, Top 10 Chats on Bot, Top 10 Played in a chat, etc.",
    )
    .add(
        "ar",
        f"<b>★ {pick_commands('GSTATS_COMMAND')}</b> - احصل على أفضل 10 إحصائيات للأغاني عالميًا، أفضل 10 مستخدمين للبوت، أفضل 10 محادثات في البوت، أفضل 10 تشغيلات في محادثة، إلخ.",
    )
    .add(
        "as",
        f"<b>★ {pick_commands('GSTATS_COMMAND')}</b> - শীৰ্ষ 10 ট্ৰেক গ্ল'বেল ষ্টেটচ, বটৰ শীৰ্ষ 10 ব্যৱহাৰকাৰী, বটৰ শীৰ্ষ 10 চেট, চেটত শীৰ্ষ 10 প্লে আদি পাব।",
    )
    .add(
        "hi",
        f"<b>★ {pick_commands('GSTATS_COMMAND')}</b> - टॉप 10 ट्रैक्स ग्लोबल स्टैट्स, बॉट के टॉप 10 यूज़र्स, बॉट के टॉप 10 चैट्स, चैट में टॉप 10 प्ले आदि प्राप्त करें।",
    )
    .add(
        "ku",
        f"<b>★ {pick_commands('GSTATS_COMMAND')}</b> - 10 گۆرانییە باشترینەکان لە جیھاندا، 10 بەکارهێنەرە سەرەکییەکانی بۆت، 10 چاتی سەرەکیی بۆت، 10 یاری زۆرترین لە چات، هتد بگرە.",
    )
    .add(
        "tr",
        f"<b>★ {pick_commands('GSTATS_COMMAND')}</b> - En iyi 10 Parça Global İstatistikleri, Botun En İyi 10 Kullanıcısı, Botun En İyi 10 Sohbeti, Sohbette En Çok Çalınan 10 Parça vb. alın.",
    )
)

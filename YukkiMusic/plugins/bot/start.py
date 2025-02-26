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
import time

from pyrogram import filters
from pyrogram.enums import ChatType, ParseMode
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message
from youtubesearchpython.__future__ import VideosSearch

from telethon import Button, events
import config
from config import BANNED_USERS, START_IMG_URL
from config.config import OWNER_ID
from strings import command, get_string
from YukkiMusic import Platform, tbot
from YukkiMusic.platforms import youtube, telegram
from YukkiMusic.misc import SUDOERS, _boot_
from YukkiMusic.plugins.bot.help import paginate_modules
from YukkiMusic.plugins.play.playlist import del_plist_msg
from YukkiMusic.plugins.sudo.sudoers import sudoers_list
from YukkiMusic.utils.database import (
    add_served_chat,
    add_served_user,
    blacklisted_chats,
    get_assistant,
    get_lang,
    get_userss,
    is_on_off,
    is_served_private_chat,
)
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.formatters import get_readable_time
from YukkiMusic.utils.functions import MARKDOWN, WELCOMEHELP
from YukkiMusic.utils.inline import private_panel, start_pannel

loop = asyncio.get_running_loop()


@tbot.on_message(flt.command("START_COMMAND", True) & flt.private & ~flt.user(BANNED_USERS))
@language(no_check=True)
async def start_comm(event, _):
    chat_id = event.chat_id
    await add_served_user(event.sender_id)
    if len(event.text.split()) > 1:
        name = event.text.split(None, 1)[1]
        if name[0:4] == "help":
            keyboard = await paginate_modules(0, chat_id, close=True)

            if config.START_IMG_URL:
                return await event.reply(
                    file=START_IMG_URL,
                    message=_["help_1"],
                    buttons=keyboard,
                )
            else:
                return await event.reply(
                    message=_["help_1"],
                    buttons=keyboard,
                )
        if name[0:4] == "song":
            await event.reply(_["song_2"])
            return
        if name == "mkdwn_help":
            await message.reply(
                MARKDOWN,
                parse_mode="HTML",
                link_preview=False,
            )
        if name == "greetings":
            await message.reply(
                WELCOMEHELP,
                parse_mode="HTML",
                link_preview=False,
            )
        if name[0:3] == "sta":
            m = await event.reply("🔎 Fetching Your personal stats.!")
            stats = await get_userss(event.sender_id)
            tot = len(stats)
            if not stats:
                await asyncio.sleep(1)
                return await m.edit(_["ustats_1"])

            def get_stats():
                msg = ""
                limit = 0
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
                    return m.edit(_["ustats_1"])
                tota = 0
                videoid = None
                for vidid, count in list_arranged.items():
                    tota += count
                    if limit == 10:
                        continue
                    if limit == 0:
                        videoid = vidid
                    limit += 1
                    details = stats.get(vidid)
                    title = (details["title"][:35]).title()
                    if vidid == "telegram":
                        msg += f"🔗[Telegram Files and Audio]({config.SUPPORT_GROUP}) ** played {count} Times**\n\n"
                    else:
                        msg += f"🔗 [{title}](https://www.youtube.com/watch?v={vidid}) ** played {count} Times**\n\n"
                msg = _["ustats_2"].format(tot, tota, limit) + msg
                return videoid, msg

            try:
                videoid, msg = await loop.run_in_executor(None, get_stats)
            except Exception as e:
                return
            thumbnail = await youtube.thumbnail(videoid, True)
            await m.delete()
            await event.reply(file=thumbnail, message=msg)
            return
        if name[0:3] == "sud":
            await sudoers_list(client=event.client, message=event, _=_)
            await asyncio.sleep(1)
            if await is_on_off(config.LOG):
            	sender = await event.get_sender()
                sender_id = sender.id
                sender_name = sender.first_name
                return await tbot.send_message(
                    config.LOG_GROUP_ID,
                    f"{await tbot.create_mention(sender)} Has just started bot to check `Sudolist`\n\n**User Id:** {sender_id}\n**User Name:** {sender_name}",
                )
            return
        if name[0:3] == "lyr":
            query = (str(name)).replace("lyrics_", "", 1)
            lyrical = config.lyrical
            lyrics = lyrical.get(query)
            if lyrics:
                async for chunk_msg in telegram.split_text(lyrics):
                    await event.reply(chunk_msg)
                return
            else:
                await event.reply("Failed to get lyrics ")
                return
        if name[0:3] == "del":
            await del_plist_msg(client=event.client, message=event, _=_)
            await asyncio.sleep(1)
        if name[0:3] == "inf":
            m = await event.reply("🔎 Fetching info..")
            query = (str(name)).replace("info_", "", 1)
            query = f"https://www.youtube.com/watch?v={query}"
            results = VideosSearch(query, limit=1)
            for result in (await results.next())["result"]:
                title = result["title"]
                duration = result["duration"]
                views = result["viewCount"]["short"]
                thumbnail = result["thumbnails"][0]["url"].split("?")[0]
                channellink = result["channel"]["link"]
                channel = result["channel"]["name"]
                link = result["link"]
                published = result["publishedTime"]
            searched_text = f"""
🔍__**Video track information **__

❇️**Title:** {title}

⏳**Duration:** {duration} Mins
👀**Views:** `{views}`
⏰**Published times:** {published}
🎥**Channel Name:** {channel}
📎**Channel Link:** [Visit from here]({channellink})
🔗**Videp linl:** [Link]({link})
"""
            key =   [
                    [
                        Button.url(text="🎥 Watch ", url=f"{link}"),
                        Button.inline(text="🔄 Close", data="close"),
                    ],
                ]
            await m.delete()
            await event.respond(
                file=thumbnail,
                message=searched_text,
                buttons=key,
            )
            await asyncio.sleep(1)
            if await is_on_off(config.LOG):
            	sender = await event.get_sender()
                sender_id = event.sender_id
                sender_name = sender.first_name
                return await tbot.send_message(
                    config.LOG_GROUP_ID,
                    f"{await tbot.create_mention(sender)} Has just started bot ot check `Video information `\n\n**User Id:** {sender_id}\n**User Name** {sender_name}",
                )
    else:
        try:
            await tbot.get_entity(OWNER_ID[0])
            OWNER = OWNER_ID[0]
        except Exception:
            OWNER = None
        out = private_panel(_, OWNER)
        if config.START_IMG_URL:
            try:
                await event.reply(
                    file=config.START_IMG_URL,
                    message=_["start_1"].format(tbot.mention),
                    buttons=out,
                )
            except Exception:
                await event.reply(
                    message=_["start_1"].format(tbot.mention),
                    buttons=out,
                )
        else:
            await event.reply(
                message=_["start_1"].format(tbot.mention),
                buttons=out,
            )
        if await is_on_off(config.LOG):
            sender_id = event.sender_id
            sender = await event.get_sender()
            sender_name = sender.first_name
            return await tbot.send_message(
                config.LOG_GROUP_ID,
                f"{await tbot.create_mention(sender)} Has started bot. \n\n**User id :** {sender_id}\n**User name:** {sender_name}",
            )


@tbot.on_message(flt.command("START_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS))
@language(no_check=True)
async def testbot(event , _):
    uptime = int(time.time() - _boot_)
    await event.reply(_["start_7"].format(get_readable_time(uptime)))
    return await add_served_chat(event.chat_id)


@tbot.on_message(filters.new_chat_members, group=-1) # Not completed yet
async def welcome(client, message: Message):
    chat_id = event.chat_id
    if config.PRIVATE_BOT_MODE == str(True):
        if not await is_served_private_chat(event.chat_id):
            await event.reply(
                "This Bot's private mode has been enabled only my owner can use this if want to use in your chat so say my Owner to authorize your chat."
            )
            return await tbot.leave_chat(event.chat_id)
    else:
        await add_served_chat(chat_id)
    for member in message.new_chat_members:
        try:
            language = await get_lang(event.chat_id)
            _ = get_string(language)
            if member.id == tbot.id:
                chat_type = message.chat.type
                if chat_type != ChatType.SUPERGROUP:
                    await event.reply(_["start_5"])
                    return await tbot.leave_chat(event.chat_id)
                if chat_id in await blacklisted_chats():
                    await event.reply(
                        _["start_6"].format(
                            f"https://t.me/{tbot.username}?start=sudolist"
                        )
                    )
                    return await tbot.leave_chat(chat_id)
                userbot = await get_assistant(event.chat_id)
                out = start_pannel(_)
                await event.reply(
                    _["start_2"].format(
                        tbot.mention,
                        userbot.username,
                        userbot.id,
                    ),
                    buttons=out,
                )
            if member.id in config.OWNER_ID:
                return await event.reply(
                    _["start_3"].format(tbot.mention, member.mention)
                )
            if member.id in SUDOERS:
                return await event.reply(
                    _["start_4"].format(tbot.mention, member.mention)
                )
            return
        except Exception:
            return
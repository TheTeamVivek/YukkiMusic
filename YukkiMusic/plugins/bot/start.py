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

from telethon import Button, events, utils
from telethon.extensions import markdown
from telethon.tl.types import Channel, Chat
from youtubesearchpython.__future__ import VideosSearch

import config
from config import OWNER_ID, START_IMG_URL
from strings import get_string
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.misc import BANNED_USERS, SUDOERS, _boot_
from YukkiMusic.platforms import youtube
from YukkiMusic.plugins.bot.help import paginate_modules
from YukkiMusic.plugins.play.playlist import del_group_message as del_plist_msg
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


@tbot.on_message(flt.command("START_COMMAND", True) & flt.private & ~BANNED_USERS)
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
            await event.reply(
                MARKDOWN,
                parse_mode="HTML",
                link_preview=False,
            )
        if name == "greetings":
            await event.reply(
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
            except Exception:
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
                text, entities = markdown.parse(lyrics)
                for text, entities in utils.split_text(text, entities):
                    await event.reply(text, formatting_entities=entities)
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
            key = [
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


@tbot.on_message(flt.command("START_COMMAND", True) & flt.group & ~BANNED_USERS)
@language(no_check=True)
async def start_group(event, _):
    uptime = int(time.time() - _boot_)
    await event.reply(_["start_7"].format(get_readable_time(uptime)))
    return await add_served_chat(event.chat_id)


@tbot.on(events.ChatAction(func=flt.new_chat_members))
async def welcome(event):
    chat_id = event.chat_id
    chat = await event.get_chat()

    if isinstance(chat, Channel) and not chat.megagroup:
        return

    if config.PRIVATE_BOT_MODE:
        if not await is_served_private_chat(chat_id):
            await event.reply(
                "This Bot's private mode has been enabled. Only my owner can use this. "
                "If you want to use it in your chat, ask my Owner to authorize your chat."
            )
            return await tbot.leave_chat(chat_id)
    else:
        await add_served_chat(chat_id)
        language = await get_lang(chat_id)
        _ = get_string(language)

        if isinstance(chat, Chat):
            await event.reply(_["start_5"])
            return await tbot.leave_chat(chat_id)

    for user in event.users:
        if user.id == tbot.id:
            if chat_id in await blacklisted_chats():
                await event.reply(
                    _["start_6"].format(f"https://t.me/{tbot.username}?start=sudolist")
                )
                return await tbot.leave_chat(chat_id)

            userbot = await get_assistant(chat_id)
            out = start_pannel(_)
            await event.reply(
                _["start_2"].format(
                    tbot.mention,
                    userbot.username,
                    userbot.id,
                ),
                buttons=out,
            )
            continue

        mention = await tbot.create_mention(user)

        if user.id in config.OWNER_ID:
            await event.reply(_["start_3"].format(tbot.mention, mention))

        elif user.id in SUDOERS:
            await event.reply(_["start_4"].format(tbot.mention, mention))

        else:  # We can add check about the user is banned on bot and try to kick from the chat
            pass

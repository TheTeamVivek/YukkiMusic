#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import random
import time

from pyrogram import Client, filters
from pyrogram.enums import ParseMode
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message
from youtubesearchpython.__future__ import VideosSearch

import config
from config import BANNED_USERS, PHOTO
from config.config import OWNER_ID
from strings import get_string
from YukkiMusic import Telegram, YouTube
from YukkiMusic.misc import SUDOERS, _boot_
from YukkiMusic.plugins.play.playlist import del_plist_msg
from YukkiMusic.plugins.sudo.sudoers import sudoers_list
from YukkiMusic.utils.database import (
    blacklisted_chats,
    get_assistant,
    get_lang,
)
from YukkiMusic.utils.decorators.language import LanguageStart
from YukkiMusic.utils.formatters import get_readable_time
from YukkiMusic.utils.inline import (
    alive_panel,
    help_pannel,
    private_panel,
    start_pannel,
)

loop = asyncio.get_running_loop()


@Client.on_message(filters.command(["start"]) & filters.private & ~BANNED_USERS)
@LanguageStart
async def start_comm(client, message: Message, _):
    chat_id = message.chat.id
    me = await client.get_me()
    if len(message.text.split()) > 1:
        name = message.text.split(None, 1)[1]
        if name[0:4] == "help":
            keyboard = help_pannel(_)
            if config.START_IMG_URL:
                return await message.reply_photo(
                    photo=config.START_IMG_URL,
                    caption=_["help_1"],
                    reply_markup=keyboard,
                )
                
            else:
                return await message.reply_photo(
                    photo=random.choice(PHOTO),
                    caption=_["help_1"],
                    reply_markup=keyboard,
                )
        if name[0:4] == "song":
            return await message.reply_text(_["song_2"])     
        if name[0:3] == "lyr":
            query = (str(name)).replace("lyrics_", "", 1)
            lyrical = config.lyrical
            lyrics = lyrical.get(query)
            if lyrics:
                return await Telegram.send_split_text(message, lyrics)
                
            else:
                return await message.reply_text("ғᴀɪʟᴇᴅ ᴛᴏ ɢᴇᴛ ʟʏʀɪᴄs.")
        if name[0:3] == "del":
            await del_plist_msg(client=client, message=message, _=_)
            await asyncio.sleep(1)
        if name[0:3] == "inf":
            m = await message.reply_text("🔎 ғᴇᴛᴄʜɪɴɢ ɪɴғᴏ!")
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
🔍__**ᴠɪᴅᴇᴏ ᴛʀᴀᴄᴋ ɪɴғᴏʀᴍᴀᴛɪᴏɴ**__

❇️**ᴛɪᴛʟᴇ:** {title}

⏳**ᴅᴜʀᴀᴛɪᴏɴ:** {duration} Mins
👀**ᴠɪᴇᴡs:** `{views}`
⏰**ᴘᴜʙʟɪsʜᴇᴅ ᴛɪᴍᴇ:** {published}
🎥**ᴄʜᴀɴɴᴇʟ ɴᴀᴍᴇ:** {channel}
📎**ᴄʜᴀɴɴᴇʟ ʟɪɴᴋ:** [ᴠɪsɪᴛ ғʀᴏᴍ ʜᴇʀᴇ]({channellink})
🔗**ᴠɪᴅᴇᴏ ʟɪɴᴋ:** [ʟɪɴᴋ]({link})

⚡️ __sᴇᴀʀᴄʜᴇᴅ ᴘᴏᴡᴇʀᴇᴅ ʙʏ {me.mention}__"""
            key = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(text="🎥 ᴡᴀᴛᴄʜ ", url=f"{link}"),
                        InlineKeyboardButton(text="🔄 ᴄʟᴏsᴇ", callback_data="close"),
                    ],
                ]
            )
            await m.delete()
            await client.send_photo(
                message.chat.id,
                photo=thumbnail,
                caption=searched_text,
                parse_mode=ParseMode.MARKDOWN,
                reply_markup=key,
            )
    else:
        try:
            await client.resolve_peer(OWNER_ID[0])
            OWNER = OWNER_ID[0]
        except:
            OWNER = None
        out = private_panel(_, client.username, OWNER)
        era = await message.reply_text(
            text=f"{message.from_user.first_name} जय श्री राधे कृष्णा जी, आपका {me.mention} में हार्दिक स्वागत है।"
        )
        await asyncio.sleep(0.5)
        await era.delete()
        if config.START_IMG_URL:
            try:
                await message.reply_photo(
                    photo=config.START_IMG_URL,
                    caption=_["start_2"].format(me.mention),
                    reply_markup=InlineKeyboardMarkup(out),
                )
            except:
                await message.reply_photo(
                    photo=random.choice(PHOTO),
                    caption=_["start_2"].format(me.mention),
                    reply_markup=InlineKeyboardMarkup(out),
                )
        else:
            await message.reply_photo(
                photo=random.choice(PHOTO),
                caption=_["start_2"].format(me.mention),
                reply_markup=InlineKeyboardMarkup(out),
            )
            
@Client.on_message(filters.command(["start"]) & filters.group & ~BANNED_USERS)
@LanguageStart
async def testbot(client, message: Message, _):
    out = alive_panel(_)
    me = await client.get_me()
    uptime = int(time.time() - _boot_)
    if config.START_IMG_URL: 	
        return await message.reply_photo(
            photo=config.START_IMG_URL,
            caption=_["start_8"].format(me.mention, get_readable_time(uptime)),
            reply_markup=InlineKeyboardMarkup(out),
        )
    else:
    	return await message.reply_photo(
            photo=random.choice(PHOTO),
            caption=_["start_8"].format(me.mention, get_readable_time(uptime)),
            reply_markup=InlineKeyboardMarkup(out),
        )
    
@Client.on_message(filters.new_chat_members, group=-1)
async def welcome(client, message: Message):
    chat_id = message.chat.id
    me = await client.get_me()
    for member in message.new_chat_members:
        try:
            language = await get_lang(message.chat.id)
            _ = get_string(language)
            if member.id == me.id:
                if chat_id in await blacklisted_chats():
                    await message.reply_text(
                        _["start_7"].format(
                            f"https://t.me/{me.username}?start=sudolist"
                        )
                    )
                    return await client.leave_chat(chat_id)
                userbot = await get_assistant(chat_id)
                out = start_pannel(_)
                try:
                	await userbot.join_chat(chat_id)
                except Exception:
                	pass
                
                await message.reply_text(
                    _["start_3"].format(
                        me.mention,
                        userbot.username,
                        userbot.id,
                    ),
                    reply_markup=InlineKeyboardMarkup(out),
                )
            return
        except:
            return
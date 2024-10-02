#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from pyrogram import filters
from pyrogram.types import Message

from config import BANNED_USERS, LOG_GROUP_ID
from YukkiMusic import app, userbot
from YukkiMusic.core.userbot import assistants
from YukkiMusic.utils.database import get_assistant, save_assistant, set_assistant, get_client
from YukkiMusic.utils.decorators import AdminActual
from YukkiMusic.core.userbot import assistants

async def get_assistant_details():
    
    msg = "**ᴜsᴀsɢᴇ** : /setassistant [ᴀssɪsᴛᴀɴᴛ ɴᴏ ] ᴛᴏ ᴄʜᴀɴɢᴇ ᴀɴᴅ sᴇᴛ ᴍᴀɴᴜᴀʟʟʏ ɢʀᴏᴜᴘ ᴀssɪsᴛᴀɴᴛ \n ʙᴇʟᴏᴡ sᴏᴍᴇ ᴀᴠᴀɪʟᴀʙʟᴇ ᴀssɪsᴛᴀɴᴛ ᴅᴇᴛᴀɪʟ's\n"
    for cnt in assistants:
        try:
            a = await get_client(cnt)
            msg += f"ᴀssɪsᴛᴀɴᴛ ɴᴜᴍʙᴇʀ:- `{cnt}` \nɴᴀᴍᴇ :- [{a.name}](https://t.me/{a.username})  \nᴜsᴇʀɴᴀᴍᴇ :-  @{a.username} \nɪᴅ :- {a.id}\n\n"
        except:
            pass
    return msg


@app.on_message(filters.command("changeassistant") & ~BANNED_USERS)
@AdminActual
async def assis_change(client, message: Message, _):
    if len(userbot.clients) == 1:
        return await message.reply_text(
            "sᴏʀʀʏ sɪʀ! ɪɴ ʙᴏᴛ sᴇʀᴠᴇʀ ᴏɴʟʏ ᴏɴʀ ᴀssɪsᴛᴀɴᴛ ᴀᴠᴀɪʟᴀʙʟᴇ ᴛʜᴇʀᴇғᴏʀᴇ ʏᴏᴜ ᴄᴀɴᴛ ᴄʜᴀɴɢᴇ ᴀssɪsᴛᴀɴᴛ"
        )
    a = await get_assistant(message.chat.id)
    DETAILS = f"ʏᴏᴜʀ ᴄʜᴀᴛ's ᴀssɪsᴛᴀɴᴛ ʜᴀs ʙᴇᴇɴ ᴄʜᴀɴɢᴇᴅ ғʀᴏᴍ [{a.name}](https://t.me/{a.username}) "
    if not message.chat.id == LOG_GROUP_ID:
        try:
            await a.leave_chat(message.chat.id)
        except:
            pass
    b = await set_assistant(message.chat.id)
    DETAILS += f"ᴛᴏ [{b.name}](https://t.me/{b.username})"
    try:
        await b.join_chat(message.chat.id)
    except:
        pass
    await message.reply_text(DETAILS, disable_web_page_preview=True)


@app.on_message(filters.command("setassistant") & ~BANNED_USERS)
@AdminActual
async def assis_set(client, message: Message, _):
    if len(userbot.clients) == 1:
        return await message.reply_text(
            "sᴏʀʀʏ ɪɴ ʙᴏᴛ sᴇʀᴠᴇʀ ᴏɴʟʏ ᴏɴᴇ ᴀssɪsᴛᴀɴᴛ ᴀᴠᴀɪʟᴀʙʟᴇ ᴛʜᴇʀᴇғᴏʀᴇ ʏᴏᴜ ᴄᴀɴ'ᴛ ᴄʜᴀɴɢᴇ ᴀssɪsᴛᴀɴᴛ"
        )
    query = message.text.split(None, 1)[1].strip()
    if query not in assistants:
        return await message.reply_text(usage, disable_web_page_preview=True)
    a = await get_assistant(message.chat.id)
    if not message.chat.id == LOG_GROUP_ID:
        try:
            await a.leave_chat(message.chat.id)
        except:
            pass
    await save_assistant(message.chat.id, query)
    b = await get_assistant(message.chat.id)
    try:
        await b.join_chat(message.chat.id)
    except:
        pass
    await message.reply_text(
        "**Yᴏᴜʀ ᴄʜᴀᴛ's ɴᴇᴡ ᴀssɪsᴛᴀɴᴛ ᴅᴇᴛᴀɪʟs:**\nAssɪsᴛᴀɴᴛ Nᴀᴍᴇ :- {b.name}\nUsᴇʀɴᴀᴍᴇ :- @{b.username}\nID:- {b.id}",
        disable_web_page_preview=True,
    )


@app.on_message(filters.command("checkassistant") & filters.group & ~BANNED_USERS)
@AdminActual
async def check_ass(client, message: Message, _):
    a = await get_assistant(message.chat.id)
    await message.reply_text(
        "**Yᴏᴜʀ ᴄʜᴀᴛ's ᴀssɪsᴛᴀɴᴛ ᴅᴇᴛᴀɪʟs:**\nAssɪsᴛᴀɴᴛ Nᴀᴍᴇ :- {a.name}\nAssɪsᴛᴀɴᴛ\nUsᴇʀɴᴀᴍᴇ :- @{a.username}\nAssɪsᴛᴀɴᴛ ID:- {a.id}",
        disable_web_page_preview=True,
    )
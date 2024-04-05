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
from pyrogram import Client, filters
from pyrogram.types import Message
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import get_assistant


@app.on_message(filters.command(["leave"]) & SUDOERS)
async def leave_group(client: Client, message: Message):
    if len(message.command) != 2:
        await message.reply_text("ɢɪᴠᴇ ᴍᴇ ᴀ ᴄʜᴀᴛɪᴅ ᴀғᴛᴇʀ /leave ᴛᴏ ʟᴇᴀᴠᴇ")
        return

    group_id = message.command[1]
    try:
        chat = await client.get_chat(int(group_id))

        if chat is None:
            await message.reply(
                "ɪ ᴛʜɪɴᴋ ᴛʜᴇ ᴄʜᴀᴛɪᴅ ɪs ᴡʀᴏɴɢ ᴄᴀɴ ʏᴏᴜ ᴄʜᴇᴄᴋ ᴛʜɪs ᴀɢᴀɪɴ ᴘʟᴇᴀsᴇ"
            )
            return

        try:
            lol = await message.reply(f"ʟᴇᴀᴠɪɴ ғʀᴏᴍ {chat.title}")
            await client.leave_chat(int(group_id))
            await asyncio.sleep(1)
            await lol.edit(f"ʙᴏᴛ ʟᴇғᴛᴇᴅ ғʀᴏᴍ {chat.title}")
        except Exception as e:
            await message.reply(f"sᴏᴍᴇ ᴇxᴄᴇᴘᴛɪᴏɴ ᴡʜɪʟᴇ ʟᴇᴀᴠɪɴɢ \n {e}")
            return
    except ValueError:
        await message.reply("ɪɴᴠᴀʟɪᴅ ᴄʜᴀᴛ ɪᴅ")

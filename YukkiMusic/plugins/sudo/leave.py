#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram import filters

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import delete_served_chat, get_assistant


@app.on_message(filters.command("leave") & SUDOERS)
async def bot_assistant_leave(_, message):
    if len(message.command) != 2:
        await message.reply_text("ɢɪᴠᴇ ᴍᴇ ᴀ ᴄʜᴀᴛɪᴅ ᴀғᴛᴇʀ /leave ᴛᴏ ʟᴇᴀᴠᴇ")
        return

    group_id = message.command[1]
    try:
        chat = await app.get_chat(int(group_id))

        if chat is None:
            await message.reply(
                "ɪ ᴛʜɪɴᴋ ʙᴏᴛ ɪs ɴᴏᴛ ɪɴ ᴘʀᴏᴠɪᴅᴇᴅ ᴄʜᴀᴛ ɪᴅ ᴏʀ ᴛʜᴇ ᴄʜᴀᴛ ɪᴅ ɪs ᴡʀᴏɴɢ ᴄᴀɴ ʏᴏᴜ ᴄʜᴇᴄᴋ ᴛʜɪs ᴀɢᴀɪɴ ᴘʟᴇᴀsᴇ"
            )
            return

        try:
            ass = await get_assistant(int(group_id))
            lol = await message.reply(f"ʟᴇᴀᴠɪɴɢ ғʀᴏᴍ {chat.title}")
            await app.leave_chat(int(group_id))
            await delete_served_chat(int(group_id))
            await lol.edit(f"ʙᴏᴛ ʟᴇᴀᴠᴇᴅ ғʀᴏᴍ {chat.title}")
            await ass.leave_chat(int(group_id))

        except Exception as e:
            print(e)  # Handle or log the exception as per your requirement

    except Exception as e:
        print(e)  # Handle or log the exception as per your requirement

#
# Copyright (C) 2024 - present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from config import LOG_GROUP_ID
from YukkiMusic import app
from YukkiMusic.utils.database import is_on_off


async def play_logs(message, streamtype):
    if await is_on_off(2):
        if message.chat.username:
            chatusername = f"@{message.chat.username}"
        else:
            chatusername = "ᴘʀɪᴠᴀᴛᴇ ɢʀᴏᴜᴘ"

        logger_text = f"""
**{app.mention} ᴘʟᴀʏ ʟᴏɢ**

**ᴄʜᴀᴛ ɪᴅ :** `{message.chat.id}`
**ᴄʜᴀᴛ ɴᴀᴍᴇ :** {message.chat.title}
**ᴄʜᴀᴛ ᴜsᴇʀɴᴀᴍᴇ :** {chatusername}

**ᴜsᴇʀ ɪᴅ :** `{message.from_user.id}`
**ɴᴀᴍᴇ :** {message.from_user.mention}
**ᴜsᴇʀɴᴀᴍᴇ :** @{message.from_user.username}

**ǫᴜᴇʀʏ :** {message.text.split(None, 1)[1]}
**sᴛʀᴇᴀᴍᴛʏᴘᴇ :** {streamtype}"""
        if message.chat.id != LOG_GROUP_ID:
            try:
                await app.send_message(
                    chat_id=LOG_GROUP_ID,
                    text=logger_text,
                    disable_web_page_preview=True,
                )
            except Exception as e:
                print(e)
        return

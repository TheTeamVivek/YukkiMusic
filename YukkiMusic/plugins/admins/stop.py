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

from config import BANNED_USERS
from strings import get_command
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import set_loop, delete_filter
from YukkiMusic.utils.decorators import AdminRightsCheck

# Commands
STOP_COMMAND = get_command("STOP_COMMAND")


@app.on_message(filters.command(STOP_COMMAND) & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def stop_music(cli, message: Message, _, chat_id):
    if len(message.command) < 2:
        await Yukki.stop_stream(chat_id)
        await set_loop(chat_id, 0)
        await message.reply_text(_["admin_9"].format(message.from_user.mention))
    else:
        filter = " ".join(message.command[1:])
        deleted = await delete_filter(message.chat.id, name)
        if deleted:
            await message.reply_text(f"**ᴅᴇʟᴇᴛᴇᴅ ғɪʟᴛᴇʀ {name}.**")
        else:
            await message.reply_text("**ɴᴏ sᴜᴄʜ ғɪʟᴛᴇʀ.**")

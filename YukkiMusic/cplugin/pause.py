#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram import Client, filters
from pyrogram.types import Message

from config import BANNED_USERS
from strings import get_command
from YukkiMusic.core.gcall import Yukki
from YukkiMusic.utils.database import is_music_playing, music_off
from YukkiMusic.utils.decorators import CAdminRightsCheck

# Commands
PAUSE_COMMAND = get_command("PAUSE_COMMAND")


@Client.on_message(filters.command(PAUSE_COMMAND) & filters.group & ~BANNED_USERS)
@CAdminRightsCheck
async def pause_admin(cli, message: Message, _, chat_id):
    if not len(message.command) == 1:
        return await message.reply_text(_["general_2"])
    if not await is_music_playing(chat_id):
        return await message.reply_text(_["admin_1"])
    await music_off(chat_id)
    await Yukki.pause_stream(chat_id)
    await message.reply_text(_["admin_2"].format(message.from_user.mention))

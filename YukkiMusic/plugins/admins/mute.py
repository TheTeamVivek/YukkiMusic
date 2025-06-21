#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
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
from strings import command
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import is_muted, mute_on
from YukkiMusic.utils.decorators import AdminRightsCheck


@app.on_message(command("MUTE_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def mute_admin(cli, message: Message, _, chat_id):
    if not len(message.command) == 1 or message.reply_to_message:
        return
    if await is_muted(chat_id):
        return await message.reply_text(_["mute_1"])
    await mute_on(chat_id)
    await Yukki.mute_stream(chat_id)
    await message.reply_text(_["mute_2"].format(message.from_user.mention))

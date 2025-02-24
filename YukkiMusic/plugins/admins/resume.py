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
from YukkiMusic import app, tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import is_music_playing, music_on
from YukkiMusic.utils.decorators import admin_rights_check


@tbot.on_message(flt.command("RESUME_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS))
@admin_rights_check
async def resume_com(event, _, chat_id):
    if not len(event.text.split()) == 1:
        return await event.reply(_["COMMAND_USAGE_ERROR"])
    if await is_music_playing(chat_id):
        return await event.reply(_["admin_3"])
    await music_on(chat_id)
    await Yukki.resume_stream(chat_id)
    await event.reply(_["admin_4"].format(await tbot.create_mention(await event.get_sender())))

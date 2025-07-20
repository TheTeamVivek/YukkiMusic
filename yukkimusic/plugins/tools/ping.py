#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import time

from pyrogram import types

from config import BANNED_USERS, PING_IMG_URL
from strings import command
from yukkimusic import app
from yukkimusic.core.call import yukki
from yukkimusic.utils import bot_sys_stats
from yukkimusic.utils.decorators.language import language
from yukkimusic.utils.inline import support_group_markup


@app.on_message(command("PING_COMMAND") & ~BANNED_USERS)
@language
async def ping_com(_, message: types.Message, lang):
    start = time.time()
    m = await message.reply_text(
        lang["ping_1"].format(app.mention),
    )
    end = time.time()
    pytgping = await yukki.ping()
    up, cpu, ram, disk = await bot_sys_stats()
    resp = round((end - start) * 1000, 3)
    await m.edit_media(
        media=types.InputMediaPhoto(
            media=PING_IMG_URL,
            caption=lang["ping_2"].format(
                resp,
                app.mention,
                up,
                ram,
                cpu,
                disk,
                pytgping,
            ),
        ),
        reply_markup=support_group_markup(lang),
    )
    await message.stop_propagation()

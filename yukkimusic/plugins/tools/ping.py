#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import time

import psutil
from pyrogram import types

from config import BANNED_USERS, PING_IMG_URL
from strings import command
from yukkimusic import app
from yukkimusic.core.call import yukki
from yukkimusic.misc import _boot_
from yukkimusic.utils.decorators.language import language
from yukkimusic.utils.formatters import get_readable_time
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
    uptime = get_readable_time(int(time.time() - _boot_))
    resp = round((end - start) * 1000, 3)
    cpu_usage = f"{await asyncio.to_thread(psutil.cpu_percent, interval=0.5)}%"
    ram_usage = f"{psutil.virtual_memory().percent}%"
    disk_usage = f"{psutil.disk_usage('/').percent}%"

    await m.edit_media(
        media=types.InputMediaPhoto(
            media=PING_IMG_URL,
            caption=lang["ping_2"].format(
                resp,
                app.mention,
                uptime,
                ram_usage,
                cpu_usage,
                disk_usage,
                pytgping,
            ),
        ),
        reply_markup=support_group_markup(lang),
    )
    await message.stop_propagation()

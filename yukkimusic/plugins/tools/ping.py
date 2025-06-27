#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#
from datetime import datetime

from pyrogram.types import Message

from config import BANNED_USERS, PING_IMG_URL
from strings import command
from yukkimusic import app
from yukkimusic.core.call import yukki
from yukkimusic.utils import bot_sys_stats
from yukkimusic.utils.decorators.language import language
from yukkimusic.utils.inline import support_group_markup


@app.on_message(command("PING_COMMAND") & ~BANNED_USERS)
@language
async def ping_com(client, message: Message, _):
    response = await message.reply_photo(
        photo=PING_IMG_URL,
        caption=_["ping_1"].format(app.mention),
    )
    start = datetime.now()
    pytgping = await yukki.ping()
    UP, CPU, RAM, DISK = await bot_sys_stats()
    resp = (datetime.now() - start).microseconds / 1000
    await response.edit_text(
        _["ping_2"].format(
            resp,
            app.mention,
            UP,
            RAM,
            CPU,
            DISK,
            pytgping,
        ),
        reply_markup=support_group_markup(_),
    )

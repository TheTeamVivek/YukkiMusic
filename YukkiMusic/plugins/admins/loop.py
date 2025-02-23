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
from YukkiMusic.utils.database.memorydatabase import get_loop, set_loop
from YukkiMusic.utils.decorators import admin_rights_check

@tbot.on_message(flt.command("LOOP_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS))
@admin_rights_check
async def admins(event, _, chat_id):
    usage = _["admin_24"]
    if len(event.text.split()) != 2:
        return await event.reply(usage)
    state = event.text.split(None, 1)[1].strip()
    await event.get_sender()
    if state.isnumeric():
        state = int(state)
        if 1 <= state <= 10:
            got = await get_loop(chat_id)
            if got != 0:
                state = got + state
            if int(state) > 10:
                state = 10
            await set_loop(chat_id, state)
            return await event.reply(
                _["admin_25"].format(event.sender.first_name, state)
            )
        else:
            return await event.reply(_["admin_26"])
    elif state.lower() == "enable":
        await set_loop(chat_id, 10)
        return await event.reply(
            _["admin_25"].format(event.sender.first_name, 10)
        )
    elif state.lower() == "disable":
        await set_loop(chat_id, 0)
        return await event.reply(_["admin_27"])
    else:
        return await event.reply(usage)
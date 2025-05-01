#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from telethon.tl.types import ChannelParticipantsAdmins

from config import adminlist
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.misc import BANNED_USERS
from YukkiMusic.utils.database import get_authuser_names
from YukkiMusic.utils.decorators import language
from YukkiMusic.utils.formatters import alpha_to_int


@tbot.on_message(flt.command("RELOAD_COMMAND", True) & flt.group & ~BANNED_USERS)
@language
async def reload_admin_cache(event, _):
    try:
        chat_id = event.chat_id
        authusers = await get_authuser_names(chat_id)
        adminlist[chat_id] = []
        async for user in tbot.iter_participants(
            chat_id, filter=ChannelParticipantsAdmins
        ):
            if user.participant.admin_rights.manage_call:
                adminlist[chat_id].append(user.id)
        for user in authusers:
            user_id = await alpha_to_int(user)
            adminlist[chat_id].append(user_id)
        await event.reply(_["admin_20"])
    except Exception:
        await event.reply(_["admin_21"])

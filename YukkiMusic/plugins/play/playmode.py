#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.utils.database import (
    get_playmode,
    get_playtype,
    is_nonadmin_chat,
)
from YukkiMusic.utils.decorators import language
from YukkiMusic.utils.inline.settings import playmode_users_markup


@tbot.on_message(flt.command("PLAYMODE_COMMAND", True) & flt.group & ~BANNED_USERS)
@language
async def playmode_(event, _):
    chat_id = event.chat_id
    chat_title = (await event.get_chat()).title
    playmode = await get_playmode(chat_id)
    if playmode == "DIRECT":
        direct = True
    else:
        direct = None
    is_non_admin = await is_nonadmin_chat(chat_id)
    if not is_non_admin:
        group = True
    else:
        group = None
    playty = await get_playtype(chat_id)
    if playty == "EVERYONE":
        playtype = None
    else:
        playtype = True
    buttons = playmode_users_markup(_, direct, group, playtype)
    await event.reply(
        _["playmode_1"].format(chat_title),
        buttons=buttons,
    )

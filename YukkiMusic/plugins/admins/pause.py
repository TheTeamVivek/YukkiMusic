#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from strings import get_command
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import BANNED_USERS
from YukkiMusic.utils.database import is_music_playing, music_off
from YukkiMusic.utils.decorators import admin_rights_check

PAUSE_COMMAND = get_command("PAUSE_COMMAND")


@tbot.on_message(flt.command(PAUSE_COMMAND) & flt.group & ~BANNED_USERS)
@admin_rights_check
async def pause_admin(event, _, chat_id):
    # if not len(event.text.split()) == 1:
    #    return await event.reply(_["general_2"])
    if not await is_music_playing(chat_id):
        return await event.reply(_["admin_1"])
    await music_off(chat_id)
    await Yukki.pause_stream(chat_id)
    await event.reply(_["admin_2"].format((await event.get_sender()).first_name))

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
from YukkiMusic.utils.database import is_muted, mute_on
from YukkiMusic.utils.decorators import admin_rights_check

MUTE_COMMAND = get_command("MUTE_COMMAND")

@tbot.on_message(flt.command(MUTE_COMMAND) & flt.group & ~BANNED_USERS)
@admin_rights_check
async def mute_admin(event, _, chat_id):
    if not len(event.text.split()) == 1 or event.is_reply:
        return
    if await is_muted(chat_id):
        return await event.reply(_["admin_5"], link_preview=False)
    await mute_on(chat_id)
    await Yukki.mute_stream(chat_id)
    mention = await tbot.create_mention(await event.get_sender())
    await event.reply(_["admin_6"].format(mention), link_preview=False)

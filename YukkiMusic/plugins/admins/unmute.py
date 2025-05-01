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
from YukkiMusic.utils.database import is_muted, mute_off
from YukkiMusic.utils.decorators import admin_rights_check

UNMUTE_COMMAND = get_command("UNMUTE_COMMAND")


@tbot.on_message(flt.command(UNMUTE_COMMAND) & flt.group & ~BANNED_USERS)
@admin_rights_check
async def unmute_admin(event, _, chat_id):
    if not len(event.text.split()) == 1 or event.is_reply:
        return
    if not await is_muted(chat_id):
        return await event.reply(_["admin_7"], link_preview=False)
    await mute_off(chat_id)
    await Yukki.unmute_stream(chat_id)
    mention = await tbot.create_mention(await event.get_sender())
    await event.reply(_["admin_8"].format(mention), link_preview=False)

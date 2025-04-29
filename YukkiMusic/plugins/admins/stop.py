#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import BANNED_USERS, db
from YukkiMusic.utils.database import set_loop
from YukkiMusic.utils.decorators import admin_rights_check


@tbot.on_message(
    flt.command("STOP_COMMAND", True) & flt.group & ~BANNED_USERS
)
@admin_rights_check
async def stop_music(event, _, chat_id):
    try:
        check = db.get(chat_id)
        if check[0].get("mystic"):
            await check[0].get("mystic").delete()
    except Exception:
        pass

    await Yukki.stop_stream(chat_id)
    await set_loop(chat_id, 0)
    mention = await tbot.create_mention(await event.get_sender())
    await event.reply(_["admin_9"].format(mention))

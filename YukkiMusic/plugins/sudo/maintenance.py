#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from strings import get_string
from YukkiMusic import tbot
from YukkiMusic.core import filters
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils import (
    get_lang,
    is_maintenance,
    maintenance_off,
    maintenance_on,
)


@tbot.on_message(filters.command("MAINTENANCE_COMMAND", True) & SUDOERS)
async def maintenance(event):
    try:
        language = await get_lang(event.chat_id)
        _ = get_string(language)
    except Exception:
        _ = get_string("en")
    usage = _["maint_1"]
    if len(event.text.split()) != 2:
        return await event.reply(usage)
    state = event.message.text.split(None, 1)[1].strip()
    state = state.lower()
    if state == "enable":
        if await is_maintenance() is False:
            await event.reply(_["maint_6"])
        else:
            await maintenance_on()
            await event.reply(_["maint_2"])
    elif state == "disable":
        if await is_maintenance() is False:
            await maintenance_off()
            await event.reply(_["maint_3"])
        else:
            await event.reply(_["maint_5"])
    else:
        await event.reply(usage)

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
from YukkiMusic.utils.database import autoend_off, autoend_on


@tbot.on_message(flt.command("AUTOEND_COMMAND", True) & flt.user(BANNED_USERS))
async def auto_end_stream(event):
    usage = "**Usage:**\n\n/autoend [enable|disable]"
    if len(event.text.split()) != 2:
        return await event.reply(usage)
    state = event.text.split(None, 1)[1].strip()
    state = state.lower()
    if state == "enable":
        await autoend_on()
        await event.reply(
            "Auto End enabled.\n\nBot will leave voicechat automatically after 30 secinds if one is listening song with a warning message.."
        )
    elif state == "disable":
        await autoend_off()
        await event.reply("Autoend disabled")
    else:
        await event.reply(usage)

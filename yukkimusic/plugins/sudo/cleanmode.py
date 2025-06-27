#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import config
from strings import command
from yukkimusic import app
from yukkimusic.misc import SUDOERS
from yukkimusic.utils.database import add_off, add_on
from yukkimusic.utils.decorators.language import language


@app.on_message(command("CLEANMODE_COMMAND") & SUDOERS)
@language
async def setcleanmode(client, message, _):
    if len(message.command) < 2 or message.command[1].lower() not in ["on", "off"]:
        return await message.reply_text(
            "Usage:\n"
            "`/cleanmode on` - Enable cleanmode\n"
            "`/cleanmode off` - Disable cleanmode"
        )

    state = message.text.split(None, 1)[1].strip().lower()

    if state == "on":
        await add_on(config.CLEANMODE)
        return await message.reply_text(
            "✅ Cleanmode has been successfully **enabled**."
        )

    elif state == "off":
        await add_off(config.CLEANMODE)
        return await message.reply_text(
            "❌ Cleanmode has been successfully **disabled**."
        )

    return await message.reply_text(
        "Invalid option.\n"
        "Usage:\n"
        "`/cleanmode on` - Enable cleanmode\n"
        "`/cleanmode off` - Disable cleanmode"
    )

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import random

from pyrogram import filters
from pyrogram.types import Message

from config import BANNED_USERS
from strings import command
from YukkiMusic import app, tbot
from YukkiMusic.misc import db
from YukkiMusic.utils.decorators import admin_rights_check


@tbot.on_message(flt.command("SHUFFLE_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS))
@admin_rights_check
async def admins(event_, chat_id):
    if not len(event.text.split()) == 1:
        return await event.reply(_["COMMAND_USAGE_ERROR"])
    check = db.get(chat_id)
    if not check:
        return await event.reply(_["admin_21"])
    try:
        popped = check.pop(0)
    except Exception:
        return await event.reply(_["admin_22"])
    check = db.get(chat_id)
    if not check:
        check.insert(0, popped)
        return await event.reply(_["admin_22"])
    random.shuffle(check)
    check.insert(0, popped)
    mention = await tbot.create_mention(await event.get_sender())
    await event.reply(_["admin_23"].format(mention))

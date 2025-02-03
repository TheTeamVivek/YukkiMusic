#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from decorator import decorator
from telethon.tl.types import User

from strings import get_string
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_lang,
    is_commanddelete_on,
    is_maintenance,
)


@decorator
async def language(func, event, no_check=False):
    chat = await event.get_chat()
    try:
        language = await get_lang(chat.id)
        language = get_string(language)
    except Exception:
        language = get_string("en")

    if no_check:
        return await func(event, language)

    if not await is_maintenance():
        if event.sender_id not in SUDOERS:
            if isinstance(chat, User):
                if event.message:
                    return await event.reply(language["maint_4"])
                return await event.answer(language["maint_4"], alert=True)
            return

    if event.message and await is_commanddelete_on(chat.id):
        try:
            await update.delete()
        except Exception:
            pass

    return await func(event, language)

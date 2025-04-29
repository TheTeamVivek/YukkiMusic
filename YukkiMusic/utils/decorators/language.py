#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import functools

from telethon.tl.types import User

from strings import get_string
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_lang,
    is_commanddelete_on,
    is_maintenance,
)

__all__ = ["language"]


def language(func=None, *, no_check=False):
    def decorator(f):
        @functools.wraps(f)
        async def wrapper(event):
            chat = await event.get_chat()
            try:
                lang_code = await get_lang(chat.id)
                language = get_string(lang_code)
            except Exception:
                language = get_string("en")

            if no_check:
                return await f(event, language)

            if not await is_maintenance():
                if event.sender_id not in SUDOERS:
                    if isinstance(chat, User):
                        if event.message:
                            return await event.reply(language["maint_4"])
                        return await event.answer(language["maint_4"], alert=True)
                    return

            if event.message and await is_commanddelete_on(chat.id):
                try:
                    await event.delete()
                except Exception:
                    pass

            return await f(event, language)

        return wrapper

    if func is None:
        return decorator
    else:
        return decorator(func)

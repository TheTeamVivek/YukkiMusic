#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from functools import wraps

from pyrogram import types
from pyrogram.enums import ChatType

from strings import get_string
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_lang,
    is_commanddelete_on,
    is_maintenance,
)


def language(_func=None, *, no_check=False):
    def decorator(func):
        @wraps(func)
        async def wrapper(client, update: types.Message | types.CallbackQuery):
            is_callback = isinstance(update, types.CallbackQuery)
            chat = update.message.chat if is_callback else update.chat
            chat_id = chat.id
            language = await get_lang(chat_id)
            language = get_string(language)
            if no_check:
                return await func(client, update)
            if await is_maintenance():
                if update.from_user.id not in SUDOERS:
                    if chat.type == ChatType.PRIVATE:
                        if is_callback:
                            return await update.answer(
                                language["maint_4"],
                                show_alert=True,
                            )
                        return await update.reply_text(language["maint_4"])
                return
            if await is_commanddelete_on(chat_id) and not is_callback:
                try:
                    await update.delete()
                except Exception:
                    pass
            return await func(client, update, language)

        return wrapper

    if _func is None:
        return decorator
    elif callable(_func):
        return decorator(_func)

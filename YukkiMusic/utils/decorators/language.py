#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from pyrogram.enums import ChatType

from strings import get_string
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_lang,
    is_commanddelete_on,
    is_maintenance,
)


def language(mystic):
    async def wrapper(_, message, **kwargs):
        try:
            language = await get_lang(message.chat.id)
            language = get_string(language)
        except Exception:
            language = get_string("en")
        if not await is_maintenance():
            if message.from_user.id not in SUDOERS:
                if message.chat.type == ChatType.PRIVATE:
                    return await message.reply_text(language["maint_4"])
                return
        if await is_commanddelete_on(message.chat.id):
            try:
                await message.delete()
            except Exception:
                pass
        return await mystic(_, message, language)

    return wrapper


def languageCB(mystic):
    async def wrapper(_, query, **kwargs):
        try:
            language = await get_lang(query.message.chat.id)
            language = get_string(language)
        except Exception:
            language = get_string("en")
        if not await is_maintenance():
            if query.from_user.id not in SUDOERS:
                if query.message.chat.type == ChatType.PRIVATE:
                    return await query.answer(
                        language["maint_4"],
                        show_alert=True,
                    )
                return

        return await mystic(_, query, language)

    return wrapper


def LanguageStart(mystic):
    async def wrapper(_, message, **kwargs):
        try:
            language = await get_lang(message.chat.id)
            language = get_string(language)
        except Exception:
            language = get_string("en")
        return await mystic(_, message, language)

    return wrapper

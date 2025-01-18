#
# Copyright (C) 2024-2025-2025-2025-2025-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from decorator import decorator
from pyrogram.enums import ChatType
from pyrogram.types import CallbackQuery, Message

from strings import get_string
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_lang,
    is_commanddelete_on,
    is_maintenance,
)


@decorator
async def language(
    func,
    client,
    update: Message | CallbackQuery,
    no_check: bool = False,  # no_check this is used in start for let don't check about maintnece
):
    is_message = isinstance(update, Message)

    try:
        chat_id = update.chat.id if is_message else update.message.chat.id
        user_id = update.from_user.id if is_message else update.from_user.id
        chat_type = update.chat.type if is_message else update.message.chat.type

        language = await get_lang(chat_id)
        language = get_string(language)
    except Exception:
        language = get_string("en")

    if no_check:
        return await func(client, update, language)

    if not await is_maintenance():
        if user_id not in SUDOERS:
            if chat_type == ChatType.PRIVATE:
                if is_message:
                    return await update.reply_text(language["maint_4"])
                return await update.answer(language["maint_4"], show_alert=True)
            return

    if is_message and await is_commanddelete_on(chat_id):
        try:
            await update.delete()
        except Exception:
            pass

    return await func(client, update, language)

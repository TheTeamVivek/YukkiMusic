#
# Copyright (C) 2024-2025-2025-2025-2025-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from YukkiMusic import app
from YukkiMusic.utils.database import get_cmode


async def get_channeplay_cb(_, command, query):
    if command == "c":
        chat_id = await get_cmode(query.message.chat.id)
        if chat_id is None:
            try:
                return await query.answer(_["setting_12"], show_alert=True)
            except Exception:
                return
        try:
            chat = await app.get_chat(chat_id)
            channel = chat.title
        except Exception:
            try:
                return await query.answer(_["cplay_4"], show_alert=True)
            except Exception:
                return
    else:
        chat_id = query.message.chat.id
        channel = None
    return chat_id, channel

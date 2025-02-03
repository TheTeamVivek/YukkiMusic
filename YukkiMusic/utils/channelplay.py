#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from YukkiMusic.utils.database import get_cmode


async def get_channeplay_cb(_, command, event):
    if command == "c":
        chat_id = await get_cmode(event.chat_id)
        if chat_id is None:
            try:
                return await event.answer(_["setting_12"], alert=True)
            except Exception:
                return
        try:
            chat = await event.client.get_entity(chat_id)
            channel = chat.title
        except Exception:
            try:
                return await event.answer(_["cplay_4"], alert=True)
            except Exception:
                return
    else:
        chat_id = event.chat_id
        channel = None
    return chat_id, channel

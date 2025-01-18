#
# Copyright (C) 2024-2025-2025-2025-2025-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
from config import LOG, LOG_GROUP_ID
from YukkiMusic import tbot
from YukkiMusic.utils.database import is_on_off


async def play_logs(message, streamtype):
    if await is_on_off(LOG):
        if message.chat.username:
            chatusername = f"@{message.chat.username}"
        else:
            chatusername = "Private Group"

        logger_text = f"""
**{tbot.mention} Play Log**

**Chat ID:** `{message.chat.id}`
**Chat Name:** {message.chat.title}
**Chat Username:** {chatusername}

**User ID:** `{message.from_user.id}`
**Name:** {message.from_user.mention}
**Username:** @{message.from_user.username}

**Query:** {message.text.split(None, 1)[1]}
**Stream Type:** {streamtype}"""
        if message.chat.id != LOG_GROUP_ID:
            try:
                await tbot.send_message(
                    entity=LOG_GROUP_ID,
                    message=logger_text,
                    link_preview=False,
                )
            except Exception:
                pass
        return

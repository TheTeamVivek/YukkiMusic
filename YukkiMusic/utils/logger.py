#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
from config import LOG, LOG_GROUP_ID
from YukkiMusic.utils.database import is_on_off

__all__ = ["play_logs"]


async def play_logs(_, event, streamtype):
    if await is_on_off(LOG):
        from YukkiMusic import tbot
        chat = await event.get_chat()
        if chat.username:
            chatusername = f"@{chat.username}"
        else:
            chatusername = "Private Group"
        sender = await event.get_sender()
        if event.is_reply:
            query = "Replied Message"
        else:
            query = event.text.split(None, 1)[1]
        logger_text = _["logger_text"].format(
            bot_mention=tbot.mention,
            chat_id=event.chat_id,
            title=chat.title,
            chatusername=chatusername,
            sender_id=event.sender_id,
            user_mention=await tbot.create_mention(sender),
            username=f"@{sender.username}" if sender.username else "No Username",
            query=query,
            streamtype=streamtype,
        )
        if event.chat_id != LOG_GROUP_ID:
            try:
                await tbot.send_message(
                    entity=LOG_GROUP_ID,
                    message=logger_text,
                    link_preview=False,
                )
            except Exception:
                pass
        return

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
from datetime import datetime, timedelta

from pyrogram.enums import ChatType

import config
from strings import get_string
from YukkiMusic import tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import (
    get_assistant,
    get_client,
    get_lang,
    is_active_chat,
    is_autoend,
)

autoend = {}
next_run = {"auto_leave": datetime.utcnow(), "auto_end": datetime.utcnow()}


async def auto_leave():
    if config.AUTO_LEAVING_ASSISTANT:
        return

    if datetime.utcnow() < next_run["auto_leave"]:
        return

    from YukkiMusic.core.userbot import assistants

    async def leave_inactive_chats(client):
        left = 0
        try:
            async for dialog in client.get_dialogs():
                chat = dialog.chat
                if chat.type in {ChatType.SUPERGROUP, ChatType.GROUP, ChatType.CHANNEL}:
                    chat_id = chat.id
                    if chat_id not in {
                        config.LOG_GROUP_ID,
                        -1002159045835,
                        -1002146211959,
                    }:
                        if left >= 20:
                            break
                        if not await is_active_chat(chat_id):
                            try:
                                await client.leave_chat(chat_id)
                                left += 1
                            except Exception:
                                continue
        except Exception:
            pass

    tasks = [leave_inactive_chats(await get_client(num)) for num in assistants]
    await asyncio.gather(*tasks)

    next_run["auto_leave"] = datetime.utcnow() + timedelta(
        seconds=config.ASSISTANT_LEAVE_TIME
    )


async def auto_end():
    if datetime.utcnow() < next_run["auto_end"]:
        return

    if not await is_autoend():
        return

    for chat_id, timer in list(autoend.items()):
        if datetime.utcnow() > timer:
            if not await is_active_chat(chat_id):
                del autoend[chat_id]
                continue

            userbot = await get_assistant(chat_id)
            members = []

            try:
                async for member in userbot.get_call_members(chat_id):
                    if member:
                        members.append(member)
            except ValueError:
                try:
                    await Yukki.stop_stream(chat_id)
                except Exception:
                    pass
                continue

            if len(members) <= 1:
                try:
                    await Yukki.stop_stream(chat_id)
                    language = get_string(await get_lang(chat_id) or "en")
                    await tbot.send_message(chat_id, language["misc_1"])
                except Exception:
                    pass

            del autoend[chat_id]

    next_run["auto_end"] = datetime.utcnow() + timedelta(seconds=30)


async def run_all_tasks():
    while True:
        await asyncio.gather(auto_leave(), auto_end())
        await asyncio.sleep(1)


asyncio.create_task(run_all_tasks())

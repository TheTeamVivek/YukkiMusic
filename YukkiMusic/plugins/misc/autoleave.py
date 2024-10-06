#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import asyncio
from datetime import datetime

from pyrogram.enums import ChatType

import config
from strings import get_string
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import (
    get_assistant,
    get_client,
    get_lang,
    is_active_chat,
    is_autoend,
)

autoend = {}


async def auto_leave():
    if config.AUTO_LEAVING_ASSISTANT == str(True):
        from YukkiMusic.core.userbot import assistants

        async def leave_inactive_chats(client):
            left = 0
            try:
                async for i in client.get_dialogs():
                    chat_type = i.chat.type
                    if chat_type in [
                        ChatType.SUPERGROUP,
                        ChatType.GROUP,
                        ChatType.CHANNEL,
                    ]:
                        chat_id = i.chat.id
                        if chat_id not in [
                            config.LOG_GROUP_ID,
                            -1002159045835,
                            -1002146211959,
                        ]:
                            if left == 20:
                                break
                            if not await is_active_chat(chat_id):
                                try:
                                    await client.leave_chat(chat_id)
                                    left += 1
                                except:
                                    continue
            except:
                pass

        while not await asyncio.sleep(config.AUTO_LEAVE_ASSISTANT_TIME):
            tasks = []
            for num in assistants:
                client = await get_client(num)
                tasks.append(leave_inactive_chats(client))

            # Using asyncio.gather for running the leave_inactive_chats and same time for all assistant
            await asyncio.gather(*tasks)


async def auto_end():
    while True:
        await asyncio.sleep(30)
        if not await is_autoend():
            continue
        for chat_id, timer in list(autoend.items()):
            if datetime.now() > timer:
                if not await is_active_chat(chat_id):
                    del autoend[chat_id]
                    continue

                userbot = await get_assistant(chat_id)
                members = []

                try:
                    async for member in userbot.get_call_members(chat_id):
                        if member is None:
                            continue
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
                    except Exception:
                        pass

                    try:
                        language = await get_lang(message.chat.id)
                        language = get_string(language)
                    except:
                        language = get_string("en")
                    try:
                        await app.send_message(
                            chat_id,
                            language["misc_1"],
                        )
                    except Exception:
                        pass

                del autoend[chat_id]


asyncio.create_task(auto_leave(), name="autoleave")
asyncio.create_task(auto_end(), name="autoend")

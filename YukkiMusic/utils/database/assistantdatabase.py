#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import random

from pytgcalls import PyTgCalls

from YukkiMusic import userbot
from YukkiMusic.core.mongo import mongodb

db = mongodb.assistants

assistantdict = {}


async def get_client(assistant: int):
    clients = userbot.clients
    if 1 <= assistant <= len(userbot.clients):
        return clients[assistant - 1]
    return None


async def save_assistant(chat_id, number):
    number = int(number)
    assistantdict[chat_id] = number
    await db.update_one(
        {"chat_id": chat_id},
        {"$set": {"assistant": number}},
        upsert=True,
    )
    return await get_assistant(chat_id)


async def set_assistant(chat_id):
    from YukkiMusic.core.userbot import assistants

    dbassistant = await db.find_one({"chat_id": chat_id})
    current_assistant = dbassistant["assistant"] if dbassistant else None

    available_assistants = [assi for assi in assistants if assi != current_assistant]

    if len(available_assistants) <= 1:
        ran_assistant = random.choice(assistants)
    else:
        ran_assistant = random.choice(available_assistants)

    assistantdict[chat_id] = ran_assistant
    await db.update_one(
        {"chat_id": chat_id},
        {"$set": {"assistant": ran_assistant}},
        upsert=True,
    )

    userbot = await get_client(ran_assistant)
    return userbot


async def get_assistant(chat_id: int) -> str:
    from YukkiMusic.core.userbot import assistants

    assistant = assistantdict.get(chat_id)
    if not assistant:
        dbassistant = await db.find_one({"chat_id": chat_id})
        if not dbassistant:
            userbot = await set_assistant(chat_id)
            return userbot
        else:
            got_assis = dbassistant["assistant"]
            if got_assis in assistants:
                assistantdict[chat_id] = got_assis
                userbot = await get_client(got_assis)
                return userbot
            else:
                userbot = await set_assistant(chat_id)
                return userbot
    else:
        if assistant in assistants:
            userbot = await get_client(assistant)
            return userbot
        else:
            userbot = await set_assistant(chat_id)
            return userbot


async def set_calls_assistant(chat_id):
    from YukkiMusic.core.userbot import assistants

    ran_assistant = random.choice(assistants)
    assistantdict[chat_id] = ran_assistant
    await db.update_one(
        {"chat_id": chat_id},
        {"$set": {"assistant": ran_assistant}},
        upsert=True,
    )
    return ran_assistant


async def group_assistant(self, chat_id: int) -> PyTgCalls:
    from YukkiMusic.core.userbot import assistants

    assistant = assistantdict.get(chat_id)
    if not assistant:
        dbassistant = await db.find_one({"chat_id": chat_id})
        if not dbassistant:
            assis = await set_calls_assistant(chat_id)
        else:
            assis = dbassistant["assistant"]
            if assis in assistants:
                assistantdict[chat_id] = assis
            else:
                assis = await set_calls_assistant(chat_id)
    else:
        if assistant in assistants:
            assis = assistant
        else:
            assis = await set_calls_assistant(chat_id)

    assistant_index = int(assis) - 1

    if 0 <= assistant_index < len(self.calls):
        return self.calls[assistant_index]
    else:
        raise ValueError(f"Assistant index {assistant_index + 1} is out of range.")

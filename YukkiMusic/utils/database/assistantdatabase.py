#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import random as _random

from YukkiMusic.core.mongo import mongodb as _mongodb

_db = _mongodb.assistants

assistantdict = {}


async def get_client(index: int):
    from YukkiMusic import userbot

    clients = userbot.clients
    if 1 <= index <= len(clients):
        return clients[index - 1]  # -1 Because the index start from 1 not from 0
    return None


async def save_assistant(chat_id, index):
    index = int(index)
    assistantdict[chat_id] = index
    await _db.update_one(
        {"chat_id": chat_id},
        {"$set": {"assistant": index}},
        upsert=True,
    )
    return await get_assistant(chat_id)


async def set_assistant(chat_id):
    from YukkiMusic.core.userbot import assistants

    dbassistant = await _db.find_one({"chat_id": chat_id})
    current_assistant = dbassistant["assistant"] if dbassistant else None

    available_assistants = [ass for ass in assistants if ass != current_assistant]

    if len(available_assistants) <= 1:
        ran_assistant = _random.choice(assistants)
    else:
        ran_assistant = _random.choice(available_assistants)

    assistantdict[chat_id] = ran_assistant
    await _db.update_one(
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
        dbassistant = await _db.find_one({"chat_id": chat_id})
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

    ran_assistant = _random.choice(assistants)
    assistantdict[chat_id] = ran_assistant
    await _db.update_one(
        {"chat_id": chat_id},
        {"$set": {"assistant": ran_assistant}},
        upsert=True,
    )
    return ran_assistant


async def group_assistant(self, chat_id: int):
    from YukkiMusic.core.userbot import assistants

    assistant = assistantdict.get(chat_id)
    if not assistant:
        dbassistant = await _db.find_one({"chat_id": chat_id})
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

    if 1 <= assis <= len(self.clients):
        return self.clients[assis - 1]
    raise ValueError(f"Assistant index {assis + 1} is out of range.")

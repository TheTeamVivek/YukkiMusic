#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from config import BANNED_USERS
from YukkiMusic import app
from YukkiMusic.utils.database import (
    blacklist_chat,
    blacklisted_chats,
    whitelist_chat,
)
from YukkiMusic.utils.decorators.language import language


@tbot.on_message(flt.command("BLACKLISTCHAT_COMMAND", True) & flt.user(BANNED_USERS))
@language
async def blacklist_chat_func(event, _):
    if len(event.text.split()) != 2:
        return await event.reply(_["black_1"])
    chat_id = int(message.text.strip().split()[1])
    if chat_id in await blacklisted_chats():
        return await event.reply(_["black_2"])
    blacklisted = await blacklist_chat(chat_id)
    if blacklisted:
        await event.reply(_["black_3"])
    else:
        await event.reply("Something wrong happened.")
    try:
        await app.leave_chat(chat_id)
    except Exception:
        pass


@tbot.on_message(flt.command("WHITELISTCHAT_COMMAND", True) & flt.user(BANNED_USERS))
@language
async def white_funciton(event, _):
    if len(event.text.split()) != 2:
        return await event.reply(_["black_4"])
    chat_id = int(event.text.strip().split()[1])
    if chat_id not in await blacklisted_chats():
        return await event.reply(_["black_5"])
    whitelisted = await whitelist_chat(chat_id)
    if whitelisted:
        return await event.reply(_["black_6"])
    await event.reply("Something wrong happened")


@tbot.on_message(flt.command("BLACKLISTEDCHAT_COMMAND", True) & flt.user(BANNED_USERS))
@language
async def all_chats(event, _):
    text = _["black_7"]
    j = 0
    for count, chat_id in enumerate(await blacklisted_chats(), 1):
        try:
            title = (await app.entity(chat_id)).title
        except Exception:
            title = "Private"
        j = 1
        text += f"**{count}. {title}** [`{chat_id}`]\n"
    if j == 0:
        await event.reply(_["black_8"])
    else:
        await event.reply(text)

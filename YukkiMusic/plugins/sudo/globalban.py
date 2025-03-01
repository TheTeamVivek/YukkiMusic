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

from telethon.errors import FloodWaitError
from pyrogram.types import Message

from config import BANNED_USERS
from strings import command
from YukkiMusic import tbot
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils import get_readable_time
from YukkiMusic.utils.database import (
    add_banned_user,
    get_banned_count,
    get_banned_users,
    get_served_chats,
    is_banned_user,
    remove_banned_user,
)
from YukkiMusic.utils.decorators.language import language


@tbot.on_message(flt.command("GBAN_COMMAND", True) & flt.user(SUDOERS))
@language
async def gbanuser(event, _):
    if not event.is_reply:
        if len(event.text.split()) != 2:
            return await event.reply(_["USER_IDENTIFIER_REQUIRED"])
        user = event.text.split(None, 1)[1]
        user = await tbot.get_entity(user)
        user_id = user.id
        mention  = await tbot.create_mention(user)
    else:
        rmsg = await event.get_reply_message()
        user_id = rmsg.sender_id
        mention  = await tbot.create_mention(user_id)
    if user_id == event.sender_id:
        return await event.reply(_["gban_1"])
    elif user_id == tbot.id:
        return await event.reply(_["gban_2"])
    elif user_id in SUDOERS:
        return await event.reply(_["gban_3"])
    is_gbanned = await is_banned_user(user_id)
    if is_gbanned:
        return await event.reply(_["gban_4"].format(mention))
    if user_id not in BANNED_USERS:
        BANNED_USERS.add(user_id)
    served_chats = []
    chats = await get_served_chats()
    for chat in chats:
        served_chats.append(int(chat["chat_id"]))
    time_expected = len(served_chats)
    time_expected = get_readable_time(time_expected)
    mystic = await event.reply(_["gban_5"].format(mention, time_expected))
    number_of_chats = 0
    for chat_id in served_chats:
        try:
            await tbot.edit_permissions(chat_id, user_id, view_messages=False)
            number_of_chats += 1
        except FloodWaitError as e:
            await asyncio.sleep(int(e.seconds))
        except Exception:
            pass
    await add_banned_user(user_id)
    await event.reply(_["gban_6"].format(mention, number_of_chats))
    await mystic.delete()


@tbot.on_message(flt.command("UNGBAN_COMMAND", True) & flt.user(SUDOERS))
@language
async def gungabn(event, _):
    if not event.is_reply:
        if len(event.text.split()) != 2:
            return await event.reply(_["USER_IDENTIFIER_REQUIRED"])
        user = event.text.split(None, 1)[1]
        user = await tbot.get_entity(user)
        user_id = user.id
        mention  = await tbot.create_mention(user)
    else:
        rmsg = await event.get_reply_message()
        user_id = rmsg.sender_id
        mention  = await tbot.create_mention(user_id)
    is_gbanned = await is_banned_user(user_id)
    if not is_gbanned:
        return await event.reply(_["gban_7"].format(mention))
    if user_id in BANNED_USERS:
        BANNED_USERS.remove(user_id)
    served_chats = []
    chats = await get_served_chats()
    for chat in chats:
        served_chats.append(int(chat["chat_id"]))
    time_expected = len(served_chats)
    time_expected = get_readable_time(time_expected)
    mystic = await event.reply(_["gban_8"].format(mention, time_expected))
    number_of_chats = 0
    for chat_id in served_chats:
        try:
            await tbot.edit_permissions(chat_id, user_id, view_messages=True)
            number_of_chats += 1
        except FloodWaitError as e:
            await asyncio.sleep(int(e.seconds))
        except Exception:
            pass
    await remove_banned_user(user_id)
    await event.reply(_["gban_9"].format(mention, number_of_chats))
    await mystic.delete()


@tbot.on_message(flt.command("GBANNED_COMMAND", True) & flt.user(SUDOERS))
@language
async def gbanned_list(event, _):
    counts = await get_banned_count()
    if counts == 0:
        return await event.reply(_["gban_10"])
    mystic = await event.reply(_["gban_11"])
    msg = "Gbanned Users:\n\n"
    count = 0
    users = await get_banned_users()
    for user_id in users:
        count += 1
        try:
            user = await tbot.get_entity(user_id)
            mention  = await tbot.create_mention(user)
            msg += f"{count}➤ {user}\n"
        except Exception:
            msg += f"{count}➤ [Unfetched User]{user_id}\n"
            continue
    if count == 0:
        return await mystic.edit(_["gban_10"])
    else:
        return await mystic.edit(msg)
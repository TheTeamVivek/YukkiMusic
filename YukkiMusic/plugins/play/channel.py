#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import re

from pyrogram import filters
from pyrogram.enums import ChatMembersFilter, ChatMemberStatus, ChatType
from pyrogram.errors import ChatAdminRequired, UserAlreadyParticipant
from pyrogram.types import Message
from config import BANNED_USERS
from strings import command, get_command
from YukkiMusic import app
from YukkiMusic.utils.database import get_lang, set_cmode
from YukkiMusic.utils.decorators.admins import AdminActual


@app.on_message(command("CHANNELPLAY_COMMAND") & filters.group & ~BANNED_USERS)
@AdminActual
async def playmode_(client, message: Message, _):
    try:
        lang_code = await get_lang(message.chat.id)
        CHANNELPLAY_COMMAND = get_command(lang_code)["CHANNELPLAY_COMMAND"]
    except Exception:
        CHANNELPLAY_COMMAND = get_command("en")["CHANNELPLAY_COMMAND"]

    if len(message.command) < 2:
        return await message.reply_text(
            _["cplay_1"].format(message.chat.title, CHANNELPLAY_COMMAND[0])
        )

    raw_query = message.text.split(None, 2)[1].strip()

    if raw_query.lower() == "disable":
        await set_cmode(message.chat.id, None)
        return await message.reply_text("Channel Play Disabled")

    elif raw_query.lower() == "linked":
        chat = await app.get_chat(message.chat.id)
        if chat.linked_chat:
            chat_id = chat.linked_chat.id
            await set_cmode(message.chat.id, chat_id)
            return await message.reply_text(
                _["cplay_3"].format(chat.linked_chat.title, chat.linked_chat.id)
            )
        else:
            return await message.reply_text(_["cplay_2"])

    # Normalize and handle invite link
    query = raw_query.strip()
    if query.startswith("https://t.me/+"):
        try:
            chat = await app.join_chat(query)
        except UserAlreadyParticipant:
            chat = await app.get_chat(query)
        except Exception:
            return await message.reply_text(_["cplay_4"])
    else:
        # Handle @username, username, or t.me/username
        if query.startswith("https://t.me/"):
            query = re.sub(r"(https://t\.me/)", "", query)
        if query.startswith("@"):
            query = query[1:]
        try:
            chat = await app.get_chat(query)
        except Exception:
            return await message.reply_text(_["cplay_4"])

    if chat.type != ChatType.CHANNEL:
        return await message.reply_text(_["cplay_5"])

    try:
        admins = app.get_chat_members(chat.id, filter=ChatMembersFilter.ADMINISTRATORS)
    except Exception:
        return await message.reply_text(_["cplay_4"])

    try:
        async for users in admins:
            if users.status == ChatMemberStatus.OWNER:
                creatorusername = users.user.username
                creatorid = users.user.id
    except ChatAdminRequired:
        return await message.reply_text(_["cplay_4"])

    if creatorid != message.from_user.id:
        return await message.reply_text(
            _["cplay_6"].format(chat.title, f"@{creatorusername}" if creatorusername else creatorid)
        )

    await set_cmode(message.chat.id, chat.id)
    return await message.reply_text(_["cplay_3"].format(chat.title, chat.id))

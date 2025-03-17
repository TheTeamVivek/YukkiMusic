#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from telethon.tl.functions.channels import GetFullChannelRequest
from telethon.tl.types import Channel

from config import BANNED_USERS
from YukkiMusic import tbot, utils
from YukkiMusic.utils.database import set_cmode
from YukkiMusic.utils.decorators.admins import admin_actual


@tbot.on_message(
    flt.command("CHANNELPLAY_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS)
)
@admin_actual
async def playmode_(language, _):
    chat = await event.get_chat()
    if len(event.text.split()) < 2:
        comm = _["CHANNELPLAY_COMMAND"][0]
        return await event.reply(
            _["cplay_1"].format(chat.title, comm)
        )
    query = event.text.split(None, 2)[1].lower().strip()
    if query == "disable":
        await set_cmode(event.chat_id, None)
        return await event.reply("Channel Play Disabled")

    elif query == "linked":
        chat = (await tbot(GetFullChannelRequest(channel=chat.id))).full_chat
        if chat.linked_chat_id:
            chat_id = chat.linked_chat_id
            chat_id = int(f"-100{chat_id}")
            linked_chat = await tbot.get_entity(chat_id)
            await set_cmode(event.chat_id, chat_id)
            return await event.reply(
                _["cplay_3"].format(linked_chat.title, linked_chat.id)
            )
        else:
            return await event.reply(_["cplay_2"])
    else:
        try:
            chat = await tbot.get_entity(query)
        except Exception:
            return await event.reply(_["cplay_4"])
        if not isinstance(chat, Channel) or getattr(chat, "megagroup", False):
            return await event.reply(_["cplay_5"])
        try:
            creator, status = await tbot.get_chat_member(chat.id, event.sender_id)
        except Exception:
            return await event.reply(_["cplay_4"])

        if status != "OWNER":
            return await event.reply(_["cplay_6"].format(chat.title, creator.username))
        await set_cmode(event.chat_id, chat.id)
        return await event.reply(_["cplay_3"].format(chat.title, chat.id))
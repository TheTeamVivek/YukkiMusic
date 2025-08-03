#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from pyrogram import filters
from pyrogram.enums import ChatMembersFilter, ChatMemberStatus, ChatType
from pyrogram.errors import ChatAdminRequired
from pyrogram.types import Message

from config import BANNED_USERS
from strings import command, get_command, pick_commands
from yukkimusic import app
from yukkimusic.utils.database import get_lang, set_cmode
from yukkimusic.utils.decorators.admins import AdminActual

from . import mhelp


@app.on_message(command("CHANNELPLAY_COMMAND") & filters.group & ~BANNED_USERS)
@AdminActual
async def playmode_(client, message: Message, _):
    lang_code = await get_lang(message.chat.id)
    CHANNELPLAY_COMMAND = get_command("CHANNELPLAY_COMMAND", lang_code)
    if len(message.command) < 2:
        return await message.reply_text(
            _["cplay_1"].format(message.chat.title, CHANNELPLAY_COMMAND[0])
        )
    query = message.text.split(None, 2)[1].lower().strip()
    if (str(query)).lower() == "disable":
        await set_cmode(message.chat.id, None)
        return await message.reply_text("Channel Play Disabled")
    elif str(query) == "linked":
        chat = await app.get_chat(message.chat.id)
        if chat.linked_chat:
            chat_id = chat.linked_chat.id
            await set_cmode(message.chat.id, chat_id)
            return await message.reply_text(
                _["cplay_3"].format(chat.linked_chat.title, chat.linked_chat.id)
            )
        else:
            return await message.reply_text(_["cplay_2"])
    else:
        try:
            chat = await app.get_chat(query)
        except Exception:
            return await message.reply_text(_["cplay_4"])
        if chat.type != ChatType.CHANNEL:
            return await message.reply_text(_["cplay_5"])
        try:
            admins = app.get_chat_members(
                chat.id, filter=ChatMembersFilter.ADMINISTRATORS
            )
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
                _["cplay_6"].format(chat.title, creatorusername)
            )
        await set_cmode(message.chat.id, chat.id)
        return await message.reply_text(_["cplay_3"].format(chat.title, chat.id))


(
    mhelp.add(
        "en",
        f"<b>✧ {pick_commands('CHANNELPLAY_COMMAND')}</b> - Connect channel to a group and stream music on channel's voice chat from your group.",
    )
    .add(
        "ar",
        f"<b>✧ {pick_commands('CHANNELPLAY_COMMAND')}</b> - اربط القناة بالمجموعة وشغّل الموسيقى في الدردشة الصوتية للقناة من مجموعتك.",
    )
    .add(
        "as",
        f"<b>✧ {pick_commands('CHANNELPLAY_COMMAND')}</b> - চেনেলক এটা গ্ৰুপৰ সৈতে সংযোগ কৰক আৰু আপোনাৰ গ্ৰুপৰ পৰা চেনেলৰ ভইচ চেটত সংগীত স্ট্ৰিম কৰক।",
    )
    .add(
        "hi",
        f"<b>✧ {pick_commands('CHANNELPLAY_COMMAND')}</b> - चैनल को ग्रुप से जोड़ें और अपने ग्रुप से चैनल की वॉयस चैट पर म्यूजिक स्ट्रीम करें।",
    )
    .add(
        "ku",
        f"<b>✧ {pick_commands('CHANNELPLAY_COMMAND')}</b> - کەناڵەکە بە گرووپێک ببەستێت و میوزیک لە ڕووبارەوەی دەنگی کەناڵەکەی خۆت بڵاو بکەرەوە.",
    )
    .add(
        "tr",
        f"<b>✧ {pick_commands('CHANNELPLAY_COMMAND')}</b> - Kanalı bir gruba bağlayın ve grubunuzdan kanalın sesli sohbetinde müzik yayınlayın.",
    )
)

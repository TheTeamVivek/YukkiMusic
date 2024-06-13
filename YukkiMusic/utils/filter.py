#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from pyrogram import filters

from pyrogram.enums import ChatMemberStatus, ChatType
from pyrogram.types import Message

from YukkiMusic.misc import SUDOERS


async def admin_check(message: Message) -> bool:
    if not message.from_user:
        return False

    if message.chat.type not in [ChatType.SUPERGROUP, ChatType.GROUP]:
        return False

    if message.from_user.id in [
        777000,  # Telegram Service Notifications
        1087968824,  # GroupAnonymousBot
    ]:
        return True

    if message.from_user.id in SUDOERS:
        return True

    client = message._client
    chat_id = message.chat.id
    user_id = message.from_user.id

    check_status = await client.get_chat_member(chat_id=chat_id, user_id=user_id)
    if check_status.status not in [
        ChatMemberStatus.OWNER,
        ChatMemberStatus.ADMINISTRATOR,
    ]:
        return False
    else:
        return True


async def admin_filter_f(filt, client, message):
    return not message.edit_date and await admin_check(message)


admin_filter = filters.create(func=admin_filter_f, name="AdminFilter")

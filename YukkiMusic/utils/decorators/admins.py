#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram.enums import ChatMemberStatus, ChatType
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from config import adminlist
from strings import get_string
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_authuser_names,
    get_cmode,
    get_lang,
    is_active_chat,
    is_commanddelete_on,
    is_maintenance,
    is_nonadmin_chat,
)

from ..formatters import int_to_alpha


def AdminRightsCheck(mystic):
    async def wrapper(client, message):
        if await is_maintenance():
            if message.from_user.id not in SUDOERS:
                return
        if await is_commanddelete_on(message.chat.id):
            try:
                await message.delete()
            except Exception:
                pass
        try:
            language = await get_lang(message.chat.id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")
        if message.sender_chat:
            upl = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            text="How to Fix this? ",
                            callback_data="AnonymousAdmin",
                        ),
                    ]
                ]
            )
            return await message.reply_text(_["general_4"], reply_markup=upl)
        if message.command[0][0] == "c":
            chat_id = await get_cmode(message.chat.id)
            if chat_id is None:
                return await message.reply_text(_["setting_12"])
            try:
                await app.get_chat(chat_id)
            except Exception:
                return await message.reply_text(_["cplay_4"])
        else:
            chat_id = message.chat.id
        if not await is_active_chat(chat_id):
            return await message.reply_text(_["general_6"])
        is_non_admin = await is_nonadmin_chat(message.chat.id)
        if not is_non_admin:
            if message.from_user.id not in SUDOERS:
                admins = adminlist.get(message.chat.id)
                if not admins:
                    return await message.reply_text(_["admin_1"])
                else:
                    if message.from_user.id not in admins:
                        return await message.reply_text(_["admin_2"])
        return await mystic(client, message, _, chat_id)

    return wrapper


def AdminActual(mystic):
    async def wrapper(client, message):
        if await is_maintenance():
            if message.from_user.id not in SUDOERS:
                return

        if await is_commanddelete_on(message.chat.id):
            try:
                await message.delete()
            except Exception:
                pass

        try:
            language = await get_lang(message.chat.id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")

        if message.sender_chat:
            upl = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            text="How to Fix this?",
                            callback_data="AnonymousAdmin",
                        ),
                    ]
                ]
            )
            return await message.reply_text(_["general_4"], reply_markup=upl)

        if message.from_user.id not in SUDOERS:
            try:
                member = await client.get_chat_member(
                    message.chat.id, message.from_user.id
                )

                if member.status not in [
                    ChatMemberStatus.ADMINISTRATOR,
                    ChatMemberStatus.OWNER,
                ] or (
                    member.privileges is None
                    or not member.privileges.can_manage_video_chats
                ):
                    return await message.reply(_["general_5"])

            except Exception as e:
                return await message.reply(f"Error: {str(e)}")

        return await mystic(client, message, _)

    return wrapper


def ActualAdminCB(mystic):
    async def wrapper(client, query):
        try:
            language = await get_lang(query.message.chat.id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")

        if await is_maintenance():
            if query.from_user.id not in SUDOERS:
                return await query.answer(
                    _["maint_4"],
                    show_alert=True,
                )

        if query.message.chat.type == ChatType.PRIVATE:
            return await mystic(client, query, _)

        is_non_admin = await is_nonadmin_chat(query.message.chat.id)
        if not is_non_admin:
            try:
                a = await app.get_chat_member(
                    query.message.chat.id,
                    query.from_user.id,
                )

                if a is None or (
                    a.privileges is None or not a.privileges.can_manage_video_chats
                ):
                    if query.from_user.id not in SUDOERS:
                        token = await int_to_alpha(query.from_user.id)
                        _check = await get_authuser_names(query.from_user.id)
                        if token not in _check:
                            return await query.answer(
                                _["general_5"],
                                show_alert=True,
                            )

            except Exception as e:
                return await query.answer(f"Error: {str(e)}")

        return await mystic(client, query, _)

    return wrapper

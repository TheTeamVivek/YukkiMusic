#
# Copyright (C) 2021-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram.enums import ChatType
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from config import adminlist
from strings import get_string
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_authuser_names, get_cmode, get_lang,
    is_active_chat, is_commanddelete_on, is_maintenance, is_nonadmin_chat
)
from ..formatters import int_to_alpha

async def handle_maintenance(client, user_id):
    if not await is_maintenance() and user_id not in SUDOERS:
        return await client.send_message(user_id, "Bot is under maintenance. Please wait for some time...")

async def handle_command_delete(client, message):
    if await is_commanddelete_on(message.chat.id):
        try:
            await message.delete()
        except:
            pass

async def get_string_or_default(chat_id, lang_code, default_lang="en"):
    try:
        return get_string(await get_lang(chat_id))
    except:
        return get_string(default_lang)

async def check_admin_privileges(client, chat, user):
    try:
        member = await app.get_chat_member(chat.id, user.id)
        return member.privileges.can_manage_video_chats
    except:
        return False

def AdminRightsCheck(mystic):
    async def wrapper(client, message):
        handle_maintenance(client, message.from_user.id)
        handle_command_delete(client, message)

        _ = get_string_or_default(message.chat.id, get_string("en"))

        if message.sender_chat:
            upl = InlineKeyboardMarkup([
                [InlineKeyboardButton(text="How to Fix this?", callback_data="AnonymousAdmin")]
            ])
            return await message.reply_text(_["general_4"], reply_markup=upl)

        if message.command[0][0] == "c":
            chat_id = await get_cmode(message.chat.id)
            if chat_id is None:
                return await message.reply_text(_["setting_12"])
            try:
                await app.get_chat(chat_id)
            except:
                return await message.reply_text(_["cplay_4"])
        else:
            chat_id = message.chat.id

        if not await is_active_chat(chat_id):
            return await message.reply_text(_["general_6"])

        is_non_admin = await is_nonadmin_chat(message.chat.id)
        if not is_non_admin:
            admins = adminlist.get(message.chat.id)
            if not admins or message.from_user.id not in admins:
                return await message.reply_text(_["admin_18"] if not admins else _["admin_19"])

        return await mystic(client, message, _, chat_id)

    return wrapper

def AdminActual(mystic):
    async def wrapper(client, message):
        handle_maintenance(client, message.from_user.id)
        handle_command_delete(client, message)

        _ = get_string_or_default(message.chat.id, get_string("en"))

        if message.sender_chat and (message.from_user.id not in SUDOERS or not check_admin_privileges(client, message.chat, message.from_user)):
            upl = InlineKeyboardMarkup([
                [InlineKeyboardButton(text="How to Fix this?", callback_data="AnonymousAdmin")]
            ])
            return await message.reply_text(_["general_4"], reply_markup=upl)

        return await mystic(client, message, _)

    return wrapper

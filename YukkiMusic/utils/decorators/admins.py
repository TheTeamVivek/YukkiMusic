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

from config import adminlist, SUDOERS
from strings import get_string
from YukkiMusic import app
from YukkiMusic.utils.database import (
    get_authuser_names,
    get_cmode,
    get_lang,
    is_active_chat,
    is_commanddelete_on,
    is_maintenance,
    is_nonadmin_chat,
)
from YukkiMusic.utils.formatters import int_to_alpha

def handle_maintenance_message(message):
    if not is_maintenance() and message.from_user.id not in SUDOERS:
        return message.reply_text("Bot is under maintenance. Please wait for some time...")

def handle_command_deletion(message):
    if is_commanddelete_on(message.chat.id):
        try:
            message.delete()
        except:
            pass

def get_string_by_language(chat_id):
    try:
        language = get_lang(chat_id)
        return get_string(language)
    except:
        return get_string("en")

def handle_sender_chat(message):
    if message.sender_chat:
        upl = InlineKeyboardMarkup([
            [InlineKeyboardButton(text="How to Fix this? ", callback_data="AnonymousAdmin")],
        ])
        return message.reply_text(get_string_by_language(message.chat.id)["general_4"], reply_markup=upl)

def handle_cplay_command(message, chat_id):
    if message.command[0][0] == "c":
        chat_id = get_cmode(message.chat.id)
        if chat_id is None:
            return message.reply_text(get_string_by_language(message.chat.id)["setting_12"])
        try:
            app.get_chat(chat_id)
        except:
            return message.reply_text(get_string_by_language(message.chat.id)["cplay_4"])
    return chat_id

def is_admin(user_id, chat_id):
    is_non_admin = is_nonadmin_chat(chat_id)
    if not is_non_admin:
        if user_id not in SUDOERS:
            admins = adminlist.get(chat_id)
            if not admins:
                return False
            elif user_id not in admins:
                return False
    return True

def AdminRightsCheck(mystic):
    async def wrapper(client, message):
        handle_maintenance_message(message)
        handle_command_deletion(message)
        handle_sender_chat(message)
        chat_id = handle_cplay_command(message, message.chat.id)
        if not is_active_chat(chat_id):
            return message.reply_text(get_string_by_language(message.chat.id)["general_6"])
        if not is_admin(message.from_user.id, chat_id):
            return message.reply_text(get_string_by_language(message.chat.id)["admin_19"])
        return mystic(client, message, get_string_by_language(message.chat.id), chat_id)

    return wrapper

def AdminActual(mystic):
    async def wrapper(client, message):
        handle_maintenance_message(message)
        handle_command_deletion(message)
        handle_sender_chat(message)
        if not is_admin(message.from_user.id, message.chat.id):
            member = await app.get_chat_member(message.chat.id, message.from_user.id)
            if not member.privileges.can_manage_video_chats:
                return message.reply(get_string_by_language(message.chat.id)["general_5"])
        return mystic(client, message, get_string_by_language(message.chat.id))

    return wrapper

def ActualAdminCB(mystic):
    async def wrapper(client, CallbackQuery):
        handle_maintenance_message(CallbackQuery)
        if CallbackQuery.message.chat.type == ChatType.PRIVATE:
            return mystic(client, CallbackQuery, get_string_by_language(CallbackQuery.message.chat.id))
        is_non_admin = is_nonadmin_chat(CallbackQuery.message.chat.id)
        if not is_non_admin:
            a = await app.get_chat_member(CallbackQuery.message.chat.id, CallbackQuery.from_user.id)
            if not a.privileges.can_manage_video_chats and CallbackQuery.from_user.id not in SUDOERS:
                token = await int_to_alpha(CallbackQuery.from_user.id)
                _check = await get_authuser_names(CallbackQuery.from_user.id)
                if token not in _check:
                    try:
                        return CallbackQuery.answer(get_string_by_language(CallbackQuery.message.chat.id)["general_5"], show_alert=True)
                    except:
                        return
        return mystic(client, CallbackQuery, get_string_by_language(CallbackQuery.message.chat.id))

    return wrapper

#
# Copyright (C) 2024 by AnonymousX888@Github, < https://github.com/AnonymousX888 >.
#
# This file is part of < https://github.com/hakutakaid/YukkiMusicFork > project,
# and is released under the MIT License.
# Please see < https://github.com/hakutakaid/YukkiMusicFork/blob/master/LICENSE >
#
# All rights reserved.
#
from pyrogram import filters
from pyrogram.types import Message
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup


from config import BANNED_USERS
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import (
    set_loop,
    delete_filter,
    is_maintenance,
    is_commanddelete_on,
    get_lang,
    is_nonadmin_chat,
    is_active_chat,
    get_cmode,
)

from config import adminlist
from YukkiMusic.plugins import extra_plugins_enabled
from strings import get_string
from YukkiMusic.misc import SUDOERS


@app.on_message(
    filters.command(["stop", "end", "cstop", "cend"]) & filters.group & ~BANNED_USERS
)
async def stop_music(cli, message: Message):
    if await is_maintenance() is False:
        if message.from_user.id not in SUDOERS:
            return await message.reply_text(
                "Bot is under maintenance. Please wait for some time..."
            )
    if not len(message.command) < 2:
        if extra_plugins_enabled:
            if not message.command[0][0] == "c" and not message.command[0][0] == "e":
                filter = " ".join(message.command[1:])
                deleted = await delete_filter(message.chat.id, filter)
                if deleted:
                    return await message.reply_text(f"**ᴅᴇʟᴇᴛᴇᴅ ғɪʟᴛᴇʀ {filter}.**")
                else:
                    return await message.reply_text("**ɴᴏ sᴜᴄʜ ғɪʟᴛᴇʀ.**")

    if await is_commanddelete_on(message.chat.id):
        try:
            await message.delete()
        except:
            pass
    try:
        language = await get_lang(message.chat.id)
        _ = get_string(language)
    except:
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
        except:
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
                return await message.reply_text(_["admin_18"])
            else:
                if message.from_user.id not in admins:
                    return await message.reply_text(_["admin_19"])
    await Yukki.stop_stream(chat_id)
    await set_loop(chat_id, 0)
    await message.reply_text(_["admin_9"].format(message.from_user.mention))

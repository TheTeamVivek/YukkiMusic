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
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from config import BANNED_USERS, EXTRA_PLUGINS, adminlist
from strings import command, get_string, pick_commands
from yukkimusic import app
from yukkimusic.core.call import yukki
from yukkimusic.core.help import ModuleHelp
from yukkimusic.misc import SUDOERS, db
from yukkimusic.utils.database import (
    delete_filter,
    get_cmode,
    get_lang,
    is_active_chat,
    is_commanddelete_on,
    is_maintenance,
    is_nonadmin_chat,
    set_loop,
)


@app.on_message(command("STOP_COMMAND") & filters.group & ~BANNED_USERS)
async def stop_music(cli, message: Message):
    if await is_maintenance():
        if message.from_user.id not in SUDOERS:
            return
    if not len(message.command) < 2:
        if EXTRA_PLUGINS:
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
    try:
        check = db.get(chat_id)
        if check[0].get("mystic"):
            await check[0].get("mystic").delete()
    except Exception:
        pass
    await yukki.stop_stream(chat_id)
    await set_loop(chat_id, 0)
    await message.reply_text(_["stop_1"].format(message.from_user.mention))


(
    ModuleHelp("Admins")
    .name("en", "Admins")
    .add(
        "en",
        f"<b>✧ {pick_commands('STOP_COMMAND', 'en')}</b> - Stop the currently playing music and clear the queue.",
        priority=16,
    )
    .name("ar", "المسؤولين")
    .add(
        "ar",
        f"<b>✧ {pick_commands('STOP_COMMAND', 'ar')}</b> - إيقاف تشغيل الموسيقى الحالية ومسح قائمة الانتظار.",
        priority=16,
    )
    .name("as", "প্ৰশাসক")
    .add(
        "as",
        f"<b>✧ {pick_commands('STOP_COMMAND', 'as')}</b> - বৰ্তমান সংগীত ৰখাওক আৰু কিউ পৰিষ্কাৰ কৰক।",
        priority=16,
    )
    .name("hi", "प्रशासक")
    .add(
        "hi",
        f"<b>✧ {pick_commands('STOP_COMMAND', 'hi')}</b> - वर्तमान में चल रहे संगीत को बंद करें और कतार को साफ़ करें।",
        priority=16,
    )
    .name("ku", "بەڕێوەبەرەکان")
    .add(
        "ku",
        f"<b>✧ {pick_commands('STOP_COMMAND', 'ku')}</b> - گۆرانییە چالاکەکە وەستا و ڕیزبەندییەکە پاک بکەوە.",
        priority=16,
    )
    .name("tr", "Yöneticiler")
    .add(
        "tr",
        f"<b>✧ {pick_commands('STOP_COMMAND', 'tr')}</b> - Çalan müziği durdur ve sırayı temizle.",
        priority=16,
    )
)

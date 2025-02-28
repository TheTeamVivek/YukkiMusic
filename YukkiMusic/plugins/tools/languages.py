#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from pykeyboard import InlineKeyboard
from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, Message

from config import BANNED_USERS
from strings import command, get_string, languages_present
from YukkiMusic import app
from YukkiMusic.utils.database import get_lang, set_lang
from YukkiMusic.utils.decorators import actual_admin_cb, language

# Languages Available


def lanuages_keyboard(_):
    keyboard = InlineKeyboard(row_width=2)
    keyboard.add(
        *[
            (
                InlineKeyboardButton(
                    text=value,
                    callback_data=f"languages:{key}",
                )
            )
            for key, value in languages_present.items()
        ]
    )
    keyboard.row(
        InlineKeyboardButton(
            text=_["BACK_BUTTON"],
            callback_data=f"settingsback_helper",
        ),
        InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data=f"close"),
    )
    return keyboard


@app.on_message(command("LANGUAGE_COMMAND") & filters.group & ~BANNED_USERS)
@language
async def langs_command(client, message: Message, _):
    keyboard = lanuages_keyboard(_)
    await message.reply(
        _["setting_1"].format(message.chat.title, message.chat.id),
        buttons=keyboard,
    )


@app.on_callback_query(filters.regex("LG") & ~BANNED_USERS)
@language
async def lanuagecb(client, CallbackQuery, _):
    try:
        await CallbackQuery.answer()
    except Exception:
        pass
    keyboard = lanuages_keyboard(_)
    return await CallbackQuery.edit_message_reply_markup(buttons=keyboard)


@app.on_callback_query(filters.regex(r"languages:(.*?)") & ~BANNED_USERS)
@actual_admin_cb
async def language_markup(client, CallbackQuery, _):
    langauge = (CallbackQuery.data).split(":")[1]
    old = await get_lang(CallbackQuery.message.chat.id)
    if str(old) == str(langauge):
        return await CallbackQuery.answer(
            "You are already using same language", show_alert=True
        )
    try:
        _ = get_string(langauge)
        await CallbackQuery.answer(
            "Your language changed successfully..", show_alert=True
        )
    except Exception:
        return await CallbackQuery.answer(
            "Failed to change language or language in under Upadte",
            show_alert=True,
        )
    await set_lang(CallbackQuery.message.chat.id, langauge)
    keyboard = lanuages_keyboard(_)
    return await CallbackQuery.edit_message_reply_markup(buttons=keyboard)

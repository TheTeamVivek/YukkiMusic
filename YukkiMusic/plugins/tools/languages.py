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
from YukkiMusic.utils.decorators import ActualAdminCB, language

# Languages Available


def lanuages_keyboard(_):
    keyboard = InlineKeyboard(row_width=2)
    keyboard.add(
        *[
            (
                InlineKeyboardButton(
                    text=languages_present[i],
                    callback_data=f"languages:{i}",
                )
            )
            for i in languages_present
        ]
    )
    keyboard.row(
        InlineKeyboardButton(
            text=_["BACK_BUTTON"],
            callback_data="settingsback_helper",
        ),
        InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
    )
    return keyboard


@app.on_message(command("LANGUAGE_COMMAND") & filters.group & ~BANNED_USERS)
@language
async def langs_command(_, message: Message, lng):
    keyboard = lanuages_keyboard(lng)
    await message.reply_text(
        lng["lang_1"],
        reply_markup=keyboard,
    )


@app.on_callback_query(filters.regex("LG") & ~BANNED_USERS)
@language
async def lanuagecb(_, query, lang):
    await query.answer()
    keyboard = lanuages_keyboard(lang)
    return await query.edit_message_reply_markup(reply_markup=keyboard)


@app.on_callback_query(filters.regex(r"languages:(.*?)") & ~BANNED_USERS)
@ActualAdminCB
async def language_markup(_, query, lang):
    langauge = (query.data).split(":")[1]
    old = await get_lang(query.message.chat.id)
    if str(old) == str(langauge):
        return await query.answer(lang["lang_2"], show_alert=True)
    try:
        lang = get_string(langauge)
        await query.answer(lang["lang_3"], show_alert=True)
    except KeyError:
        return await query.answer(
            lang["lang_4"],
            show_alert=True,
        )
    await set_lang(query.message.chat.id, langauge)
    keyboard = lanuages_keyboard(lang)
    return await query.edit_message_reply_markup(reply_markup=keyboard)

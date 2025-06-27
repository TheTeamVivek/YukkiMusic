#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import logging
import traceback
from math import ceil

from pyrogram import filters, types
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from config import BANNED_USERS, START_IMG_URL
from strings import command, get_string
from yukkimusic import HELPABLE, app
from yukkimusic.utils.database import get_lang, is_commanddelete_on
from yukkimusic.utils.decorators.language import language
from yukkimusic.utils.inline.help import private_help_panel

COLUMN_SIZE = 3  # Number of button height
NUM_COLUMNS = 3  # Number of button width

logger = logging.getLogger(__name__)


async def paginate_modules(page_n, chat_id: int, close: bool = False):
    lang = await get_lang(chat_id)
    string = get_string(lang)

    helper_buttons = [
        InlineKeyboardButton(
            text=helper_key.replace("_HELPER", "").title(),
            callback_data=f"help_helper({helper_key}:{page_n}:{int(close)})",
        )
        for helper_key in sorted(string.keys())
        if helper_key.endswith("_HELPER")
    ]

    module_buttons = [
        InlineKeyboardButton(
            x.__MODULE__,
            callback_data="help_module({}:{}:{})".format(
                x.__MODULE__.lower(), page_n, int(close)
            ),
        )
        for x in sorted(HELPABLE.values(), key=lambda m: m.__MODULE__.lower())
    ]

    all_buttons = helper_buttons + module_buttons
    pairs = [
        all_buttons[i : i + NUM_COLUMNS]
        for i in range(0, len(all_buttons), NUM_COLUMNS)
    ]
    max_num_pages = ceil(len(pairs) / COLUMN_SIZE) if len(pairs) > 0 else 1
    modulo_page = page_n % max_num_pages

    navigation_buttons = [
        InlineKeyboardButton(
            "❮",
            callback_data="help_prev({}:{})".format(
                modulo_page - 1 if modulo_page > 0 else max_num_pages - 1,
                int(close),
            ),
        ),
        InlineKeyboardButton(
            "close" if close else "Back",
            callback_data="close" if close else "settingsback_helper",
        ),
        InlineKeyboardButton(
            "❯",
            callback_data=f"help_next({modulo_page + 1}:{int(close)})",
        ),
    ]

    if len(pairs) > COLUMN_SIZE:
        pairs = pairs[modulo_page * COLUMN_SIZE : COLUMN_SIZE * (modulo_page + 1)] + [
            navigation_buttons
        ]
    else:
        pairs.append(
            [
                InlineKeyboardButton(
                    "close" if close else "Back",
                    callback_data="close" if close else "settingsback_helper",
                )
            ]
        )

    return InlineKeyboardMarkup(pairs)


@app.on_message(command("HELP_COMMAND") & filters.private & ~BANNED_USERS)
@app.on_callback_query(filters.regex("settings_back_helper") & ~BANNED_USERS)
async def helper_private(client: app, update: types.Message | types.CallbackQuery):
    is_callback = isinstance(update, types.CallbackQuery)
    if is_callback:
        try:
            await update.answer()
        except Exception:
            pass
        chat_id = update.message.chat.id
        language = await get_lang(chat_id)
        _ = get_string(language)
        keyboard = await paginate_modules(0, chat_id, close=False)
        await update.edit_message_text(_["help_1"], reply_markup=keyboard)
    else:
        chat_id = update.chat.id
        if await is_commanddelete_on(update.chat.id):
            try:
                await update.delete()
            except Exception:
                pass
        language = await get_lang(chat_id)
        _ = get_string(language)
        keyboard = await paginate_modules(0, chat_id, close=True)
        if START_IMG_URL:
            await update.reply_photo(
                photo=START_IMG_URL,
                caption=_["help_1"],
                reply_markup=keyboard,
            )
        else:
            await update.reply_text(
                text=_["help_1"],
                reply_markup=keyboard,
            )


@app.on_message(command("HELP_COMMAND") & filters.group & ~BANNED_USERS)
@language(no_check=True)
async def help_com_group(client, message: Message, _):
    keyboard = private_help_panel(_)
    await message.reply_text(_["help_2"], reply_markup=InlineKeyboardMarkup(keyboard))


@app.on_callback_query(filters.regex(r"help_(\w+)\(([\w:]+)\)"))
async def help_button(client, query):
    pattern_match = query.matches[0]
    key = pattern_match.group(1).lower()
    language = await get_lang(query.message.chat.id)
    _ = get_string(language)
    top_text = _["help_1"]

    if key == "module":

        module, prev_page_num, close = pattern_match.group(2).split(":")
        prev_page_num = int(prev_page_num)
        close = bool(int(close))
        text = HELPABLE[module].__HELP__
        key = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="↪️ Back",
                        callback_data=f"help_prev({prev_page_num}:{int(close)})",
                    ),
                    InlineKeyboardButton(text="🔄 Close", callback_data="close"),
                ],
            ]
        )
        await query.message.edit(
            text=text,
            reply_markup=key,
        )
    elif key == "prev":
        curr_page, close = pattern_match.group(2).split(":")
        curr_page = int(curr_page)
        close = bool(int(close))
        await query.message.edit(
            text=top_text,
            reply_markup=await paginate_modules(
                curr_page, query.message.chat.id, close=close
            ),
        )
    elif key == "next":
        next_page, close = pattern_match.group(2).split(":")
        next_page = int(next_page)
        close = bool(int(close))
        await query.message.edit(
            text=top_text,
            reply_markup=await paginate_modules(
                next_page, query.message.chat.id, close=close
            ),
        )
    elif key == "helper":
        helper_key, page_n, close = pattern_match.group(2).split(":")
        page_n = int(page_n)
        close = bool(int(close))

        key = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="↪️ Back", callback_data=f"help_prev({page_n}:{int(close)})"
                    ),
                    InlineKeyboardButton(text="🔄 Close", callback_data="close"),
                ]
            ]
        )
        try:
            await query.message.edit(
                text=_[helper_key],
                reply_markup=key,
            )
        except Exception:
            traceback.print_exc()

    await client.answer_callback_query(query.id)

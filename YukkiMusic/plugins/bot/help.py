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
import re
from math import ceil

from pyrogram import filters, types
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from config import BANNED_USERS, START_IMG_URL
from strings import command, get_command, get_string, helpers
from YukkiMusic import HELPABLE, app
from YukkiMusic.utils.database import get_lang, is_commanddelete_on
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.inline.help import private_help_panel

from telethon import events
COLUMN_SIZE = 4  # Number of button height
NUM_COLUMNS = 3  # Number of button width

async def paginate_modules(page_n, chat_id: int, close: bool = False):
    language = await get_lang(chat_id)
    helpers_dict = helpers.get(language, helpers.get("en", {}))

    all_buttons = [
        EqInlineKeyboardButton(
            text=helper_key,
            callback_data=f"help_helper({helper_key},{page_n},{int(close)})",
        )
        for helper_key in helpers_dict
    ] + [
        EqInlineKeyboardButton(
            x.__MODULE__,
            callback_data="help_module({},{},{})".format(
                x.__MODULE__.lower(), page_n, int(close)
            ),
        )
        for x in HELPABLE.values()
    ]

    pairs = [
        all_buttons[i : i + NUM_COLUMNS]
        for i in range(0, len(all_buttons), NUM_COLUMNS)
    ]
    max_num_pages = ceil(len(pairs) / COLUMN_SIZE) if len(pairs) > 0 else 1
    modulo_page = page_n % max_num_pages

    navigation_buttons = [
        EqInlineKeyboardButton(
            "❮",
            callback_data="help_prev({},{})".format(
                modulo_page - 1 if modulo_page > 0 else max_num_pages - 1,
                int(close),
            ),
        ),
        EqInlineKeyboardButton(
            "close" if close else "Back",
            callback_data="close" if close else "settingsback_helper",
        ),
        EqInlineKeyboardButton(
            "❯",
            callback_data=f"help_next({modulo_page + 1},{int(close)})",
        ),
    ]

    if len(pairs) > COLUMN_SIZE:
        pairs = pairs[modulo_page * COLUMN_SIZE : COLUMN_SIZE * (modulo_page + 1)] + [
            navigation_buttons
        ]
    else:
        pairs.append(
            [
                EqInlineKeyboardButton(
                    "close" if close else "Back",
                    callback_data="close" if close else "settingsback_helper",
                )
            ]
        )

    return InlineKeyboardMarkup(pairs)

@tbot.on_message(flt.command("HELP_COMMAND", True) & flt.private & ~flt.user(BANNED_USERS))
@tbot.on(events.CallbackQuery(pattern="settings_back_helper", func = ~flt.user(BANNED_USERS)))
async def helper_private(event):
    is_callback = hasattr(event, "data")
    chat_id = event.chat_id
    language = await get_lang(chat_id)
    _ = get_string(language)
    if is_callback:
        try:
            await event.answer()
        except Exception:
            pass
        keyboard = await paginate_modules(0, chat_id, close=False)
        await event.edit(_["help_1"], buttons=keyboard)
    else:
        if await is_commanddelete_on(chat_id):
            try:
                await event.delete()
            except Exception:
                pass
        keyboard = await paginate_modules(0, chat_id, close=True)
        if START_IMG_URL:
            await event.respond(
                file=START_IMG_URL,
                message=_["help_1"],
                buttons=keyboard,
            )
        else:
            await event.respond(
                message=_["help_1"],
                buttons=keyboard,
            )


@tbot.on_message(flt.command("HELP_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS))
@language(no_check=True)
async def help_com_group(event, _):
    keyboard = private_help_panel(_)
    await event.reply(_["help_2"], buttons=keyboard)


@app.on_callback_query(filters.regex(r"help_(.*?)"))
async def help_button(client, query):
    mod_match = re.match(r"help_module\((.+?),(.+?),(\d+)\)", query.data)
    prev_match = re.match(r"help_prev\((.+?),(\d+)\)", query.data)
    next_match = re.match(r"help_next\((.+?),(\d+)\)", query.data)
    helper_match = re.match(r"help_helper\((.+?),(.+?),(\d+)\)", query.data)

    try:
        language = await get_lang(query.message.chat.id)
        _ = get_string(language)
        helpers_dict = helpers.get(language, helpers.get("en"))

    except Exception:
        _ = get_string("en")
        helpers_dict = helpers.get("en", {})

    top_text = _["help_1"]

    if mod_match:
        module = mod_match.group(1)
        prev_page_num = int(mod_match.group(2))
        close = bool(int(mod_match.group(3)))
        text = (
            f"<b><u>Here is the help for {HELPABLE[module].__MODULE__}:</u></b>\n"
            + HELPABLE[module].__HELP__
        )
        key = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="↪️ Back",
                        callback_data=f"help_prev({prev_page_num},{int(close)})",
                    ),
                    InlineKeyboardButton(text="🔄 Close", callback_data="close"),
                ],
            ]
        )
        await query.message.edit(
            text=text,
            buttons=key,
            link_preview=False,
        )
    elif prev_match:
        curr_page = int(prev_match.group(1))
        close = bool(int(prev_match.group(2)))
        await query.message.edit(
            text=top_text,
            buttons=await paginate_modules(
                curr_page, query.message.chat.id, close=close
            ),
            link_preview=False,
        )
    elif next_match:
        next_page = int(next_match.group(1))
        close = bool(int(next_match.group(2)))
        await query.message.edit(
            text=top_text,
            buttons=await paginate_modules(
                next_page, query.message.chat.id, close=close
            ),
            link_preview=False,
        )
    elif helper_match:
        helper_key = helper_match.group(1)
        page_n = int(helper_match.group(2))
        close = bool(int(helper_match.group(3)))
        raw_text = helpers_dict.get(helper_key, None)
        formatted_text = _["helper_key"]
        key = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="↪️ Back", callback_data=f"help_prev({page_n},{int(close)})"
                    ),
                    InlineKeyboardButton(text="🔄 Close", callback_data="close"),
                ]
            ]
        )
        try:
            await query.message.edit(
                text=f"<b>{helper_key}:</b>\n{formatted_text}",
                buttons=key,
                link_preview=False,
            )
        except Exception as e:
            logging.exception(e)

    await client.answer_callback_query(query.id)

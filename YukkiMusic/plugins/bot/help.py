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

from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from telethon import events, Button

from config import BANNED_USERS, START_IMG_URL
from strings import get_string, helpers
from YukkiMusic import HELPABLE, app
from YukkiMusic.utils.database import get_lang, is_commanddelete_on
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.inline.help import private_help_panel

COLUMN_SIZE = 4  # Number of button height
NUM_COLUMNS = 3  # Number of button width


async def paginate_modules(page_n, 
async def paginate_modules(page_n, chat_id: int, close: bool = False):
    lang = await get_lang(chat_id)
    string = get_string(lang)

    helper_buttons = [
        Button.inline(helper_key.replace("_HELPER", "").lower(), data=f"help_helper:{helper_key}:{page_n}:{int(close)}")
        for helper_key in string.keys()
        if helper_key.endswith("_HELPER")
    ]

    module_buttons = [
        Button.inline(
            x.__MODULE__,
            data=f"help_module:{x.__MODULE__.lower()}:{page_n}:{int(close)}",
        )
        for x in HELPABLE.values()
    ]

    all_buttons = helper_buttons + module_buttons
    pairs = [
        all_buttons[i : i + NUM_COLUMNS]
        for i in range(0, len(all_buttons), NUM_COLUMNS)
    ]
    max_num_pages = ceil(len(pairs) / COLUMN_SIZE) if len(pairs) > 0 else 1
    modulo_page = page_n % max_num_pages

    navigation_buttons = [
        Button.inline("❮", data=f"help_prev:{modulo_page - 1 if modulo_page > 0 else max_num_pages - 1}:{int(close)}"),
        Button.inline(string["CLOSE_BUTTON"] if close else string["CLOSE_BUTTON"], data="close" if close else "settings_back_helper"),
        Button.inline("❯", data=f"help_next:{modulo_page + 1}:{int(close)}"),
    ]

    if len(pairs) > COLUMN_SIZE:
        pairs = pairs[modulo_page * COLUMN_SIZE : COLUMN_SIZE * (modulo_page + 1)] + [
            navigation_buttons
        ]
    else:
        pairs.append([navigation_buttons[1]])

    return pairs


@tbot.on_message(
    flt.command("HELP_COMMAND", True) & flt.private & ~flt.user(BANNED_USERS)
)
@tbot.on(
    events.CallbackQuery(pattern="settings_back_helper", func=~flt.user(BANNED_USERS))
)
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


@tbot.on_message(
    flt.command("HELP_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS)
)
@language(no_check=True)
async def help_com_group(event, _):
    keyboard = private_help_panel(_)
    await event.reply(_["help_2"], buttons=keyboard)


@tbot.on(events.CallbackQuery(pattern=r"^help_(.+)"))
async def help_button(event):
    pattern_match = event.pattern_match.group(1)
    lang = await get_lang(event.chat_id)
    string = get_string(lang)
        
    if pattern_match.startswith("module"):
        _, module, prev_page_num, close = pattern_match.split(":")
        close = bool(int(close))
        text = f"<b><u>Here is the help for {HELPABLE[module].__MODULE__}:</u></b>\n" + HELPABLE[module].__HELP__

        buttons = [[
            Button.inline(string["BACK_BUTTON"], data=f"help_prev:{prev_page_num}:{int(close)}"),
            Button.inline(string["CLOSE_BUTTON"], data="close"),
        ]]

        await event.edit(text, buttons=buttons, link_preview=False)

    elif pattern_match.startswith("prev"):
        _, curr_page, close = pattern_match.split(":")
        close = bool(int(close))
        chat_id = event.chat_id

        keyboard = await paginate_modules(int(curr_page), chat_id, close=close)
        lang = await get_lang(chat_id)
        string = get_string(lang)

        await event.edit(string["help_1"], buttons=keyboard, link_preview=False)

    elif pattern_match.startswith("next"):
        _, next_page, close = pattern_match.split(":")
        close = bool(int(close))
        chat_id = event.chat_id

        keyboard = await paginate_modules(int(next_page), chat_id, close=close)
        lang = await get_lang(chat_id)
        string = get_string(lang)

        await event.edit(string["help_1"], buttons=keyboard, link_preview=False)

    elif pattern_match.startswith("helper"):
        _, helper_key, page_n, close = pattern_match.split(":")
        close = bool(int(close))
        helper_key = helper_key.upper() + "_HELPER"
        text = string.get(helper_key, f"No help available for {helper_key}.")

        buttons = [[
            Button.inline(string["BACK_BUTTON"], data=f"help_prev:{page_n}:{int(close)}"),
            Button.inline(string["CLOSE_BUTTON"], data="close"),
        ]]

        try:
            await event.edit(f"<b>{helper_key}:</b>\n{text}", buttons=buttons, link_preview=False)
        except Exception as e:
            logging.exception(e)

    await event.answer()

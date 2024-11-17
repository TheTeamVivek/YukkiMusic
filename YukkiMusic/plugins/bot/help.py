#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import re
import logging
from math import ceil
from typing import Union

from pyrogram import filters, types
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from config import BANNED_USERS, START_IMG_URL
from strings import get_string, command, helpers, get_command
from YukkiMusic import HELPABLE, app
from YukkiMusic.utils.database import get_lang, is_commanddelete_on
from YukkiMusic.utils.decorators.language import LanguageStart
from YukkiMusic.utils.inline.help import private_help_panel

COLUMN_SIZE = 4  # Number of button height
NUM_COLUMNS = 3  # Number of button width


class EqInlineKeyboardButton(InlineKeyboardButton):
    def __eq__(self, other):
        return self.text == other.text

    def __lt__(self, other):
        return self.text < other.text

    def __gt__(self, other):
        return self.text > other.text


async def format_helper_text(lng, helper_key: str, text: str) -> str:
    if not text:
        return ""
    _ = get_command(lng)

    def _cmd(key):
        return " ".join(f"/{cmd}" for cmd in _[key])

    if helper_key == "Auth":
        return text.format(
            _cmd("AUTH_COMMAND"), _cmd("UNAUTH_COMMAND"), _cmd("AUTHUSERS_COMMAND")
        )
    elif helper_key == "Admin":
        return text.format(
            _cmd("PAUSE_COMMAND"),
            _cmd("RESUME_COMMAND"),
            _cmd("MUTE_COMMAND"),
            _cmd("UNMUTE_COMMAND"),
            _cmd("SKIP_COMMAND"),
            _cmd("STOP_COMMAND"),
            _cmd("SHUFFLE_COMMAND"),
            _cmd("SEEK_COMMAND"),
            _cmd("SEEK_BACK_COMMAND"),
            _cmd("REBOOT_COMMAND"),
            _cmd("LOOP_COMMAND"),
        )
    elif helper_key == "Active":
        return text.format(
            _cmd("ACTIVEVC_COMMAND"),
            _cmd("ACTIVEVIDEO_COMMAND"),
            _cmd("AC_COMMAND"),
            _cmd("STATS_COMMAND"),
        )
    elif helper_key == "Play":
        return text.format(
            _cmd("PLAY_COMMAND"),
            _cmd("PLAYMODE_COMMAND"),
            _cmd("CHANNELPLAY_COMMAND"),
            _cmd("STREAM_COMMAND"),
        )
    elif helper_key == "G-cast":
        return text.format(_cmd("BROADCAST_COMMAND"))
    elif helper_key == "Bot":
        return text.format(
            _cmd("GSTATS_COMMAND"),
            _cmd("SUDOUSERS_COMMAND"),
            _cmd("LYRICS_COMMAND"),
            _cmd("SONG_COMMAND"),
            _cmd("QUEUE_COMMAND"),
            _cmd("AUTHORIZE_COMMAND"),
            _cmd("UNAUTHORIZE_COMMAND"),
            _cmd("AUTHORIZED_COMMAND"),
        )
    elif helper_key == "P-List":
        return text.format(
            _cmd("PLAYLIST_COMMAND"),
            _cmd("DELETE_PLAYLIST_COMMAND"),
            _cmd("PLAY_PLAYLIST_COMMAND"),
            _cmd("PLAY_PLAYLIST_COMMAND"),
        )
    elif helper_key == "B-list":
        return text.format(
            _cmd("BLACKLISTCHAT_COMMAND"),
            _cmd("WHITELISTCHAT_COMMAND"),
            _cmd("BLACKLISTEDCHAT_COMMAND"),
            _cmd("BLOCK_COMMAND"),
            _cmd("UNBLOCK_COMMAND"),
            _cmd("BLOCKED_COMMAND"),
            _cmd("GBAN_COMMAND"),
            _cmd("UNGBAN_COMMAND"),
            _cmd("GBANNED_COMMAND"),
        )
    elif helper_key == "Dev":
        return text.format(
            _cmd("ADDSUDO_COMMAND"),
            _cmd("DELSUDO_COMMAND"),
            _cmd("SUDOUSERS_COMMAND"),
            _cmd("USAGE_COMMAND"),
            _cmd("GETVAR_COMMAND"),
            _cmd("DELVAR_COMMAND"),
            _cmd("SETVAR_COMMAND"),
            _cmd("RESTART_COMMAND"),
            _cmd("UPDATE_COMMAND"),
            _cmd("SPEEDTEST_COMMAND"),
            _cmd("MAINTENANCE_COMMAND"),
            _cmd("LOGGER_COMMAND"),
            _cmd("GETLOG_COMMAND"),
            _cmd("AUTOEND_COMMAND"),
        )
    else:
        return text


async def paginate_modules(page_n, chat_id: int, close: bool = False):
    language = await get_lang(chat_id)
    helpers_dict = helpers.get(language, helpers.get("en", {}))

    helper_buttons = [
        EqInlineKeyboardButton(
            text=helper_key,
            callback_data=f"help_helper({helper_key},{page_n},{int(close)})",
        )
        for helper_key in helpers_dict
    ]

    module_buttons = [
        EqInlineKeyboardButton(
            x.__MODULE__,
            callback_data="help_module({},{},{})".format(
                x.__MODULE__.lower(), page_n, int(close)
            ),
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
            callback_data="help_next({},{})".format(modulo_page + 1, int(close)),
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


@app.on_message(command("HELP_COMMAND") & filters.private & ~BANNED_USERS)
@app.on_callback_query(filters.regex("settings_back_helper") & ~BANNED_USERS)
async def helper_private(
    client: app, update: Union[types.Message, types.CallbackQuery]
):
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
@LanguageStart
async def help_com_group(client, message: Message, _):
    keyboard = private_help_panel(_)
    await message.reply_text(_["help_2"], reply_markup=InlineKeyboardMarkup(keyboard))


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
            reply_markup=key,
            disable_web_page_preview=True,
        )
    elif prev_match:
        curr_page = int(prev_match.group(1))
        close = bool(int(prev_match.group(2)))
        await query.message.edit(
            text=top_text,
            reply_markup=await paginate_modules(
                curr_page, query.message.chat.id, close=close
            ),
            disable_web_page_preview=True,
        )
    elif next_match:
        next_page = int(next_match.group(1))
        close = bool(int(next_match.group(2)))
        await query.message.edit(
            text=top_text,
            reply_markup=await paginate_modules(
                next_page, query.message.chat.id, close=close
            ),
            disable_web_page_preview=True,
        )
    elif helper_match:
        helper_key = helper_match.group(1)
        page_n = int(helper_match.group(2))
        close = bool(int(helper_match.group(3)))
        raw_text = helpers_dict.get(helper_key, None)
        formatted_text = await format_helper_text(language, helper_key, raw_text)
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
                reply_markup=key,
                disable_web_page_preview=True,
            )
        except Exception as e:
            logging.exception(e)

    await client.answer_callback_query(query.id)

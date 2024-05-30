#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

import logging
import random
import re
from typing import Union

from pyrogram import filters, types
from pyrogram.types import InlineKeyboardMarkup, Message

from config import BANNED_USERS, PHOTO, START_IMG_URL
from strings import get_command, get_string, helpers
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import get_lang, is_commanddelete_on
from YukkiMusic.utils.decorators.language import LanguageStart, languageCB
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from YukkiMusic.utils.inline.help import (
    help_back_markup,
    help_mark,
    help_pannel,
    private_help_panel,
)

from config import BANNED_USERS, OWNER_ID
from YukkiMusic.utils.decorators.language import LanguageStart
from YukkiMusic.utils.inline import private_panel
from YukkiMusic.utils.inlinefunction import paginate_modules
from YukkiMusic.__main__ import HELPABLE
### Command
HELP_COMMAND = get_command("HELP_COMMAND")


@app.on_message(filters.command(HELP_COMMAND) & filters.private & ~BANNED_USERS)
@app.on_callback_query(filters.regex("settings_back_helper") & ~BANNED_USERS)
async def helper_private(
    client: app, update: Union[types.Message, types.CallbackQuery]
):
    is_callback = isinstance(update, types.CallbackQuery)
    if is_callback:
        try:
            await update.answer()
        except:
            pass
        chat_id = update.message.chat.id
        language = await get_lang(chat_id)
        _ = get_string(language)
        keyboard = help_mark
        await update.edit_message_text(_["help_1"], reply_markup=keyboard)
    else:
        chat_id = update.chat.id
        if await is_commanddelete_on(update.chat.id):
            try:
                await update.delete()
            except:
                pass
        language = await get_lang(chat_id)
        _ = get_string(language)
        keyboard = help_mark
        if START_IMG_URL:
            await update.reply_photo(
                photo=START_IMG_URL,
                caption=_["help_1"],
                reply_markup=keyboard,
            )

        else:
            await update.reply_photo(
                photo=random.choice(PHOTO),
                caption=_["help_1"],
                reply_markup=keyboard,
            )


@app.on_message(filters.command(HELP_COMMAND) & filters.group & ~BANNED_USERS)
@LanguageStart
async def help_com_group(client, message: Message, _):
    keyboard = private_help_panel(_)
    await message.reply_text(_["help_2"], reply_markup=InlineKeyboardMarkup(keyboard))


@app.on_callback_query(filters.regex("only_music_help") & ~BANNED_USERS)
@languageCB
async def yukki_pages(client, CallbackQuery, _):
    keyboard = help_pannel(_)
    try:
        await CallbackQuery.message.edit_text(_["help_1"], reply_markup=keyboard)
        return
    except:
        return


@app.on_callback_query(filters.regex("helpcallback") & ~BANNED_USERS)
@languageCB
async def helper_cb(client, CallbackQuery, _):
    callback_data = CallbackQuery.data.strip()
    cb = callback_data.split(None, 1)[1]
    keyboard = help_back_markup(_)
    try:
        if cb == "hb5":
            if CallbackQuery.from_user.id not in SUDOERS:
                return await CallbackQuery.answer(
                    "ᴏɴʟʏ ғᴏʀ sᴜᴅᴏ ᴜsᴇʀ's", show_alert=True
                )
            else:
                await CallbackQuery.edit_message_text(
                    helpers.HELP_5, reply_markup=keyboard
                )
                return await CallbackQuery.answer()
        try:
            await CallbackQuery.answer()
        except:
            pass
        if cb == "hb1":
            await CallbackQuery.edit_message_text(helpers.HELP_1, reply_markup=keyboard)
        elif cb == "hb2":
            await CallbackQuery.edit_message_text(helpers.HELP_2, reply_markup=keyboard)
        elif cb == "hb3":
            await CallbackQuery.edit_message_text(helpers.HELP_3, reply_markup=keyboard)
        elif cb == "hb4":
            await CallbackQuery.edit_message_text(helpers.HELP_4, reply_markup=keyboard)
        elif cb == "hb6":
            await CallbackQuery.edit_message_text(helpers.HELP_6, reply_markup=keyboard)

        elif cb == "hb7":
            await CallbackQuery.edit_message_text(helpers.HELP_7, reply_markup=keyboard)
        elif cb == "hb8":
            await CallbackQuery.edit_message_text(helpers.HELP_8, reply_markup=keyboard)
        elif cb == "hb9":
            await CallbackQuery.edit_message_text(helpers.HELP_9, reply_markup=keyboard)
        elif cb == "hb10":
            await CallbackQuery.edit_message_text(
                helpers.HELP_10, reply_markup=keyboard
            )
        elif cb == "hb11":
            await CallbackQuery.edit_message_text(
                helpers.HELP_11, reply_markup=keyboard
            )
        elif cb == "hb12":
            await CallbackQuery.edit_message_text(
                helpers.HELP_12, reply_markup=keyboard
            )
    except Exception as e:
        logging.exception(e)


async def help_parser(name, keyboard=None):
    if not keyboard:
        keyboard = InlineKeyboardMarkup(paginate_modules(0, HELPABLE, "help"))
    return (
        """ʜᴇʟʟᴏ {first_name},

ᴄʟɪᴄᴋ ᴏɴ ʙᴇʟᴏᴡ ʙᴜᴛᴛᴏɴs ғᴏʀ ᴍᴏʀᴇ ɪɴғᴏʀᴍᴀᴛɪᴏɴ.

ᴀʟʟ ᴄᴏᴍᴍᴀɴᴅs sᴛᴀʀᴛs ᴡɪᴛʜ :-  /
""".format(
            first_name=name
        ),
        keyboard,
    )


@app.on_callback_query(filters.regex("shikharbro"))
async def shikhar(_, CallbackQuery):
    text, keyboard = await help_parser(CallbackQuery.from_user.mention)
    await CallbackQuery.message.edit(text, reply_markup=keyboard)


@app.on_callback_query(filters.regex(r"help_(.*?)"))
@LanguageStart
async def help_button(client, query, _):
    home_match = re.match(r"help_home\((.+?)\)", query.data)
    mod_match = re.match(r"help_module\((.+?),(.+?)\)", query.data)
    prev_match = re.match(r"help_prev\((.+?)\)", query.data)
    next_match = re.match(r"help_next\((.+?)\)", query.data)
    back_match = re.match(r"help_back\((\d+)\)", query.data)
    create_match = re.match(r"help_create", query.data)

    top_text = f"""ʜᴇʟʟᴏ {query.from_user.first_name},

ᴄʟɪᴄᴋ ᴏɴ ʙᴇʟᴏᴡ ʙᴜᴛᴛᴏɴs ғᴏʀ ᴍᴏʀᴇ ɪɴғᴏʀᴍᴀᴛɪᴏɴ.

ᴀʟʟ ᴄᴏᴍᴍᴀɴᴅs sᴛᴀʀᴛs ᴡɪᴛʜ :-  /
"""

    if mod_match:
        module = mod_match.group(1)
        prev_page_num = int(mod_match.group(2))
        text = (
            "{} **{}**:\n".format(
                "**ʜᴇʀᴇ ɪs ᴛʜᴇ ʜᴇʟᴘ ғᴏʀ**", HELPABLE[module].__MODULE__
            )
            + HELPABLE[module].__HELP__
        )
        try:
            await app.resolve_peer(OWNER_ID[0])
            OWNER = OWNER_ID[0]
        except:
            OWNER = None
        out = private_panel(_, app.username, OWNER)

        key = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="↪️ Back", callback_data=f"help_back({prev_page_num})"
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

    elif home_match:
        await app.send_message(
            query.from_user.id,
            text=home_text_pm,
            reply_markup=InlineKeyboardMarkup(out),
        )
        await query.message.delete()

    elif prev_match:
        curr_page = int(prev_match.group(1))
        if curr_page < 0:
            curr_page = max_num_pages - 1
        await query.message.edit(
            text=top_text,
            reply_markup=InlineKeyboardMarkup(
                paginate_modules(curr_page, HELPABLE, "help")
            ),
            disable_web_page_preview=True,
        )

    elif next_match:
        next_page = int(next_match.group(1))
        await query.message.edit(
            text=top_text,
            reply_markup=InlineKeyboardMarkup(
                paginate_modules(next_page, HELPABLE, "help")
            ),
            disable_web_page_preview=True,
        )

    elif back_match:
        prev_page_num = int(back_match.group(1))
        await query.message.edit(
            text=top_text,
            reply_markup=InlineKeyboardMarkup(
                paginate_modules(prev_page_num, HELPABLE, "help")
            ),
            disable_web_page_preview=True,
        )

    elif create_match:
        text, keyboard = await help_parser(query)
        await query.message.edit(
            text=text,
            reply_markup=keyboard,
            disable_web_page_preview=True,
        )

    return await client.answer_callback_query(query.id)

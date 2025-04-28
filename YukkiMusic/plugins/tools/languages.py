#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/The TeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from pykeyboard.telethon import InlineKeyboard
from telethon import Button, events

from strings import get_string, languages_present
from YukkiMusic import tbot
from YukkiMusic.core import filters
from YukkiMusic.misc import BANNED_USERS
from YukkiMusic.utils import actual_admin_cb, get_lang, language, set_lang

# Languages Available


def lanuages_keyboard(_):
    keyboard = InlineKeyboard(row_width=2)
    keyboard.add(
        *[
            (
                Button.inline(
                    text=value,
                    data=f"languages:{key}",
                )
            )
            for key, value in languages_present.items()
        ]
    )
    keyboard.row(
        Button.inline(
            text=_["BACK_BUTTON"],
            data=f"settingsback_helper",
        ),
        Button.inline(text=_["CLOSE_BUTTON"], callback_data=f"close"),
    )
    return keyboard


@tbot.on_message(
    filters.command("LANGUAGE_COMMAND", True) & filters.group & ~BANNED_USERS
)
@language
async def langs_command(event, _):
    keyboard = lanuages_keyboard(_)
    chat = await event.get_chat()
    await event.reply(
        _["setting_1"].format(chat.title, event.chat_id),
        buttons=keyboard,
    )


@tbot.on(events.CallbackQuery("LG", func=~BANNED_USERS))
@language
async def lanuagecb(event, _):
    try:
        await event.answer()
    except Exception:
        pass
    keyboard = lanuages_keyboard(_)
    return await event.edit(buttons=keyboard)


@tbot.on(events.CallbackQuery(r"languages:(.*?)", func=~BANNED_USERS))
@actual_admin_cb
async def language_markup(event, _):
    langauge = event.data.decode("utf-8").split(":")[1]
    old = await get_lang(event.chat_id)
    if str(old) == str(langauge):
        return await event.answer(_["lang_1"], alert=True)
    try:
        _ = get_string(langauge)
        await event.answer(_["lang_2"], alert=True)
    except Exception:
        return await event.answer(
            _["lang_3"],
            alert=True,
        )
    await set_lang(event.chat_id, langauge)
    keyboard = lanuages_keyboard(_)
    return await event.edit(buttons=keyboard)

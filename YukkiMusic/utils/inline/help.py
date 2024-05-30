#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#


from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from YukkiMusic import app


def help_back_markup(_):
    upl = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    text=_["BACK_BUTTON"], callback_data=f"settings_back_helper"
                ),
                InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data=f"close"),
            ]
        ]
    )
    return upl


def private_help_panel(_):
    buttons = [
        [
            InlineKeyboardButton(
                text=_["S_B_1"], url=f"https://t.me/{app.username}?start=help"
            )
        ],
    ]
    return buttons

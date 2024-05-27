#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from typing import Union

from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from YukkiMusic import app


def first_page(_):
    firsts_page = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(text=_["H_B_1"], callback_data="helpcallback hb1"),
                InlineKeyboardButton(text=_["H_B_9"], callback_data="helpcallback hb8"),
                InlineKeyboardButton(text=_["H_B_8"], callback_data="helpcallback hb7"),
            ],
            [
                InlineKeyboardButton(text=_["H_B_3"], callback_data="helpcallback hb3"),
                InlineKeyboardButton(
                    text=_["H_B_10"], callback_data="helpcallback hb9"
                ),
                InlineKeyboardButton(text=_["H_B_7"], callback_data="helpcallback hb6"),
            ],
            [
                InlineKeyboardButton(text=_["H_B_6"], callback_data="helpcallback hb5"),
                InlineKeyboardButton(text=_["H_B_4"], callback_data="helpcallback hb4"),
                InlineKeyboardButton(text=_["H_B_2"], callback_data="helpcallback hb2"),
            ],
            [
                InlineKeyboardButton(
                    text=_["H_B_11"], callback_data="helpcallback hb10"
                ),
                InlineKeyboardButton(
                    text=_["H_B_13"], callback_data="helpcallback hb12"
                ),
            ],
            [InlineKeyboardButton(text=_["CLOSEMENU_BUTTON"], callback_data=f"close")],
        ]
    )
    return firsts_page


def help_pannel(_, START: Union[bool, int] = None):
    mark = [InlineKeyboardButton(text=_["CLOSEMENU_BUTTON"], callback_data=f"close")]

    upl = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(text=_["H_B_1"], callback_data="helpcallback hb1"),
                InlineKeyboardButton(text=_["H_B_9"], callback_data="helpcallback hb8"),
                InlineKeyboardButton(text=_["H_B_8"], callback_data="helpcallback hb7"),
            ],
            [
                InlineKeyboardButton(text=_["H_B_3"], callback_data="helpcallback hb3"),
                InlineKeyboardButton(
                    text=_["H_B_10"], callback_data="helpcallback hb9"
                ),
                InlineKeyboardButton(text=_["H_B_7"], callback_data="helpcallback hb6"),
            ],
            [
                InlineKeyboardButton(text=_["H_B_6"], callback_data="helpcallback hb5"),
                InlineKeyboardButton(text=_["H_B_4"], callback_data="helpcallback hb4"),
                InlineKeyboardButton(text=_["H_B_2"], callback_data="helpcallback hb2"),
            ],
            [
                InlineKeyboardButton(
                    text=_["H_B_11"], callback_data="helpcallback hb10"
                ),
            ],
            mark,
        ]
    )
    return upl


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


help_mark = InlineKeyboardMarkup(
    [
        [InlineKeyboardButton(text="Mᴜsɪᴄ Cᴏᴍᴍᴀɴᴅs", callback_data="only_music_help")],
        [InlineKeyboardButton(text="Aʟʟ Cᴏᴍᴍᴀɴᴅs", callback_data="shikharbro")],
        [InlineKeyboardButton(text="〆 ᴄʟᴏsᴇ 〆", callback_data="close")],
    ]
)

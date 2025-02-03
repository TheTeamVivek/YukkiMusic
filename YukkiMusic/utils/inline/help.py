#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from telethon import Button

from config import SUPPORT_GROUP
from YukkiMusic import tbot


def support_group_markup(_):
    upl =   [
            [
                Button.url(
                    text=_["S_B_3"],
                    url=SUPPORT_GROUP,
                ),
            ]
        ]
    return upl


def help_back_markup(_):
    upl =  [
            [
                Button.inline(
                    text=_["BACK_BUTTON"], data=f"settings_back_helper"
                ),
                Button.inline(text=_["CLOSE_BUTTON"], data=f"close"),
            ]
        ]
    return upl


def private_help_panel(_):
    buttons = [
        [
            Button.url(
                text=_["S_B_1"], url=f"https://t.me/{tbot.username}?start=help"
            )
        ],
    ]
    return buttons
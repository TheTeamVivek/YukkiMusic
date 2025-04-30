#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from telethon import Button as _Button
from telethon.tl import types as _types

import config as _config
from YukkiMusic import tbot as _tbot


def start_pannel(_):
    buttons = [
        [
            _Button.url(
                text=_["S_B_1"],
                url=f"https://t.me/{_tbot.username}?start=help",
            ),
            _Button.inline(text=_["S_B_2"], data="settings_helper"),
        ],
    ]
    if _config.SUPPORT_CHANNEL and _config.SUPPORT_GROUP:
        buttons.append(
            [
                _Button.url(text=_["S_B_4"], url=_config.SUPPORT_CHANNEL),
                _Button.url(text=_["S_B_3"], url=_config.SUPPORT_GROUP),
            ]
        )
    else:
        if _config.SUPPORT_CHANNEL:
            buttons.append([_Button.url(text=_["S_B_4"], url=_config.SUPPORT_CHANNEL)])
        if _config.SUPPORT_GROUP:
            buttons.append([_Button.url(text=_["S_B_3"], url=_config.SUPPORT_GROUP)])
    return buttons


def private_panel(_, owner: bool | int = None):
    buttons = [[_Button.inline(text=_["S_B_8"], data="settings_back_helper")]]
    if _config.SUPPORT_CHANNEL and _config.SUPPORT_GROUP:
        buttons.append(
            [
                _Button.url(text=_["S_B_4"], url=_config.SUPPORT_CHANNEL),
                _Button.url(text=_["S_B_3"], url=_config.SUPPORT_GROUP),
            ]
        )
    else:
        if _config.SUPPORT_CHANNEL:
            buttons.append([_Button.url(text=_["S_B_4"], url=_config.SUPPORT_CHANNEL)])
        if _config.SUPPORT_GROUP:
            buttons.append([_Button.url(text=_["S_B_3"], url=_config.SUPPORT_GROUP)])
    buttons.append(
        [
            _Button.url(
                text=_["S_B_5"],
                url=f"https://t.me/{_tbot.username}?startgroup=true",
            )
        ]
    )
    if _config.GITHUB_REPO and owner:
        buttons.append(
            [
                _types.InputKeyboardButtonUserProfile(
                    text=_["S_B_7"], user_id=owner
                ),
                _Button.url(text=_["S_B_6"], url=_config.GITHUB_REPO),
            ]
        )
    else:
        if _config.GITHUB_REPO:
            buttons.append(
                [
                    _Button.url(text=_["S_B_6"], url=_config.GITHUB_REPO),
                ]
            )

        if owner:
            buttons.append(
                [
                    _types.InputKeyboardButtonUserProfile(
                        text=_["S_B_7"], user_id=owner
                    ),
                ]
            )
    buttons.append([_Button.inline(text=_["ST_B_6"], data="LG")])
    return buttons

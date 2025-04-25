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

import config
from YukkiMusic import tbot


def start_pannel(_):
    buttons = [
        [
            Button.url(
                text=_["S_B_1"],
                url=f"https://t.me/{tbot.username}?start=help",
            ),
            Button.inline(text=_["S_B_2"], data="settings_helper"),
        ],
    ]
    if config.SUPPORT_CHANNEL and config.SUPPORT_GROUP:
        buttons.append(
            [
                Button.url(text=_["S_B_4"], url=config.SUPPORT_CHANNEL),
                Button.url(text=_["S_B_3"], url=config.SUPPORT_GROUP),
            ]
        )
    else:
        if config.SUPPORT_CHANNEL:
            buttons.append([Button.url(text=_["S_B_4"], url=config.SUPPORT_CHANNEL)])
        if config.SUPPORT_GROUP:
            buttons.append([Button.url(text=_["S_B_3"], url=config.SUPPORT_GROUP)])
    return buttons


def private_panel(_, owner: bool | int = None):
    buttons = [[Button.inline(text=_["S_B_8"], data="settings_back_helper")]]
    if config.SUPPORT_CHANNEL and config.SUPPORT_GROUP:
        buttons.append(
            [
                Button.url(text=_["S_B_4"], url=config.SUPPORT_CHANNEL),
                Button.url(text=_["S_B_3"], url=config.SUPPORT_GROUP),
            ]
        )
    else:
        if config.SUPPORT_CHANNEL:
            buttons.append([Button.url(text=_["S_B_4"], url=config.SUPPORT_CHANNEL)])
        if config.SUPPORT_GROUP:
            buttons.append([Button.url(text=_["S_B_3"], url=config.SUPPORT_GROUP)])
    buttons.append(
        [
            Button.url(
                text=_["S_B_5"],
                url=f"https://t.me/{tbot.username}?startgroup=true",
            )
        ]
    )
    if config.GITHUB_REPO and owner:
        buttons.append(
            [
                Button.url(text=_["S_B_7"], url=f"tg://user?id={owner}"),
                Button.url(text=_["S_B_6"], url=config.GITHUB_REPO),
            ]
        )
    else:
        if config.GITHUB_REPO:
            buttons.append(
                [
                    Button.url(text=_["S_B_6"], url=config.GITHUB_REPO),
                ]
            )

        if owner:
            buttons.append(
                [
                    Button.url(text=_["S_B_7"], url=f"tg://user?id={owner}"),
                ]
            )
    buttons.append([Button.inline(text=_["ST_B_6"], data="LG")])
    return buttons

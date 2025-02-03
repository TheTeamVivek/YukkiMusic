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

from config import GITHUB_REPO, SUPPORT_CHANNEL, SUPPORT_GROUP
from YukkiMusic import app


def start_pannel(_):
    buttons = [
        [
            Button.url(
                text=_["S_B_1"],
                url=f"https://t.me/{app.username}?start=help",
            ),
            Button.inline(text=_["S_B_2"], data="settings_helper"),
        ],
    ]
    if SUPPORT_CHANNEL and SUPPORT_GROUP:
        buttons.append(
            [
                Button.url(text=_["S_B_4"], url=f"{SUPPORT_CHANNEL}"),
                Button.url(text=_["S_B_3"], url=f"{SUPPORT_GROUP}"),
            ]
        )
    else:
        if SUPPORT_CHANNEL:
            buttons.append([Button.url(text=_["S_B_4"], url=f"{SUPPORT_CHANNEL}")])
        if SUPPORT_GROUP:
            buttons.append([Button.url(text=_["S_B_3"], url=f"{SUPPORT_GROUP}")])
    return buttons


def private_panel(_, owner: bool | int = None):
    buttons = [[Button.inline(text=_["S_B_8"], data="settings_back_helper")]]
    if SUPPORT_CHANNEL and SUPPORT_GROUP:
        buttons.append(
            [
                Button.url(text=_["S_B_4"], url=f"{SUPPORT_CHANNEL}"),
                Button.url(text=_["S_B_3"], url=f"{SUPPORT_GROUP}"),
            ]
        )
    else:
        if SUPPORT_CHANNEL:
            buttons.append([Button.url(text=_["S_B_4"], url=f"{SUPPORT_CHANNEL}")])
        if SUPPORT_GROUP:
            buttons.append([Button.url(text=_["S_B_3"], url=f"{SUPPORT_GROUP}")])
    buttons.append(
        [
            Button.url(
                text=_["S_B_5"],
                url=f"https://t.me/{tbot.username}?startgroup=true",
            )
        ]
    )
    if GITHUB_REPO and owner:
        buttons.append(
            [
                Button.url(text=_["S_B_7"], url=f"tg://user?id={owner}"),
                Button.url(text=_["S_B_6"], url=f"{GITHUB_REPO}"),
            ]
        )
    else:
        if GITHUB_REPO:
            buttons.append(
                [
                    Button.url(text=_["S_B_6"], url=f"{GITHUB_REPO}"),
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

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


def song_markup(_, vidid):
    buttons = [
        [
            Button.inline(
                text=_["SG_B_2"],
                data=f"song_helper audio|{vidid}",
            ),
            Button.inline(
                text=_["SG_B_3"],
                data=f"song_helper video|{vidid}",
            ),
        ],
        [
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
        ],
    ]
    return buttons

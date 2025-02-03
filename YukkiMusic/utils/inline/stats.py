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

from YukkiMusic import tbot


def back_stats_markup(_):
    upl = [
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="TOPMARKUPGET",
            ),
            Button.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def overallback_stats_markup(_):
    upl = [
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="GlobalStats",
            ),
            Button.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def get_stats_markup(_, status):
    not_sudo = [
        Button.inline(
            text=_["CLOSEMENU_BUTTON"],
            data="close",
        )
    ]
    sudo = [
        Button.inline(
            text=_["SA_B_8"],
            data="bot_stats_sudo g",
        ),
        Button.inline(
            text=_["CLOSEMENU_BUTTON"],
            data="close",
        ),
    ]
    upl = [
        [
            Button.inline(
                text=_["SA_B_7"],
                data="TOPMARKUPGET",
            )
        ],
        [
            Button.inline(
                text=_["SA_B_6"],
                url=f"https://t.me/{tbot.username}?start=stats",
            ),
            Button.inline(
                text=_["SA_B_5"],
                data="TopOverall g",
            ),
        ],
        sudo if status else not_sudo,
    ]
    return upl


def stats_buttons(_, status):
    not_sudo = [
        Button.inline(
            text=_["SA_B_5"],
            data="TopOverall s",
        )
    ]
    sudo = [
        Button.inline(
            text=_["SA_B_8"],
            data="bot_stats_sudo s",
        ),
        Button.inline(
            text=_["SA_B_5"],
            data="TopOverall s",
        ),
    ]
    upl = [
        sudo if status else not_sudo,
        [
            Button.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def back_stats_buttons(_):
    upl = [
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="GETSTATS",
            ),
            Button.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def top_ten_stats_markup(_):
    upl = [
        [
            Button.inline(
                text=_["SA_B_2"],
                data="GetStatsNow Tracks",
            ),
            Button.inline(
                text=_["SA_B_1"],
                data="GetStatsNow Chats",
            ),
        ],
        [
            Button.inline(
                text=_["SA_B_3"],
                data="GetStatsNow Users",
            ),
            Button.inline(
                text=_["SA_B_4"],
                data="GetStatsNow Here",
            ),
        ],
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="GlobalStats",
            ),
            Button.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl

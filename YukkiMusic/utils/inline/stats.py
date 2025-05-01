#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from telethon import Button as _btn

from YukkiMusic import tbot as _tbot


def back_stats_markup(_):
    upl = [
        [
            _btn.inline(
                text=_["BACK_BUTTON"],
                data="TOPMARKUPGET",
            ),
            _btn.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def overallback_stats_markup(_):
    upl = [
        [
            _btn.inline(
                text=_["BACK_BUTTON"],
                data="GlobalStats",
            ),
            _btn.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def get_stats_markup(_, status):
    not_sudo = [
        _btn.inline(
            text=_["CLOSEMENU_BUTTON"],
            data="close",
        )
    ]
    sudo = [
        _btn.inline(
            text=_["SA_B_8"],
            data="bot_stats_sudo g",
        ),
        _btn.inline(
            text=_["CLOSEMENU_BUTTON"],
            data="close",
        ),
    ]
    upl = [
        [
            _btn.inline(
                text=_["SA_B_7"],
                data="TOPMARKUPGET",
            )
        ],
        [
            _btn.url(
                text=_["SA_B_6"],
                url=f"https://t.me/{_tbot.username}?start=stats",
            ),
            _btn.inline(
                text=_["SA_B_5"],
                data="TopOverall g",
            ),
        ],
        sudo if status else not_sudo,
    ]
    return upl


def stats_buttons(_, status):
    not_sudo = [
        _btn.inline(
            text=_["SA_B_5"],
            data="TopOverall s",
        )
    ]
    sudo = [
        _btn.inline(
            text=_["SA_B_8"],
            data="bot_stats_sudo s",
        ),
        _btn.inline(
            text=_["SA_B_5"],
            data="TopOverall s",
        ),
    ]
    upl = [
        sudo if status else not_sudo,
        [
            _btn.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def back_stats_buttons(_):
    upl = [
        [
            _btn.inline(
                text=_["BACK_BUTTON"],
                data="GETSTATS",
            ),
            _btn.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl


def top_ten_stats_markup(_):
    upl = [
        [
            _btn.inline(
                text=_["SA_B_2"],
                data="GetStatsNow Tracks",
            ),
            _btn.inline(
                text=_["SA_B_1"],
                data="GetStatsNow Chats",
            ),
        ],
        [
            _btn.inline(
                text=_["SA_B_3"],
                data="GetStatsNow Users",
            ),
            _btn.inline(
                text=_["SA_B_4"],
                data="GetStatsNow Here",
            ),
        ],
        [
            _btn.inline(
                text=_["BACK_BUTTON"],
                data="GlobalStats",
            ),
            _btn.inline(
                text=_["CLOSE_BUTTON"],
                data="close",
            ),
        ],
    ]
    return upl

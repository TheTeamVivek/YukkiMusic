#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import math

from telethon import Button

from YukkiMusic.core.enum import SourceType
from YukkiMusic.core.youtube import Track
from YukkiMusic.utils.formatters import time_to_seconds


def get_progress_bar(percentage):
    umm = math.floor(percentage)

    if 0 < umm <= 10:
        return "▰▱▱▱▱▱▱▱▱"
    elif 10 < umm <= 20:
        return "▰▰▱▱▱▱▱▱▱"
    elif 20 < umm <= 30:
        return "▰▰▰▱▱▱▱▱▱"
    elif 30 < umm <= 40:
        return "▰▰▰▰▱▱▱▱▱"
    elif 40 < umm <= 50:
        return "▰▰▰▰▰▱▱▱▱"
    elif 50 < umm <= 60:
        return "▰▰▰▰▰▰▱▱▱"
    elif 60 < umm <= 70:
        return "▰▰▰▰▰▰▰▱▱"
    elif 70 < umm <= 80:
        return "▰▰▰▰▰▰▰▰▱"
    elif 80 < umm <= 90:
        return "▰▰▰▰▰▰▰▰▰"
    elif 90 < umm <= 100:
        return "▰▰▰▰▰▰▰▰▰"
    else:
        return "▱▱▱▱▱▱▱▱▱"


def stream_markup_timer(_, videoid, chat_id, played, dur):
    played_sec = time_to_seconds(played)
    duration_sec = time_to_seconds(dur)
    percentage = (played_sec / duration_sec) * 100

    status_bar = get_progress_bar(percentage)  # using for getting the bar

    buttons = [
        [
            Button.inline(
                text=f"{played} {status_bar} {dur}",
                data="GetTimer",
            )
        ],
        [
            Button.inline(text=_["P_B_7"], data=f"add_playlist {videoid}"),
            Button.inline(
                text=_["PL_B_3"],
                data=f"PanelMarkup {videoid}|{chat_id}",
            ),
        ],
        [
            Button.inline(text="▷", data=f"ADMIN Resume|{chat_id}"),
            Button.inline(text="II", data=f"ADMIN Pause|{chat_id}"),
            Button.inline(text="‣‣I", data=f"ADMIN Skip|{chat_id}"),
            Button.inline(text="▢", data=f"ADMIN Stop|{chat_id}"),
        ],
        [Button.inline(text=_["CLOSEMENU_BUTTON"], data="close")],
    ]
    return buttons


def stream_markup(_, videoid, chat_id):
    buttons = [
        [
            Button.inline(text=_["P_B_7"], data=f"add_playlist {videoid}"),
            Button.inline(
                text=_["PL_B_3"],
                data=f"PanelMarkup None|{chat_id}",
            ),
        ],
        [
            Button.inline(text="▷", data=f"ADMIN Resume|{chat_id}"),
            Button.inline(text="II", data=f"ADMIN Pause|{chat_id}"),
            Button.inline(text="‣‣I", data=f"ADMIN Skip|{chat_id}"),
            Button.inline(text="▢", data=f"ADMIN Stop|{chat_id}"),
        ],
        [Button.inline(text=_["CLOSEMENU_BUTTON"], data="close")],
    ]
    return buttons


def telegram_markup_timer(_, chat_id, played, dur):
    played_sec = time_to_seconds(played)
    duration_sec = time_to_seconds(dur)
    percentage = (played_sec / duration_sec) * 100

    status_bar = get_progress_bar(percentage)  # using for getting the bar

    buttons = [
        [
            Button.inline(
                text=f"{played} {status_bar} {dur}",
                data="GetTimer",
            )
        ],
        [
            Button.inline(
                text=_["PL_B_3"],
                data=f"PanelMarkup None|{chat_id}",
            ),
        ],
        [
            Button.inline(text="▷", data=f"ADMIN Resume|{chat_id}"),
            Button.inline(text="II", data=f"ADMIN Pause|{chat_id}"),
            Button.inline(text="‣‣I", data=f"ADMIN Skip|{chat_id}"),
            Button.inline(text="▢", data=f"ADMIN Stop|{chat_id}"),
        ],
        [
            Button.inline(text=_["CLOSEMENU_BUTTON"], data="close"),
        ],
    ]
    return buttons


def telegram_markup(_, chat_id):
    buttons = [
        [
            Button.inline(
                text=_["PL_B_3"],
                data=f"PanelMarkup None|{chat_id}",
            ),
        ],
        [
            Button.inline(text="▷", data=f"ADMIN Resume|{chat_id}"),
            Button.inline(text="II", data=f"ADMIN Pause|{chat_id}"),
            Button.inline(text="‣‣I", data=f"ADMIN Skip|{chat_id}"),
            Button.inline(text="▢", data=f"ADMIN Stop|{chat_id}"),
        ],
        [
            Button.inline(text=_["CLOSEMENU_BUTTON"], data="close"),
        ],
    ]
    return buttons


def play_markup(language: dict, chat_id: int, track: Track | None = None):
    if track.vidid is not None and track.streamtype in [
        SourceType.APPLE,
        SourceType.RESSO,
        SourceType.SPOTIFY,
        SourceType.YOUTUBE,
    ] and not (track.is_live or track.is_m3u8):
        return "stream", stream_markup(language, videoid=track.vidid, chat_id=chat_id)
    else:
        return "tg", telegram_markup(language, chat_id)


## Search Query Inline


def track_markup(_, videoid, user_id, channel, fplay):
    buttons = [
        [
            Button.inline(
                text=_["P_B_1"],
                data=f"MusicStream {videoid}|{user_id}|a|{channel}|{fplay}",
            ),
            Button.inline(
                text=_["P_B_2"],
                data=f"MusicStream {videoid}|{user_id}|v|{channel}|{fplay}",
            ),
        ],
        [Button.inline(text=_["CLOSE_BUTTON"], data=f"forceclose {videoid}|{user_id}")],
    ]
    return buttons


def playlist_markup(_, videoid, user_id, ptype, channel, fplay):
    buttons = [
        [
            Button.inline(
                text=_["P_B_1"],
                data=f"YukkiPlaylists {videoid}|{user_id}|{ptype}|a|{channel}|{fplay}",
            ),
            Button.inline(
                text=_["P_B_2"],
                data=f"YukkiPlaylists {videoid}|{user_id}|{ptype}|v|{channel}|{fplay}",
            ),
        ],
        [
            Button.inline(
                text=_["CLOSE_BUTTON"], data=f"forceclose {videoid}|{user_id}"
            ),
        ],
    ]
    return buttons


## Live Stream Markup


def livestream_markup(_, videoid, user_id, mode, channel, fplay):
    buttons = [
        [
            Button.inline(
                text=_["P_B_3"],
                data=f"LiveStream {videoid}|{user_id}|{mode}|{channel}|{fplay}",
            ),
            Button.inline(
                text=_["CLOSEMENU_BUTTON"],
                data=f"forceclose {videoid}|{user_id}",
            ),
        ],
    ]
    return buttons


## Slider Query Markup


def slider_markup(_, videoid, user_id, query, query_type, channel, fplay):
    query = f"{query[:20]}"
    buttons = [
        [
            Button.inline(
                text=_["P_B_1"],
                data=f"MusicStream {videoid}|{user_id}|a|{channel}|{fplay}",
            ),
            Button.inline(
                text=_["P_B_2"],
                data=f"MusicStream {videoid}|{user_id}|v|{channel}|{fplay}",
            ),
        ],
        [
            Button.inline(
                text="❮",
                data=f"slider B|{query_type}|{query}|{user_id}|{channel}|{fplay}",
            ),
            Button.inline(text=_["CLOSE_BUTTON"], data=f"forceclose {query}|{user_id}"),
            Button.inline(
                text="❯",
                data=f"slider F|{query_type}|{query}|{user_id}|{channel}|{fplay}",
            ),
        ],
    ]
    return buttons


def panel_markup_1(_, videoid, chat_id):
    buttons = [
        [
            Button.inline(text="⏸ Pause", data=f"ADMIN Pause|{chat_id}"),
            Button.inline(
                text="▶️ Resume",
                data=f"ADMIN Resume|{chat_id}",
            ),
        ],
        [
            Button.inline(text="⏯ Skip", data=f"ADMIN Skip|{chat_id}"),
            Button.inline(text="⏹ Stop", data=f"ADMIN Stop|{chat_id}"),
        ],
        [
            Button.inline(text="🔁 Replay ", data=f"ADMIN Replay|{chat_id}"),
        ],
        [
            Button.inline(
                text="◀️",
                data=f"Pages Back|0|{videoid}|{chat_id}",
            ),
            Button.inline(
                text="🔙 Back",
                data=f"MainMarkup {videoid}|{chat_id}",
            ),
            Button.inline(
                text="▶️",
                data=f"Pages Forw|0|{videoid}|{chat_id}",
            ),
        ],
    ]
    return buttons


def panel_markup_2(_, videoid, chat_id):
    buttons = [
        [
            Button.inline(text="🔇 Mute", data=f"ADMIN Mute|{chat_id}"),
            Button.inline(
                text="🔊 Unmute",
                data=f"ADMIN Unmute|{chat_id}",
            ),
        ],
        [
            Button.inline(
                text="🔀 Shuffle",
                data=f"ADMIN Shuffle|{chat_id}",
            ),
            Button.inline(text="🔁 Loop", data=f"ADMIN Loop|{chat_id}"),
        ],
        [
            Button.inline(
                text="◀️",
                data=f"Pages Back|1|{videoid}|{chat_id}",
            ),
            Button.inline(
                text="🔙 Back",
                data=f"MainMarkup {videoid}|{chat_id}",
            ),
            Button.inline(
                text="▶️",
                data=f"Pages Forw|1|{videoid}|{chat_id}",
            ),
        ],
    ]
    return buttons


def panel_markup_3(_, videoid, chat_id):
    buttons = [
        [
            Button.inline(
                text="⏮ 10 seconds",
                data=f"ADMIN 1|{chat_id}",
            ),
            Button.inline(
                text="⏭ 10 seconds",
                data=f"ADMIN 2|{chat_id}",
            ),
        ],
        [
            Button.inline(
                text="⏮ 30 seconds",
                data=f"ADMIN 3|{chat_id}",
            ),
            Button.inline(
                text="⏭ 30 seconds",
                data=f"ADMIN 4|{chat_id}",
            ),
        ],
        [
            Button.inline(
                text="◀️",
                data=f"Pages Back|2|{videoid}|{chat_id}",
            ),
            Button.inline(
                text="🔙 Back",
                data=f"MainMarkup {videoid}|{chat_id}",
            ),
            Button.inline(
                text="▶️",
                data=f"Pages Forw|2|{videoid}|{chat_id}",
            ),
        ],
    ]
    return buttons

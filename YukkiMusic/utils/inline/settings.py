#
# Copyright (C) 2024-2025-2025-2025-2025-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram.types import InlineKeyboardButton


def setting_markup(_):
    buttons = [
        [
            InlineKeyboardButton(text=_["ST_B_1"], callback_data="AQ"),
            InlineKeyboardButton(text=_["ST_B_2"], callback_data="VQ"),
        ],
        [
            InlineKeyboardButton(text=_["ST_B_3"], callback_data="AU"),
            InlineKeyboardButton(text=_["ST_B_6"], callback_data="LG"),
        ],
        [
            InlineKeyboardButton(text=_["ST_B_5"], callback_data="PM"),
            InlineKeyboardButton(text=_["ST_B_7"], callback_data="CM"),
        ],
        [
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        ],
    ]
    return buttons


def audio_quality_markup(
    _,
    low: bool | str = None,
    medium: bool | str = None,
    high: bool | str = None,
    studio: bool | str = None,
):
    buttons = [
        [
            InlineKeyboardButton(
                text=(
                    _["ST_B_8"].format("✅") if low is True else _["ST_B_8"].format("")
                ),
                callback_data="LOW",
            ),
            InlineKeyboardButton(
                text=(
                    _["ST_B_9"].format("✅")
                    if medium is True
                    else _["ST_B_9"].format("")
                ),
                callback_data="MEDIUM",
            ),
        ],
        [
            InlineKeyboardButton(
                text=(
                    _["ST_B_10"].format("✅")
                    if high is True
                    else _["ST_B_10"].format("")
                ),
                callback_data="HIGH",
            ),
            InlineKeyboardButton(
                text=(
                    _["ST_B_11"].format("✅")
                    if studio is True
                    else _["ST_B_11"].format("")
                ),
                callback_data="STUDIO",
            ),
        ],
        [
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data="settingsback_helper",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        ],
    ]
    return buttons


def video_quality_markup(
    _,
    sd_360p: bool | str = None,
    sd_480p: bool | str = None,
    hd_720p: bool | str = None,
    fhd_1080p: bool | str = None,
    qhd_2k: bool | str = None,
    uhd_4k: bool | str = None,
):
    buttons = [
        [
            InlineKeyboardButton(
                text=(
                    _["ST_B_12"].format("✅")
                    if sd_360p is True
                    else _["ST_B_12"].format("")
                ),
                callback_data="SD_360p",
            ),
            InlineKeyboardButton(
                text=(
                    _["ST_B_13"].format("✅")
                    if sd_480p is True
                    else _["ST_B_13"].format("")
                ),
                callback_data="SD_480p",
            ),
        ],
        [
            InlineKeyboardButton(
                text=(
                    _["ST_B_14"].format("✅")
                    if hd_720p is True
                    else _["ST_B_14"].format("")
                ),
                callback_data="HD_720p",
            ),
            InlineKeyboardButton(
                text=(
                    _["ST_B_15"].format("✅")
                    if fhd_1080p is True
                    else _["ST_B_15"].format("")
                ),
                callback_data="FHD_1080p",
            ),
        ],
        [
            InlineKeyboardButton(
                text=(
                    _["ST_B_16"].format("✅")
                    if qhd_2k is True
                    else _["ST_B_16"].format("")
                ),
                callback_data="QHD_2K",
            ),
            InlineKeyboardButton(
                text=(
                    _["ST_B_17"].format("✅")
                    if uhd_4k is True
                    else _["ST_B_17"].format("")
                ),
                callback_data="UHD_4K",
            ),
        ],
        [
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data="settingsback_helper",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        ],
    ]
    return buttons


def cleanmode_settings_markup(
    _,
    status: bool | str = None,
    dels: bool | str = None,
):
    buttons = [
        [
            InlineKeyboardButton(text=_["ST_B_7"], callback_data="CMANSWER"),
            InlineKeyboardButton(
                text=_["ST_B_18"] if status is True else _["ST_B_19"],
                callback_data="CLEANMODE",
            ),
        ],
        [
            InlineKeyboardButton(text=_["ST_B_30"], callback_data="COMMANDANSWER"),
            InlineKeyboardButton(
                text=_["ST_B_18"] if dels is True else _["ST_B_19"],
                callback_data="COMMANDELMODE",
            ),
        ],
        [
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data="settingsback_helper",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        ],
    ]
    return buttons


def auth_users_markup(_, status: bool | str = None):
    buttons = [
        [
            InlineKeyboardButton(text=_["ST_B_3"], callback_data="AUTHANSWER"),
            InlineKeyboardButton(
                text=_["ST_B_20"] if status is True else _["ST_B_21"],
                callback_data="AUTH",
            ),
        ],
        [
            InlineKeyboardButton(text=_["ST_B_22"], callback_data="AUTHLIST"),
        ],
        [
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data="settingsback_helper",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        ],
    ]
    return buttons


def playmode_users_markup(
    _,
    direct: bool | str = None,
    group: bool | str = None,
    playtype: bool | str = None,
):
    buttons = [
        [
            InlineKeyboardButton(text=_["ST_B_23"], callback_data="SEARCHANSWER"),
            InlineKeyboardButton(
                text=_["ST_B_24"] if direct is True else _["ST_B_25"],
                callback_data="MODECHANGE",
            ),
        ],
        [
            InlineKeyboardButton(text=_["ST_B_26"], callback_data="AUTHANSWER"),
            InlineKeyboardButton(
                text=_["ST_B_20"] if group is True else _["ST_B_21"],
                callback_data="CHANNELMODECHANGE",
            ),
        ],
        [
            InlineKeyboardButton(text=_["ST_B_29"], callback_data="PLAYTYPEANSWER"),
            InlineKeyboardButton(
                text=_["ST_B_20"] if playtype is True else _["ST_B_21"],
                callback_data="PLAYTYPECHANGE",
            ),
        ],
        [
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data="settingsback_helper",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        ],
    ]
    return buttons

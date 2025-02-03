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
def setting_markup(_):
    buttons = [
        [
            Button.inline(text=_["ST_B_1"], data="AQ"),
            Button.inline(text=_["ST_B_2"], data="VQ"),
        ],
        [
            Button.inline(text=_["ST_B_3"], data="AU"),
            Button.inline(text=_["ST_B_6"], data="LG"),
        ],
        [
            Button.inline(text=_["ST_B_5"], data="PM"),
            Button.inline(text=_["ST_B_7"], data="CM"),
        ],
        [
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
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
            Button.inline(
                text=(
                    _["ST_B_8"].format("✅") if low is True else _["ST_B_8"].format("")
                ),
                data="LOW",
            ),
            Button.inline(
                text=(
                    _["ST_B_9"].format("✅")
                    if medium is True
                    else _["ST_B_9"].format("")
                ),
                data="MEDIUM",
            ),
        ],
        [
            Button.inline(
                text=(
                    _["ST_B_10"].format("✅")
                    if high is True
                    else _["ST_B_10"].format("")
                ),
                data="HIGH",
            ),
            Button.inline(
                text=(
                    _["ST_B_11"].format("✅")
                    if studio is True
                    else _["ST_B_11"].format("")
                ),
                data="STUDIO",
            ),
        ],
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="settingsback_helper",
            ),
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
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
            Button.inline(
                text=(
                    _["ST_B_12"].format("✅")
                    if sd_360p is True
                    else _["ST_B_12"].format("")
                ),
                data="SD_360p",
            ),
            Button.inline(
                text=(
                    _["ST_B_13"].format("✅")
                    if sd_480p is True
                    else _["ST_B_13"].format("")
                ),
                data="SD_480p",
            ),
        ],
        [
            Button.inline(
                text=(
                    _["ST_B_14"].format("✅")
                    if hd_720p is True
                    else _["ST_B_14"].format("")
                ),
                data="HD_720p",
            ),
            Button.inline(
                text=(
                    _["ST_B_15"].format("✅")
                    if fhd_1080p is True
                    else _["ST_B_15"].format("")
                ),
                data="FHD_1080p",
            ),
        ],
        [
            Button.inline(
                text=(
                    _["ST_B_16"].format("✅")
                    if qhd_2k is True
                    else _["ST_B_16"].format("")
                ),
                data="QHD_2K",
            ),
            Button.inline(
                text=(
                    _["ST_B_17"].format("✅")
                    if uhd_4k is True
                    else _["ST_B_17"].format("")
                ),
                data="UHD_4K",
            ),
        ],
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="settingsback_helper",
            ),
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
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
            Button.inline(text=_["ST_B_7"], data="CMANSWER"),
            Button.inline(
                text=_["ST_B_18"] if status is True else _["ST_B_19"],
                data="CLEANMODE",
            ),
        ],
        [
            Button.inline(text=_["ST_B_30"], data="COMMANDANSWER"),
            Button.inline(
                text=_["ST_B_18"] if dels is True else _["ST_B_19"],
                data="COMMANDELMODE",
            ),
        ],
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="settingsback_helper",
            ),
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
        ],
    ]
    return buttons


def auth_users_markup(_, status: bool | str = None):
    buttons = [
        [
            Button.inline(text=_["ST_B_3"], data="AUTHANSWER"),
            Button.inline(
                text=_["ST_B_20"] if status is True else _["ST_B_21"],
                data="AUTH",
            ),
        ],
        [
            Button.inline(text=_["ST_B_22"], data="AUTHLIST"),
        ],
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="settingsback_helper",
            ),
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
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
            Button.inline(text=_["ST_B_23"], data="SEARCHANSWER"),
            Button.inline(
                text=_["ST_B_24"] if direct is True else _["ST_B_25"],
                data="MODECHANGE",
            ),
        ],
        [
            Button.inline(text=_["ST_B_26"], data="AUTHANSWER"),
            Button.inline(
                text=_["ST_B_20"] if group is True else _["ST_B_21"],
                data="CHANNELMODECHANGE",
            ),
        ],
        [
            Button.inline(text=_["ST_B_29"], data="PLAYTYPEANSWER"),
            Button.inline(
                text=_["ST_B_20"] if playtype is True else _["ST_B_21"],
                data="PLAYTYPECHANGE",
            ),
        ],
        [
            Button.inline(
                text=_["BACK_BUTTON"],
                data="settingsback_helper",
            ),
            Button.inline(text=_["CLOSE_BUTTON"], data="close"),
        ],
    ]
    return buttons
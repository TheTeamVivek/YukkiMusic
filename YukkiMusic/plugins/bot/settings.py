#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from telethon import Button, events
from telethon.errors import MessageNotModifiedError

from config import BANNED_USERS, CLEANMODE_DELETE_TIME, OWNER_ID
from YukkiMusic import tbot
from YukkiMusic.utils.database import (
    add_nonadmin_chat,
    cleanmode_off,
    cleanmode_on,
    commanddelete_off,
    commanddelete_on,
    get_aud_bit_name,
    get_authuser,
    get_authuser_names,
    get_playmode,
    get_playtype,
    get_vid_bit_name,
    is_cleanmode_on,
    is_commanddelete_on,
    is_nonadmin_chat,
    remove_nonadmin_chat,
    save_audio_bitrate,
    save_video_bitrate,
    set_playmode,
    set_playtype,
)
from YukkiMusic.utils.decorators.admins import actual_admin_cb
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.inline.settings import (
    audio_quality_markup,
    auth_users_markup,
    cleanmode_settings_markup,
    playmode_users_markup,
    setting_markup,
    video_quality_markup,
)
from YukkiMusic.utils.inline.start import private_panel


@tbot.on_message(
    flt.command("SETTINGS_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS)
)
@language
async def settings_mar(event, _):
    buttons = setting_markup(_)
    chat = await event.get_chat()
    await event.reply(
        _["setting_1"].format(chat.title, chat.id),
        buttons=buttons,
    )


@tbot.on(events.CallbackQuery(pattern="settings_helper", func=~flt.user(BANNED_USERS)))
@language
async def settings_cb(event, _):
    try:
        await event.answer(_["set_cb_8"])
    except Exception:
        pass
    chat = await event.get_chat()
    buttons = setting_markup(_)
    return await event.edit(
        text=_["setting_1"].format(
            chat.title,
            event.chat_id,
        ),
        buttons=buttons,
    )


@tbot.on(
    events.CallbackQuery(pattern="settingsback_helper", func=~flt.user(BANNED_USERS))
)
@language
async def settings_back_markup(event, _):
    try:
        await event.answer()
    except Exception:
        pass

    if event.is_private:
        try:
            await tbot.get_entity(OWNER_ID[0])
            OWNER = OWNER_ID[0]
        except Exception:
            OWNER = None
        buttons = private_panel(_, OWNER)
        await event.edit(
            _["start_1"].format(tbot.mention),
            buttons=buttons,
        )
    else:
        buttons = setting_markup(_)
        await event.edit(buttons=buttons)


## Audio and Video Quality
async def gen_buttons_aud(_, aud):
    if aud == "STUDIO":
        buttons = audio_quality_markup(_, STUDIO=True)
    elif aud == "HIGH":
        buttons = audio_quality_markup(_, HIGH=True)
    elif aud == "MEDIUM":
        buttons = audio_quality_markup(_, MEDIUM=True)
    elif aud == "LOW":
        buttons = audio_quality_markup(_, LOW=True)
    return buttons


async def gen_buttons_vid(_, aud):
    if aud == "UHD_4K":
        buttons = video_quality_markup(_, UHD_4K=True)
    elif aud == "QHD_2K":
        buttons = video_quality_markup(_, QHD_2K=True)
    elif aud == "FHD_1080p":
        buttons = video_quality_markup(_, FHD_1080p=True)
    elif aud == "HD_720p":
        buttons = video_quality_markup(_, HD_720p=True)
    elif aud == "SD_480p":
        buttons = video_quality_markup(_, SD_480p=True)
    elif aud == "SD_360p":
        buttons = video_quality_markup(_, SD_360p=True)
    return buttons


# without admin rights


@tbot.on(
    events.CallbackQuery(
        pattern=r"^(SEARCHANSWER|PLAYMODEANSWER|PLAYTYPEANSWER|AUTHANSWER|CMANSWER|COMMANDANSWER|CM|AQ|VQ|PM|AU)$",
        func=~flt.user(BANNED_USERS),
    )
)
@language
async def without_Admin_rights(event, _):
    command = event.pattern_match.group(0)
    if command == "SEARCHANSWER":
        try:
            return await event.answer(_["setting_3"], alert=True)
        except Exception:
            return
    if command == "PLAYMODEANSWER":
        try:
            return await event.answer(_["setting_10"], alert=True)
        except Exception:
            return
    if command == "PLAYTYPEANSWER":
        try:
            return await event.answer(_["setting_11"], alert=True)
        except Exception:
            return
    if command == "AUTHANSWER":
        try:
            return await event.answer(_["setting_4"], alert=True)
        except Exception:
            return
    if command == "CMANSWER":
        try:
            return await event.answer(
                _["setting_9"].format(CLEANMODE_DELETE_TIME),
                alert=True,
            )
        except Exception:
            return
    if command == "COMMANDANSWER":
        try:
            return await event.answer(_["setting_14"], alert=True)
        except Exception:
            return
    if command == "CM":
        try:
            await event.answer(_["set_cb_5"], alert=True)
        except Exception:
            pass
        sta = None
        cle = None
        if await is_cleanmode_on(event.chat_id):
            cle = True
        if await is_commanddelete_on(event.chat_id):
            sta = True
        buttons = cleanmode_settings_markup(_, status=cle, dels=sta)

    if command == "AQ":
        try:
            await event.answer(_["set_cb_1"], alert=True)
        except Exception:
            pass
        aud = await get_aud_bit_name(event.chat_id)
        buttons = await gen_buttons_aud(_, aud)
    if command == "VQ":
        try:
            await event.answer(_["set_cb_2"], alert=True)
        except Exception:
            pass
        aud = await get_vid_bit_name(event.chat_id)
        buttons = await gen_buttons_vid(_, aud)
    if command == "PM":
        try:
            await event.answer(_["set_cb_4"], alert=True)
        except Exception:
            pass
        playmode = await get_playmode(event.chat_id)
        if playmode == "Direct":
            Direct = True
        else:
            Direct = None
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            Group = True
        else:
            Group = None
        playty = await get_playtype(event.chat_id)
        if playty == "Everyone":
            Playtype = None
        else:
            Playtype = True
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    if command == "AU":
        try:
            await event.answer(_["set_cb_3"], alert=True)
        except Exception:
            pass
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            buttons = auth_users_markup(_, True)
        else:
            buttons = auth_users_markup(_)
        return await event.edit(buttons=buttons)


# Audio Video Quality


@tbot.on(
    events.CallbackQuery(
        pattern=r"^(LOW|MEDIUM|HIGH|STUDIO|SD_360p|SD_480p|HD_720p|FHD_1080p|QHD_2K|UHD_4K)$",
        func=~flt.user(BANNED_USERS),
    )
)
@actual_admin_cb
async def aud_vid_cb(event, _):
    command = event.pattern_match.group(0)
    try:
        await event.answer(_["set_cb_6"], alert=True)
    except Exception:
        pass
    if command == "LOW":
        await save_audio_bitrate(event.chat_id, "LOW")
        buttons = audio_quality_markup(_, LOW=True)
    if command == "MEDIUM":
        await save_audio_bitrate(event.chat_id, "MEDIUM")
        buttons = audio_quality_markup(_, MEDIUM=True)
    if command == "HIGH":
        await save_audio_bitrate(event.chat_id, "HIGH")
        buttons = audio_quality_markup(_, HIGH=True)
    if command == "STUDIO":
        await save_audio_bitrate(event.chat_id, "STUDIO")
        buttons = audio_quality_markup(_, STUDIO=True)
    if command == "SD_360p":
        await save_video_bitrate(event.chat_id, "SD_360p")
        buttons = video_quality_markup(_, SD_360p=True)
    if command == "SD_480p":
        await save_video_bitrate(event.chat_id, "SD_480p")
        buttons = video_quality_markup(_, SD_480p=True)
    if command == "HD_720p":
        await save_video_bitrate(event.chat_id, "HD_720p")
        buttons = video_quality_markup(_, HD_720p=True)
    if command == "FHD_1080p":
        await save_video_bitrate(event.chat_id, "FHD_1080p")
        buttons = video_quality_markup(_, FHD_1080p=True)
    if command == "QHD_2K":
        await save_video_bitrate(event.chat_id, "QHD_2K")
        buttons = video_quality_markup(_, QHD_2K=True)
    if command == "UHD_4K":
        await save_video_bitrate(event.chat_id, "UHD_4K")
        buttons = video_quality_markup(_, UHD_4K=True)
    return await event.edit(buttons=buttons)


@tbot.on(
    events.CallbackQuery(
        pattern=r"^(CLEANMODE|COMMANDELMODE)$", func=~flt.user(BANNED_USERS)
    )
)
@actual_admin_cb
async def cleanmode_mark(event, _):
    command = event.pattern_match.group(0)
    try:
        await event.answer(_["set_cb_6"], alert=True)
    except Exception:
        pass
    if command == "CLEANMODE":
        sta = None
        if await is_commanddelete_on(event.chat_id):
            sta = True
        cle = None
        if await is_cleanmode_on(event.chat_id):
            await cleanmode_off(event.chat_id)
        else:
            await cleanmode_on(event.chat_id)
            cle = True
        buttons = cleanmode_settings_markup(_, status=cle, dels=sta)
        return await event.edit(buttons=buttons)
    if command == "COMMANDELMODE":
        cle = None
        sta = None
        if await is_cleanmode_on(event.chat_id):
            cle = True
        if await is_commanddelete_on(event.chat_id):
            await commanddelete_off(event.chat_id)
        else:
            await commanddelete_on(event.chat_id)
            sta = True
        buttons = cleanmode_settings_markup(_, status=cle, dels=sta)
    return await event.edit(buttons=buttons)


# Play Mode Settings
@tbot.on(
    events.CallbackQuery(
        pattern=r"^(|MODECHANGE|CHANNELMODECHANGE|PLAYTYPECHANGE)$",
        func=~flt.user(BANNED_USERS),
    )
)
@actual_admin_cb
async def playmode_ans(event, _):
    command = event.pattern_match.group(0)
    if command == "CHANNELMODECHANGE":
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            await add_nonadmin_chat(event.chat_id)
            Group = None
        else:
            await remove_nonadmin_chat(event.chat_id)
            Group = True
        playmode = await get_playmode(event.chat_id)
        if playmode == "Direct":
            Direct = True
        else:
            Direct = None
        playty = await get_playtype(event.chat_id)
        if playty == "Everyone":
            Playtype = None
        else:
            Playtype = True
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    if command == "MODECHANGE":
        try:
            await event.answer(_["set_cb_6"], alert=True)
        except Exception:
            pass
        playmode = await get_playmode(event.chat_id)
        if playmode == "Direct":
            await set_playmode(event.chat_id, "Inline")
            Direct = None
        else:
            await set_playmode(event.chat_id, "Direct")
            Direct = True
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            Group = True
        else:
            Group = None
        playty = await get_playtype(event.chat_id)
        if playty == "Everyone":
            Playtype = False
        else:
            Playtype = True
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    if command == "PLAYTYPECHANGE":
        try:
            await event.answer(_["set_cb_6"], alert=True)
        except Exception:
            pass
        playty = await get_playtype(event.chat_id)
        if playty == "Everyone":
            await set_playtype(event.chat_id, "Admin")
            Playtype = False
        else:
            await set_playtype(event.chat_id, "Everyone")
            Playtype = True
        playmode = await get_playmode(event.chat_id)
        if playmode == "Direct":
            Direct = True
        else:
            Direct = None
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            Group = True
        else:
            Group = None
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    return await event.edit(buttons=buttons)


# Auth Users Settings
@tbot.on(
    events.CallbackQuery(pattern=r"^(AUTH|AUTHLIST)$", func=~flt.user(BANNED_USERS))
)
@actual_admin_cb
async def authusers_mar(event, _):
    command = event.pattern_match.group(0)
    if command == "AUTHLIST":
        _authusers = await get_authuser_names(event.chat_id)
        if not _authusers:
            try:
                return await event.answer(_["setting_5"], alert=True)
            except Exception:
                return
        else:
            try:
                await event.answer(_["set_cb_7"], alert=True)
            except Exception:
                pass
            j = 0
            await event.edit(_["auth_6"])
            msg = _["auth_7"]
            for note in _authusers:
                _note = await get_authuser(event.chat_id, note)
                user_id = _note["auth_user_id"]
                admin_id = _note["admin_id"]
                admin_name = _note["admin_name"]
                try:
                    user = await tbot.get_entity(user_id)
                    user = user.first_name
                    j += 1
                except Exception:
                    continue
                msg += f"{j}➤ {user}[`{user_id}`]\n"
                msg += f"   {_['auth_8']} {admin_name}[`{admin_id}`]\n\n"
            upl = [
                [
                    Button.inline(text=_["BACK_BUTTON"], data=f"AU"),
                    Button.inline(
                        text=_["CLOSE_BUTTON"],
                        data=f"close",
                    ),
                ]
            ]
            try:
                return await event.edit(msg, buttons=upl)
            except MessageNotModifiedError:
                return
    try:
        await event.answer(_["set_cb_6"], alert=True)
    except Exception:
        pass
    if command == "AUTH":
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            await add_nonadmin_chat(event.chat_id)
            buttons = auth_users_markup(_)
        else:
            await remove_nonadmin_chat(event.chat_id)
            buttons = auth_users_markup(_, True)
    return await event.edit(buttons=buttons)

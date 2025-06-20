#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from pyrogram import filters
from pyrogram.enums import ChatType
from pyrogram.errors import MessageNotModified, PeerIdInvalid, QueryIdInvalid
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from config import BANNED_USERS, CLEANMODE_DELETE_MINS, OWNER_ID
from strings import command
from YukkiMusic import app
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
from YukkiMusic.utils.decorators.admins import ActualAdminCB
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


@app.on_message(command("SETTINGS_COMMAND") & filters.group & ~BANNED_USERS)
@language
async def settings_mar(client, message: Message, _):
    buttons = setting_markup(_)
    await message.reply_text(
        _["setting_1"].format(message.chat.title, message.chat.id),
        reply_markup=InlineKeyboardMarkup(buttons),
    )


@app.on_callback_query(filters.regex("settings_helper") & ~BANNED_USERS)
@language
async def settings_cb(client, query, _):
    try:
        await query.answer(_["set_cb_8"])
    except QueryIdInvalid:
        pass
    buttons = setting_markup(_)
    return await query.edit_message_text(
        _["setting_1"].format(
            query.message.chat.title,
            query.message.chat.id,
        ),
        reply_markup=InlineKeyboardMarkup(buttons),
    )


@app.on_callback_query(filters.regex("settingsback_helper") & ~BANNED_USERS)
@language
async def settings_back_markup(client, query, _):
    try:
        await query.answer()
    except QueryIdInvalid:
        return

    if query.message.chat.type == ChatType.PRIVATE:
        try:
            await app.resolve_peer(OWNER_ID[0])
            OWNER = OWNER_ID[0]
        except PeerIdInvalid:
            OWNER = None
        buttons = private_panel(_, app.username, OWNER)
        try:
            await query.edit_message_text(
                _["start_1"].format(app.mention),
                reply_markup=InlineKeyboardMarkup(buttons),
            )
        except MessageNotModified:
            pass
    else:
        buttons = setting_markup(_)
        try:
            await query.edit_message_reply_markup(
                reply_markup=InlineKeyboardMarkup(buttons)
            )
        except MessageNotModified:
            pass


# Audio and Video Quality
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


@app.on_callback_query(
    filters.regex(
        pattern=r"^(SEARCHANSWER|PLAYMODEANSWER|PLAYTYPEANSWER|AUTHANSWER|CMANSWER|COMMANDANSWER|CM|AQ|VQ|PM|AU)$"
    )
    & ~BANNED_USERS
)
@language
async def without_Admin_rights(client, query, _):
    command = query.matches[0].group(1)
    chat_id = query.message.chat.id
    if command == "SEARCHANSWER":
        try:
            return await query.answer(_["setting_3"], show_alert=True)
        except QueryIdInvalid:
            return
    if command == "PLAYMODEANSWER":
        try:
            return await query.answer(_["setting_10"], show_alert=True)
        except QueryIdInvalid:
            return
    if command == "PLAYTYPEANSWER":
        try:
            return await query.answer(_["setting_11"], show_alert=True)
        except QueryIdInvalid:
            return
    if command == "AUTHANSWER":
        try:
            return await query.answer(_["setting_4"], show_alert=True)
        except QueryIdInvalid:
            return
    if command == "CMANSWER":
        try:
            return await query.answer(
                _["setting_9"].format(CLEANMODE_DELETE_MINS),
                show_alert=True,
            )
        except QueryIdInvalid:
            return
    if command == "COMMANDANSWER":
        try:
            return await query.answer(_["setting_14"], show_alert=True)
        except QueryIdInvalid:
            return
    if command == "CM":
        try:
            await query.answer(_["set_cb_5"], show_alert=True)
        except QueryIdInvalid:
            pass
        sta = None
        cle = None
        if await is_cleanmode_on(chat_id):
            cle = True
        if await is_commanddelete_on(chat_id):
            sta = True
        buttons = cleanmode_settings_markup(_, status=cle, dels=sta)

    if command == "AQ":
        try:
            await query.answer(_["set_cb_1"], show_alert=True)
        except QueryIdInvalid:
            pass
        aud = await get_aud_bit_name(chat_id)
        buttons = await gen_buttons_aud(_, aud)
    if command == "VQ":
        try:
            await query.answer(_["set_cb_2"], show_alert=True)
        except QueryIdInvalid:
            pass
        aud = await get_vid_bit_name(chat_id)
        buttons = await gen_buttons_vid(_, aud)
    if command == "PM":
        try:
            await query.answer(_["set_cb_4"], show_alert=True)
        except QueryIdInvalid:
            pass
        playmode = await get_playmode(chat_id)
        if playmode == "Direct":
            Direct = True
        else:
            Direct = None
        is_non_admin = await is_nonadmin_chat(chat_id)
        if not is_non_admin:
            Group = True
        else:
            Group = None
        playty = await get_playtype(chat_id)
        if playty == "Everyone":
            Playtype = None
        else:
            Playtype = True
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    if command == "AU":
        try:
            await query.answer(_["set_cb_3"], show_alert=True)
        except QueryIdInvalid:
            pass
        is_non_admin = await is_nonadmin_chat(chat_id)
        if not is_non_admin:
            buttons = auth_users_markup(_, True)
        else:
            buttons = auth_users_markup(_)
    try:
        return await query.edit_message_reply_markup(
            reply_markup=InlineKeyboardMarkup(buttons)
        )
    except MessageNotModified:
        return


# Audio Video Quality


@app.on_callback_query(
    filters.regex(
        pattern=r"^(LOW|MEDIUM|HIGH|STUDIO|SD_360p|SD_480p|HD_720p|FHD_1080p|QHD_2K|UHD_4K)$"
    )
    & ~BANNED_USERS
)
@ActualAdminCB
async def aud_vid_cb(client, query, _):
    command = query.matches[0].group(1)
    try:
        await query.answer(_["set_cb_6"], show_alert=True)
    except QueryIdInvalid:
        pass
    chat_id = query.message.chat.id
    if command == "LOW":
        await save_audio_bitrate(chat_id, "LOW")
        buttons = audio_quality_markup(_, LOW=True)
    if command == "MEDIUM":
        await save_audio_bitrate(chat_id, "MEDIUM")
        buttons = audio_quality_markup(_, MEDIUM=True)
    if command == "HIGH":
        await save_audio_bitrate(chat_id, "HIGH")
        buttons = audio_quality_markup(_, HIGH=True)
    if command == "STUDIO":
        await save_audio_bitrate(chat_id, "STUDIO")
        buttons = audio_quality_markup(_, STUDIO=True)
    if command == "SD_360p":
        await save_video_bitrate(chat_id, "SD_360p")
        buttons = video_quality_markup(_, SD_360p=True)
    if command == "SD_480p":
        await save_video_bitrate(chat_id, "SD_480p")
        buttons = video_quality_markup(_, SD_480p=True)
    if command == "HD_720p":
        await save_video_bitrate(chat_id, "HD_720p")
        buttons = video_quality_markup(_, HD_720p=True)
    if command == "FHD_1080p":
        await save_video_bitrate(chat_id, "FHD_1080p")
        buttons = video_quality_markup(_, FHD_1080p=True)
    if command == "QHD_2K":
        await save_video_bitrate(chat_id, "QHD_2K")
        buttons = video_quality_markup(_, QHD_2K=True)
    if command == "UHD_4K":
        await save_video_bitrate(chat_id, "UHD_4K")
        buttons = video_quality_markup(_, UHD_4K=True)
    try:
        return await query.edit_message_reply_markup(
            reply_markup=InlineKeyboardMarkup(buttons)
        )
    except MessageNotModified:
        return


@app.on_callback_query(
    filters.regex(pattern=r"^(CLEANMODE|COMMANDELMODE)$") & ~BANNED_USERS
)
@ActualAdminCB
async def cleanmode_mark(client, query, _):
    command = query.matches[0].group(1)
    chat_id = query.message.chat.id

    try:
        await query.answer(_["set_cb_6"], show_alert=True)
    except QueryIdInvalid:
        pass
    if command == "CLEANMODE":
        sta = None
        if await is_commanddelete_on(chat_id):
            sta = True
        cle = None
        if await is_cleanmode_on(chat_id):
            await cleanmode_off(chat_id)
        else:
            await cleanmode_on(chat_id)
            cle = True
        buttons = cleanmode_settings_markup(_, status=cle, dels=sta)
        return await query.edit_message_reply_markup(
            reply_markup=InlineKeyboardMarkup(buttons)
        )
    if command == "COMMANDELMODE":
        cle = None
        sta = None
        if await is_cleanmode_on(chat_id):
            cle = True
        if await is_commanddelete_on(chat_id):
            await commanddelete_off(chat_id)
        else:
            await commanddelete_on(chat_id)
            sta = True
        buttons = cleanmode_settings_markup(_, status=cle, dels=sta)
    try:
        return await query.edit_message_reply_markup(
            reply_markup=InlineKeyboardMarkup(buttons)
        )
    except MessageNotModified:
        return


# Play Mode Settings
@app.on_callback_query(
    filters.regex(pattern=r"^(|MODECHANGE|CHANNELMODECHANGE|PLAYTYPECHANGE)$")
    & ~BANNED_USERS
)
@ActualAdminCB
async def playmode_ans(client, query, _):
    command = query.matches[0].group(1)
    chat_id = query.message.chat.id
    if command == "CHANNELMODECHANGE":
        is_non_admin = await is_nonadmin_chat(chat_id)
        if not is_non_admin:
            await add_nonadmin_chat(chat_id)
            Group = None
        else:
            await remove_nonadmin_chat(chat_id)
            Group = True
        playmode = await get_playmode(chat_id)
        if playmode == "Direct":
            Direct = True
        else:
            Direct = None
        playty = await get_playtype(chat_id)
        if playty == "Everyone":
            Playtype = None
        else:
            Playtype = True
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    if command == "MODECHANGE":
        try:
            await query.answer(_["set_cb_6"], show_alert=True)
        except QueryIdInvalid:
            pass
        playmode = await get_playmode(chat_id)
        if playmode == "Direct":
            await set_playmode(chat_id, "Inline")
            Direct = None
        else:
            await set_playmode(chat_id, "Direct")
            Direct = True
        is_non_admin = await is_nonadmin_chat(chat_id)
        if not is_non_admin:
            Group = True
        else:
            Group = None
        playty = await get_playtype(chat_id)
        if playty == "Everyone":
            Playtype = False
        else:
            Playtype = True
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    if command == "PLAYTYPECHANGE":
        try:
            await query.answer(_["set_cb_6"], show_alert=True)
        except QueryIdInvalid:
            pass
        playty = await get_playtype(chat_id)
        if playty == "Everyone":
            await set_playtype(chat_id, "Admin")
            Playtype = False
        else:
            await set_playtype(chat_id, "Everyone")
            Playtype = True
        playmode = await get_playmode(chat_id)
        if playmode == "Direct":
            Direct = True
        else:
            Direct = None
        is_non_admin = await is_nonadmin_chat(chat_id)
        if not is_non_admin:
            Group = True
        else:
            Group = None
        buttons = playmode_users_markup(_, Direct, Group, Playtype)
    try:
        return await query.edit_message_reply_markup(
            reply_markup=InlineKeyboardMarkup(buttons)
        )
    except MessageNotModified:
        return


# Auth Users Settings
@app.on_callback_query(filters.regex(pattern=r"^(AUTH|AUTHLIST)$") & ~BANNED_USERS)
@ActualAdminCB
async def authusers_mar(client, query, _):
    command = query.matches[0].group(1)
    chat_id = query.message.chat.id
    if command == "AUTHLIST":
        _authusers = await get_authuser_names(chat_id)
        if not _authusers:
            try:
                return await query.answer(_["setting_5"], show_alert=True)
            except QueryIdInvalid:
                return
        else:
            try:
                await query.answer(_["set_cb_7"], show_alert=True)
            except QueryIdInvalid:
                pass
            j = 0
            await query.edit_message_text(_["auth_6"])
            msg = _["auth_7"]
            for note in _authusers:
                _note = await get_authuser(chat_id, note)
                user_id = _note["auth_user_id"]
                admin_id = _note["admin_id"]
                admin_name = _note["admin_name"]
                try:
                    user = await client.get_users(user_id)
                    user = user.first_name
                    j += 1
                except PeerIdInvalid:
                    continue
                msg += f"{j}➤ {user}[`{user_id}`]\n"
                msg += f"   {_['auth_8']} {admin_name}[`{admin_id}`]\n\n"
            upl = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(text=_["BACK_BUTTON"], callback_data="AU"),
                        InlineKeyboardButton(
                            text=_["CLOSE_BUTTON"],
                            callback_data="close",
                        ),
                    ]
                ]
            )
            try:
                return await query.edit_message_text(msg, reply_markup=upl)
            except MessageNotModified:
                return
    try:
        await query.answer(_["set_cb_6"], show_alert=True)
    except QueryIdInvalid:
        pass
    if command == "AUTH":
        is_non_admin = await is_nonadmin_chat(chat_id)
        if not is_non_admin:
            await add_nonadmin_chat(chat_id)
            buttons = auth_users_markup(_)
        else:
            await remove_nonadmin_chat(chat_id)
            buttons = auth_users_markup(_, True)
    try:
        return await query.edit_message_reply_markup(
            reply_markup=InlineKeyboardMarkup(buttons)
        )
    except MessageNotModified:
        return

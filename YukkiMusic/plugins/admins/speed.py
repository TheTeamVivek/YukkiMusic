#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from pyrogram import filters
from pyrogram.types import Message
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import SUDOERS, db
from YukkiMusic.utils import AdminRightsCheck
from YukkiMusic.utils.database import is_active_chat, is_nonadmin_chat
from YukkiMusic.utils.decorators.language import languageCB
from config import BANNED_USERS, adminlist

checker = []


def speed_markup(_, chat_id):
    upl = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    text="0.5x",
                    callback_data=f"SpeedUP {chat_id}|0.5",
                ),
                InlineKeyboardButton(
                    text="0.75x",
                    callback_data=f"SpeedUP {chat_id}|0.75",
                ),
            ],
            [
                InlineKeyboardButton(
                    text="1.0x",
                    callback_data=f"SpeedUP {chat_id}|1.0",
                ),
            ],
            [
                InlineKeyboardButton(
                    text="1.5x",
                    callback_data=f"SpeedUP {chat_id}|1.5",
                ),
                InlineKeyboardButton(
                    text="2.0x",
                    callback_data=f"SpeedUP {chat_id}|2.0",
                ),
            ],
            [
                InlineKeyboardButton(
                    text="close",
                    callback_data="close",
                ),
            ],
        ]
    )
    return upl


def close_markup(_):
    upl = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    text="close",
                    callback_data="close",
                ),
            ]
        ]
    )

    return upl


@app.on_message(
    filters.command(["cspeed", "speed", "cslow", "slow", "playback", "cplayback"])
    & filters.group
    & ~BANNED_USERS
)
@AdminRightsCheck
async def playback(cli, message: Message, _, chat_id):
    playing = db.get(chat_id)
    if not playing:
        return await message.reply_text("ǫᴜᴇᴜᴇ ᴇᴍᴘᴛʏ.")
    duration_seconds = int(playing[0]["seconds"])
    if duration_seconds == 0:
        return await message.reply_text(
            "ᴏɴʟʏ ʏᴏᴜᴛᴜʙᴇ sᴛʀᴇᴀᴍ's sᴘᴇᴇᴅ ᴄᴀɴ ʙᴇ ᴄᴏɴᴛʀᴏʟʟᴇᴅ ᴄᴜʀʀᴇɴᴛʟʏ."
        )
    file_path = playing[0]["file"]
    if "downloads" not in file_path:
        return await message.reply_text(
            "ᴏɴʟʏ ʏᴏᴜᴛᴜʙᴇ sᴛʀᴇᴀᴍ's sᴘᴇᴇᴅ ᴄᴀɴ ʙᴇ ᴄᴏɴᴛʀᴏʟʟᴇᴅ ᴄᴜʀʀᴇɴᴛʟʏ."
        )
    upl = speed_markup(_, chat_id)
    return await message.reply_text(
        text="<b><u>{0} sᴘᴇᴇᴅ ᴄᴏɴᴛʀᴏʟ ᴘᴀɴᴇʟ</b></u>\n\nᴄʟɪᴄᴋ ᴏɴ ᴛʜᴇ ʙᴜᴛᴛᴏɴs ʙᴇʟᴏᴡ ᴛᴏ ᴄʜᴀɴɢᴇ ᴛʜᴇ sᴘᴇᴇᴅ ᴏғ ᴄᴜʀʀᴇɴᴛʟʏ ᴘʟᴀʏɪɴɢ sᴛʀᴇᴀᴍ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ.".format(
            app.mention
        ),
        reply_markup=upl,
    )


@app.on_callback_query(filters.regex("SpeedUP") & ~BANNED_USERS)
@languageCB
async def del_back_playlist(client, CallbackQuery, _):
    callback_data = CallbackQuery.data.strip()
    callback_request = callback_data.split(None, 1)[1]
    chat, speed = callback_request.split("|")
    chat_id = int(chat)
    if not await is_active_chat(chat_id):
        return await CallbackQuery.answer(
            "ʙᴏᴛ ɪsɴ'ᴛ sᴛʀᴇᴀᴍɪɴɢ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ", show_alert=True
        )
    is_non_admin = await is_nonadmin_chat(CallbackQuery.message.chat.id)
    if not is_non_admin:
        if CallbackQuery.from_user.id not in SUDOERS:
            admins = adminlist.get(CallbackQuery.message.chat.id)
            if not admins:
                return await CallbackQuery.answer(
                    "ᴘʟᴇᴀsᴇ ʀᴇʟᴏᴀᴅ ᴀᴅᴍɪɴ ᴄᴀᴄʜᴇ ᴠɪᴀ : /reload", show_alert=True
                )
            else:
                if CallbackQuery.from_user.id not in admins:
                    return await CallbackQuery.answer(
                        "ʏᴏᴜ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ᴘᴇʀᴍɪssɪᴏɴs ᴛᴏ ᴍᴀɴᴀɢᴇ ᴠɪᴅᴇᴏ ᴄʜᴀᴛs.\n\nʀᴇʟᴏᴀᴅ ᴀᴅᴍɪɴ ᴄᴀᴄʜᴇ ᴠɪᴀ : /reload",
                        show_alert=True,
                    )
    playing = db.get(chat_id)
    if not playing:
        return await CallbackQuery.answer("ǫᴜᴇᴜᴇ ᴇᴍᴘᴛʏ.", show_alert=True)
    duration_seconds = int(playing[0]["seconds"])
    if duration_seconds == 0:
        return await CallbackQuery.answer(
            "ᴏɴʟʏ ʏᴏᴜᴛᴜʙᴇ sᴛʀᴇᴀᴍ's sᴘᴇᴇᴅ ᴄᴀɴ ʙᴇ ᴄᴏɴᴛʀᴏʟʟᴇᴅ ᴄᴜʀʀᴇɴᴛʟʏ.", show_alert=True
        )
    file_path = playing[0]["file"]
    if "downloads" not in file_path:
        return await CallbackQuery.answer(
            "ᴏɴʟʏ ʏᴏᴜᴛᴜʙᴇ sᴛʀᴇᴀᴍ's sᴘᴇᴇᴅ ᴄᴀɴ ʙᴇ ᴄᴏɴᴛʀᴏʟʟᴇᴅ ᴄᴜʀʀᴇɴᴛʟʏ.", show_alert=True
        )
    checkspeed = (playing[0]).get("speed")
    if checkspeed:
        if str(checkspeed) == str(speed):
            if str(speed) == str("1.0"):
                return await CallbackQuery.answer(
                    "ʙᴏᴛ ɪs ᴀʟʀᴇᴀᴅʏ ᴘʟᴀʏɪɴɢ ᴏɴ ɴᴏʀᴍᴀʟ sᴘᴇᴇᴅ",
                    show_alert=True,
                )
    else:
        if str(speed) == str("1.0"):
            return await CallbackQuery.answer(
                "ʙᴏᴛ ɪs ᴀʟʀᴇᴀᴅʏ ᴘʟᴀʏɪɴɢ ᴏɴ ɴᴏʀᴍᴀʟ sᴘᴇᴇᴅ",
                show_alert=True,
            )
    if chat_id in checker:
        return await CallbackQuery.answer(
            "» ᴘʟᴇᴀsᴇ ᴡᴀɪᴛ...\n\nsᴏᴍᴇᴏɴᴇ ᴇʟsᴇ ɪs ᴛʀʏɪɴɢ ᴛᴏ ᴄʜᴀɴɢᴇ ᴛʜᴇ sᴘᴇᴇᴅ ᴏғ ᴛʜᴇ sᴛʀᴇᴀᴍ.",
            show_alert=True,
        )
    else:
        checker.append(chat_id)
    try:
        await CallbackQuery.answer(
            "ᴄʜᴀɴɢɪɴɢ sᴘᴇᴇᴅ...",
        )
    except:
        pass
    mystic = await CallbackQuery.edit_message_text(
        text="» ᴛʀʏɪɴɢ ᴛᴏ ᴄʜᴀɴɢᴇ ᴛʜᴇ sᴘᴇᴇᴅ ᴏғ ᴛʜᴇ ᴏɴɢᴏɪɴɢ sᴛʀᴇᴀᴍ...\n\n<b>ʀᴇǫᴜᴇsᴛᴇᴅ ʙʏ :</b> {0}".format(
            CallbackQuery.from_user.mention
        ),
    )
    try:
        await Yukki.speedup_stream(
            chat_id,
            file_path,
            speed,
            playing,
        )
    except:
        if chat_id in checker:
            checker.remove(chat_id)
        return await mystic.edit_text(
            "» ғᴀɪʟᴇᴅ ᴛᴏ ᴄʜᴀɴɢᴇ ᴛʜᴇ sᴘᴇᴇᴅ ᴏғ ᴛʜᴇ ᴏɴɢᴏɪɴɢ sᴛʀᴇᴀᴍ.",
            reply_markup=close_markup(_),
        )
    if chat_id in checker:
        checker.remove(chat_id)
    await mystic.edit_text(
        text="» ᴄʜᴀɴɢᴇᴅ ᴛʜᴇ sᴘᴇᴇᴅ ᴏғ ᴛʜᴇ ᴏɴɢᴏɪɴɢ sᴛʀᴇᴀᴍ ᴛᴏ {0}x\n\n<b>ʀᴇǫᴜᴇsᴛᴇᴅ ʙʏ :</b> {1}".format(
            speed, CallbackQuery.from_user.mention
        ),
        reply_markup=close_markup(_),
    )

from pyrogram import Client, filters
from pyrogram.types import CallbackQuery, InlineKeyboardMarkup
from pytgcalls.types import MediaStream, AudioQuality

from config import *
from YukkiMusic import LOGGER
from YukkiMusic.misc import clonedb
from YukkiMusic.utils.thumbnails import gen_thumb
from .utils import (
    admin_check_cb,
    stream_off,
    stream_on,
    is_streaming,
    _clear_,
    HELP_TEXT,
    PM_START_TEXT,
    HELP_DEV,
    HELP_SUDO,
)
from .play import pytgcalls


@app.on_callback_query(filters.regex("forceclose"))
async def close_(_, CallbackQuery):
    callback_data = CallbackQuery.data.strip()
    callback_request = callback_data.split(None, 1)[1]
    query, user_id = callback_request.split("|")
    if CallbackQuery.from_user.id != int(user_id):
        try:
            return await CallbackQuery.answer(
                "» ɪᴛ'ʟʟ ʙᴇ ʙᴇᴛᴛᴇʀ ɪғ ʏᴏᴜ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs ʙᴀʙʏ.", show_alert=True
            )
        except:
            return
    await CallbackQuery.message.delete()
    try:
        await CallbackQuery.answer()
    except:
        return


@app.on_callback_query(filters.regex("close"))
async def forceclose_command(_, CallbackQuery):
    try:
        await CallbackQuery.message.delete()
    except:
        return
    try:
        await CallbackQuery.answer()
    except:
        pass


@app.on_callback_query(filters.regex(pattern=r"^(resume_cb|pause_cb|skip_cb|end_cb)$"))
@admin_check_cb
async def admin_cbs(_, query: CallbackQuery):
    try:
        await query.answer()
    except:
        pass

    data = query.matches[0].group(1)

    if data == "resume_cb":
        if await is_streaming(query.message.chat.id):
            return await query.answer(
                "ᴅɪᴅ ʏᴏᴜ ʀᴇᴍᴇᴍʙᴇʀ ᴛʜᴀᴛ ʏᴏᴜ ᴘᴀᴜsᴇᴅ ᴛʜᴇ sᴛʀᴇᴀᴍ ?", show_alert=True
            )
        await stream_on(query.message.chat.id)
        await pytgcalls.resume_stream(query.message.chat.id)
        close_key = InlineKeyboardMarkup(
            [[InlineKeyboardButton(text="✯ ᴄʟᴏsᴇ ✯", callback_data="close")]]
        )
        await query.message.reply_text(
            text=f"➻ sᴛʀᴇᴀᴍ ʀᴇsᴜᴍᴇᴅ 💫\n│ \n└ʙʏ : {query.from_user.mention} 🥀",
            reply_markup=close_key,
        )

    elif data == "pause_cb":
        if not await is_streaming(query.message.chat.id):
            return await query.answer(
                "ᴅɪᴅ ʏᴏᴜ ʀᴇᴍᴇᴍʙᴇʀ ᴛʜᴀᴛ ʏᴏᴜ ʀᴇsᴜᴍᴇᴅ ᴛʜᴇ sᴛʀᴇᴀᴍ ?", show_alert=True
            )
        await stream_off(query.message.chat.id)
        await pytgcalls.pause_stream(query.message.chat.id)
        await query.message.reply_text(
            text=f"➻ sᴛʀᴇᴀᴍ ᴩᴀᴜsᴇᴅ 🥺\n│ \n└ʙʏ : {query.from_user.mention} 🥀",
            reply_markup=close_key,
        )

    elif data == "end_cb":
        try:
            await _clear_(query.message.chat.id)
            await pytgcalls.leave_group_call(query.message.chat.id)
        except:
            pass
        await query.message.reply_text(
            text=f"➻ sᴛʀᴇᴀᴍ ᴇɴᴅᴇᴅ/sᴛᴏᴩᴩᴇᴅ ❄\n│ \n└ʙʏ : {query.from_user.mention} 🥀",
            reply_markup=close_key,
        )
        await query.message.delete()

    elif data == "skip_cb":
        get = clonedb.get(query.message.chat.id)
        if not get:
            try:
                await _clear_(query.message.chat.id)
                await pytgcalls.leave_group_call(query.message.chat.id)
                await query.message.reply_text(
                    text=f"➻ sᴛʀᴇᴀᴍ sᴋɪᴩᴩᴇᴅ 🥺\n│ \n└ʙʏ : {query.from_user.mention} 🥀\n\n**» ɴᴏ ᴍᴏʀᴇ ǫᴜᴇᴜᴇᴅ ᴛʀᴀᴄᴋs ɪɴ** {query.message.chat.title}, **ʟᴇᴀᴠɪɴɢ ᴠɪᴅᴇᴏᴄʜᴀᴛ.**",
                    reply_markup=close_key,
                )
                return await query.message.delete()
            except:
                return
        else:
            title = get[0]["title"]
            duration = get[0]["duration"]
            videoid = get[0]["videoid"]
            file_path = get[0]["file_path"]
            req_by = get[0]["req"]
            user_id = get[0]["user_id"]
            get.pop(0)

            stream = MediaStream(file_path, audio_parameters=AudioQuality.HIGH)
            try:
                await pytgcalls.change_stream(
                    query.message.chat.id,
                    stream,
                )
            except Exception as ex:
                LOGGER.error(ex)
                await _clear_(query.message.chat.id)
                return await pytgcalls.leave_group_call(query.message.chat.id)

            img = await gen_thumb(videoid, user_id)
            await query.edit_message_text(
                text=f"➻ sᴛʀᴇᴀᴍ sᴋɪᴩᴩᴇᴅ 🥺\n│ \n└ʙʏ : {query.from_user.mention} 🥀",
                reply_markup=close_key,
            )
            buttons = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(text="▷", callback_data="resume_cb"),
                        InlineKeyboardButton(text="II", callback_data="pause_cb"),
                        InlineKeyboardButton(text="‣‣I", callback_data="skip_cb"),
                        InlineKeyboardButton(text="▢", callback_data="end_cb"),
                    ]
                ]
            )
        vi = await client.get_me()
        return await query.message.reply_photo(
            photo=img,
            caption=f"**➻ sᴛᴀʀᴛᴇᴅ sᴛʀᴇᴀᴍɪɴɢ**\n\n‣ **ᴛɪᴛʟᴇ :** [{title[:27]}](https://t.me/{vi.username}?start=info_{videoid})\n‣ **ᴅᴜʀᴀᴛɪᴏɴ :** `{duration}` ᴍɪɴᴜᴛᴇs\n‣ **ʀᴇǫᴜᴇsᴛᴇᴅ ʙʏ :** {req_by}",
            reply_markup=buttons,
        )


@app.on_callback_query(filters.regex("clone_help"))
async def help_menu(_, query: CallbackQuery):
    try:
        await query.answer()
    except:
        pass

    try:
        await query.edit_message_text(
            text=f"๏ ʜᴇʏ {query.from_user.first_name}, 🥀\n\nᴘʟᴇᴀsᴇ ᴄʟɪᴄᴋ ᴏɴ ᴛʜᴇ ʙᴜᴛᴛᴏɴ ʙᴇʟᴏᴡ ғᴏʀ ᴡʜɪᴄʜ ʏᴏᴜ ᴡᴀɴɴᴀ ɢᴇᴛ ʜᴇʟᴘ.",
            helpmenu=[
                [InlineKeyboardButton(text="ᴇᴠᴇʀʏᴏɴᴇ", callback_data="clone_cb help")],
                [
                    InlineKeyboardButton(text="sᴜᴅᴏ", callback_data="clone_cb sudo"),
                    InlineKeyboardButton(text="ᴏᴡɴᴇʀ", callback_data="clone_cb owner"),
                ],
                [
                    InlineKeyboardButton(text="ʙᴀᴄᴋ", callback_data="clone_home"),
                    InlineKeyboardButton(text="ᴄʟᴏsᴇ", callback_data="close"),
                ],
            ],
            reply_markup=InlineKeyboardMarkup(helpmenu),
        )
    except Exception as e:
        LOGGER.error(e)
        return


@app.on_callback_query(filters.regex("clone_cb"))
async def open_hmenu(_, query: CallbackQuery):
    callback_data = query.data.strip()
    vi = client.get_me()
    cb = callback_data.split(None, 1)[1]
    help_back = [
        [InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=SUPPORT_GROUP)],
        [
            InlineKeyboardButton(text="ʙᴀᴄᴋ", callback_data="clone_help"),
            InlineKeyboardButton(text="ᴄʟᴏsᴇ", callback_data="close"),
        ],
    ]
    keyboard = InlineKeyboardMarkup(help_back)

    try:
        await query.answer()
    except:
        pass

    if cb == "help":
        await query.edit_message_text(
            HELP_TEXT.format(vi.mention), reply_markup=keyboard
        )
    elif cb == "sudo":
        await query.edit_message_text(
            HELP_SUDO.format(vi.mention), reply_markup=keyboard
        )
    elif cb == "owner":
        await query.edit_message_text(
            HELP_DEV.format(vi.mention), reply_markup=keyboard
        )


@app.on_callback_query(filters.regex("clone_home"))
async def home_fallen(_, query: CallbackQuery):
    try:
        await query.answer()
    except:
        pass
    try:
        vi = await client.get_me()
        pm_buttons = [
            [
                InlineKeyboardButton(
                    text="ᴀᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘ",
                    url=f"https://t.me/{vi.username}?startgroup=true",
                )
            ],
            [InlineKeyboardButton(text="ʜᴇʟᴩ & ᴄᴏᴍᴍᴀɴᴅs", callback_data="fallen_help")],
            [
                InlineKeyboardButton(text="❄ ᴄʜᴀɴɴᴇʟ ❄", url=config.SUPPORT_CHANNEL),
                InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=config.SUPPORT_CHAT),
            ],
            [
                InlineKeyboardButton(text="🥀 ᴅᴇᴠᴇʟᴏᴩᴇʀ 🥀", user_id=config.OWNER_ID),
            ],
        ]

        await query.edit_message_text(
            text=PM_START_TEXT.format(
                query.from_user.first_name,
                vi.mention,
            ),
            reply_markup=InlineKeyboardMarkup(pm_buttons),
        )
    except:
        pass

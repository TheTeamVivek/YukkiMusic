from pyrogram import Client, filters
from pyrogram.types import CallbackQuery, InlineKeyboardMarkup, InlineKeyboardButton
from pyrogram.enums import ChatMemberStatus
from pytgcalls.types import MediaStream, AudioQuality

from config import *
import logging
from YukkiMusic.utils.thumbnails import gen_thumb
from .utils import (
    HELP_TEXT,
    PM_START_TEXT,
    HELP_DEV,
    HELP_SUDO,
)
from .utils.active import (
    is_active_chat,
    is_streaming,
    stream_on,
    stream_off,
)
from YukkiMusic.misc import SUDOERS
from .utils.active import _clear_
from .utils.inline import close_key
from .play import pytgcalls
from YukkiMusic.misc import clonedb


@Client.on_callback_query(filters.regex("forceclose"))
async def close_(client, CallbackQuery):
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


@Client.on_callback_query(filters.regex("close"))
async def forceclose_command(client, CallbackQuery):
    try:
        await CallbackQuery.message.delete()
    except:
        return
    try:
        await CallbackQuery.answer()
    except:
        pass


@Client.on_callback_query(
    filters.regex(pattern=r"^(resume_cb|pause_cb|skip_cb|end_cb)$")
)
async def admin_cbs(client, query: CallbackQuery):
    try:
        i = await client.get_me()
        user_id = query.from_user.id
        chat_id = query.message.chat.id
        if not await is_active_chat(chat_id, i.id):
            return await query.answer(
                "ʙᴏᴛ ɪsɴ'ᴛ sᴛʀᴇᴀᴍɪɴɢ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ.", show_alert=True
            )
        check = await client.get_chat_member(chat_id, user_id)
        if (
            check.status not in [ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR]
            or user_id not in SUDOERS
        ):
            return await query.answer("sᴏʀʀʏ? ᴏɴʟʏ ᴀᴅᴍɪɴ ᴄᴀɴ ᴅᴏ ᴛʜɪs", show_alert=True)
        try:
            await query.answer()
        except:
            pass

        data = query.matches[0].group(1)

        if data == "resume_cb":
            if await is_streaming(query.message.chat.id, i.id):
                return await query.answer(
                    "ᴅɪᴅ ʏᴏᴜ ʀᴇᴍᴇᴍʙᴇʀ ᴛʜᴀᴛ ʏᴏᴜ ᴘᴀᴜsᴇᴅ ᴛʜᴇ sᴛʀᴇᴀᴍ ?", show_alert=True
                )
            await stream_on(query.message.chat.id, i.id)
            await pytgcalls.resume_stream(query.message.chat.id)
            await query.message.reply_text(
                text=f"➻ sᴛʀᴇᴀᴍ ʀᴇsᴜᴍᴇᴅ 💫\n│ \n└ʙʏ : {query.from_user.mention} 🥀",
            )

        elif data == "pause_cb":
            if not await is_streaming(query.message.chat.id, i.id):
                return await query.answer(
                    "ᴅɪᴅ ʏᴏᴜ ʀᴇᴍᴇᴍʙᴇʀ ᴛʜᴀᴛ ʏᴏᴜ ʀᴇsᴜᴍᴇᴅ ᴛʜᴇ sᴛʀᴇᴀᴍ ?", show_alert=True
                )
            await stream_off(query.message.chat.id, i.id)
            await pytgcalls.pause_stream(query.message.chat.id)
            await query.message.reply_text(
                text=f"➻ sᴛʀᴇᴀᴍ ᴩᴀᴜsᴇᴅ 🥺 └ʙʏ : {query.from_user.mention} 🥀",
            )

        elif data == "end_cb":
            try:
                await _clear_(query.message.chat.id, i.id)
                await pytgcalls.leave_group_call(query.message.chat.id)
            except:
                pass
            await query.message.reply_text(
                text=f"➻ sᴛʀᴇᴀᴍ ᴇɴᴅᴇᴅ/sᴛᴏᴩᴩᴇᴅ ❄ └ʙʏ : {query.from_user.mention} 🥀",
            )
            await query.message.delete()

        elif data == "skip_cb":
            get = clonedb.get(query.message.chat.id, i.id)
            if not get:
                try:
                    await _clear_(query.message.chat.id, i.id)
                    await pytgcalls.leave_group_call(query.message.chat.id)
                    await query.message.reply_text(
                        text=f"➻ sᴛʀᴇᴀᴍ sᴋɪᴩᴩᴇᴅ 🥺 └ʙʏ : {query.from_user.mention} 🥀\n**» ɴᴏ ᴍᴏʀᴇ ǫᴜᴇᴜᴇᴅ ᴛʀᴀᴄᴋs ɪɴ** {query.message.chat.title}, **ʟᴇᴀᴠɪɴɢ ᴠɪᴅᴇᴏᴄʜᴀᴛ.**",
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
                    logging.exception(ex)
                    await _clear_(query.message.chat.id, i.id)
                    return await pytgcalls.leave_group_call(query.message.chat.id)

                img = await gen_thumb(videoid)
                await query.edit_message_text(
                    text=f"➻ sᴛʀᴇᴀᴍ sᴋɪᴩᴩᴇᴅ 🥺\n└ʙʏ : {query.from_user.mention} 🥀",
                )
            return await query.message.reply_photo(
                photo=img,
                caption=f"**✮ 𝐒ʈᴧʀʈ𝛆ɗ 𝐒ʈʀ𝛆ɑɱɩŋʛ ✮**\n\n**✮ 𝐓ɩttɭ𝛆 ✮** [{title[:27]}](https://t.me/{i.username}?start=info_{videoid})\n‣ **✬ 𝐃ʋɽɑʈɩσŋ ✮** `{duration}` ᴍɪɴ\n**✭ 𝐁ɣ ✮** {req_by}",
                reply_markup=close_key,
            )

    except Exception as e:
        logging.exception(e)


@Client.on_callback_query(filters.regex("clone_help"))
async def help_menu(client, query: CallbackQuery):
    try:
        await query.answer()
    except:
        pass

    try:
        helpmenu = InlineKeyboardMarkup(
            [
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
        )
        await query.edit_message_text(
            text=f"๏ ʜᴇʏ {query.from_user.mention}, 🥀\n\nᴘʟᴇᴀsᴇ ᴄʟɪᴄᴋ ᴏɴ ᴛʜᴇ ʙᴜᴛᴛᴏɴ ʙᴇʟᴏᴡ ғᴏʀ ᴡʜɪᴄʜ ʏᴏᴜ ᴡᴀɴɴᴀ ɢᴇᴛ ʜᴇʟᴘ.",
            reply_markup=helpmenu,
        )
    except Exception as e:
        logging.exception(e)
        return


@Client.on_callback_query(filters.regex("clone_cb"))
async def open_hmenu(client, query: CallbackQuery):
    callback_data = query.data.strip()
    cb = callback_data.split(None, 1)[1]
    vi = await client.get_me()
    h = vi.mention
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
        await query.edit_message_text(HELP_TEXT.format(h), reply_markup=keyboard)
    elif cb == "sudo":
        await query.edit_message_text(HELP_SUDO.format(h), reply_markup=keyboard)
    elif cb == "owner":
        await query.edit_message_text(HELP_DEV.format(h), reply_markup=keyboard)


@Client.on_callback_query(filters.regex("clone_home"))
async def home_fallen(client, query: CallbackQuery):
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
            [InlineKeyboardButton(text="ʜᴇʟᴩ & ᴄᴏᴍᴍᴀɴᴅs", callback_data="clone_help")],
            [
                InlineKeyboardButton(text="❄ ᴄʜᴀɴɴᴇʟ ❄", url=SUPPORT_CHANNEL),
                InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=SUPPORT_GROUP),
            ],
        ]

        await query.edit_message_text(
            text=PM_START_TEXT.format(
                query.from_user.first_name,
                vi.mention,
            ),
            reply_markup=InlineKeyboardMarkup(pm_buttons),
        )
    except Exception as e:
        logging.exception(e)
        return
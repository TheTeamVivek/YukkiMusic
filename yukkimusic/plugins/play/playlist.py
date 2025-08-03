#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import os
from random import randint

from pykeyboard import InlineKeyboard
from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from config import BANNED_USERS, SERVER_PLAYLIST_LIMIT
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.platforms import carbon, youtube
from yukkimusic.utils.database import (
    delete_playlist,
    get_playlist,
    get_playlist_names,
    save_playlist,
)
from yukkimusic.utils.decorators import language
from yukkimusic.utils.decorators.play import botplaylist_markup
from yukkimusic.utils.inline.playlist import (
    get_playlist_markup,
    warning_markup,
)
from yukkimusic.utils.pastebin import Yukkibin
from yukkimusic.utils.stream.stream import stream

from . import mhelp


@app.on_message(command("PLAYLIST_COMMAND") & ~BANNED_USERS)
@language
async def check_playlist(client, message: Message, _):
    _playlist = await get_playlist_names(message.from_user.id)
    if _playlist:
        get = await message.reply_text(_["playlist_2"])
    else:
        return await message.reply_text(_["playlist_3"])
    msg = _["playlist_4"]
    count = 0
    for ptlist in _playlist:
        _note = await get_playlist(message.from_user.id, ptlist)
        title = _note["title"]
        title = title.title()
        duration = _note["duration"]
        count += 1
        msg += f"\n\n{count}- {title[:70]}\n"
        msg += _["playlist_5"].format(duration)
    link = await Yukkibin(msg)
    lines = msg.count("\n")
    if lines >= 17:
        car = os.linesep.join(msg.split(os.linesep)[:17])
    else:
        car = msg
    img = await carbon.generate(car, randint(100, 10000000000))
    await get.delete()
    await message.reply_photo(img, caption=_["playlist_15"].format(link))


async def get_keyboard(_, user_id):
    keyboard = InlineKeyboard(row_width=5)
    _playlist = await get_playlist_names(user_id)
    count = len(_playlist)
    for x in _playlist:
        _note = await get_playlist(user_id, x)
        title = _note["title"]
        title = title.title()
        keyboard.row(
            InlineKeyboardButton(
                text=title,
                callback_data=f"del_playlist {x}",
            )
        )
    keyboard.row(
        InlineKeyboardButton(
            text=_["PL_B_5"],
            callback_data="delete_warning",
        ),
        InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
    )
    return keyboard, count


@app.on_message(command("DELETE_PLAYLIST_COMMAND") & filters.group & ~BANNED_USERS)
@language
async def del_group_message(client, message: Message, _):
    upl = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    text=_["PL_B_6"],
                    url=f"https://t.me/{app.username}?start=delplaylists",
                ),
            ]
        ]
    )
    await message.reply_text(_["playlist_6"], reply_markup=upl)


@app.on_message(command("DELETE_PLAYLIST_COMMAND") & filters.private & ~BANNED_USERS)
@language
async def del_plist_msg(client, message: Message, _):
    _playlist = await get_playlist_names(message.from_user.id)
    if _playlist:
        get = await message.reply_text(_["playlist_2"])
    else:
        return await message.reply_text(_["playlist_3"])
    keyboard, count = await get_keyboard(_, message.from_user.id)
    await get.edit_text(_["playlist_7"].format(count), reply_markup=keyboard)


@app.on_callback_query(filters.regex("play_playlist") & ~BANNED_USERS)
@language
async def play_playlist(client, query, _):
    callback_data = query.data.strip()
    mode = callback_data.split(None, 1)[1]
    user_id = query.from_user.id
    _playlist = await get_playlist_names(user_id)
    if not _playlist:
        try:
            return await query.answer(
                _["playlist_3"],
                show_alert=True,
            )
        except Exception:
            return
    chat_id = query.message.chat.id
    user_name = query.from_user.first_name
    await query.message.delete()
    result = []
    try:
        await query.answer()
    except Exception:
        pass
    video = True if mode == "v" else None
    mystic = await query.message.reply_text(_["play_1"])
    for vidids in _playlist:
        result.append(vidids)
    try:
        await stream(
            _,
            mystic,
            user_id,
            result,
            chat_id,
            user_name,
            query.message.chat.id,
            video,
            streamtype="playlist",
        )
    except Exception as e:
        ex_type = type(e).__name__
        err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
        return await mystic.edit_text(err)
    return await mystic.delete()


@app.on_message(command("PLAY_PLAYLIST_COMMAND") & ~BANNED_USERS & filters.group)
@language
async def play_playlist_command(client, message, _):
    mode = message.command[0][0]
    user_id = message.from_user.id
    _playlist = await get_playlist_names(user_id)
    if not _playlist:
        try:
            return await message.reply(
                _["playlist_3"],
                quote=True,
            )
        except Exception:
            return

    chat_id = message.chat.id
    user_name = message.from_user.first_name

    try:
        await message.delete()
    except Exception:
        pass

    result = []
    video = True if mode == "v" else None
    mystic = await message.reply_text(_["play_1"])

    for vidids in _playlist:
        result.append(vidids)

    try:
        await stream(
            _,
            mystic,
            user_id,
            result,
            chat_id,
            user_name,
            message.chat.id,
            video,
            streamtype="playlist",
        )
    except Exception as e:
        ex_type = type(e).__name__
        err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
        return await mystic.edit_text(err)

    return await mystic.delete()


"""
@app.on_message(filters.command(ADD_PLAYLIST_COMMAND) & ~BANNED_USERS)
@language
async def add_playlist(client, message: Message, _):
    if len(message.command) < 2:
        return await message.reply_text(_["playlist_22"])
    query = message.command[1]

    if "youtube.com/playlist" in query:
        adding = await message.reply_text(_["playlist_21"])
        try:
            results = Playlist(url)
            for video in results.videos:
                if video.get("isPlayable"):
                    video_info = {
                        "videoid": video["id"],
                        "title": video["title"],
                        "duration": video["duration"],
                    }
                    user_id = message.from_user.id
                    await save_playlist(user_id, video["id"], video_info)

        except Exception as e:
            return await message.reply_text(
                f"Looking like not a valid youtube playlist url or\nPlaylist created by YouTube Not Supported"
            )

        user_id = message.from_user.id
        await adding.delete()
        return await message.reply_text(_["playlist_20"])
    else:
        try:
            user_id = message.from_user.id
            _check = await get_playlist(user_id, videoid)
            if _check:
                try:
                    return await message.reply_text(_["playlist_8"])
                except KeyError:
                    pass

            _count = await get_playlist_names(user_id)
            count = len(_count)
            if count == SERVER_PLAYLIST_LIMIT:
                try:
                    return await message.reply_text(
                        _["playlist_9"].format(SERVER_PLAYLIST_LIMIT)
                    )
                except KeyError:
                    pass

            m = await message.reply(_["playlist_21"])
            title, duration_min, duration_sec, thumbnail, videoid = (
                await youtube.details(videoid, True)
            )
            title = (title[:50]).title()
            plist = {
                "videoid": videoid,
                "title": title,
                "duration": duration_min,
            }

            await save_playlist(user_id, videoid, plist)
            await m.delete()
            await message.reply_photo(thumbnail, caption=_["playlist_20"])

        except KeyError:
            return await message.reply_text("**Something wrong happens **")
        except Exception:
            pass
"""


@app.on_callback_query(filters.regex("remove_playlist") & ~BANNED_USERS)
@language
async def del_plist(client, query, _):
    callback_data = query.data.strip()
    videoid = callback_data.split(None, 1)[1]
    deleted = await delete_playlist(query.from_user.id, videoid)
    if deleted:
        try:
            await query.answer(_["playlist_11"], show_alert=True)
        except Exception:
            pass
    else:
        try:
            return await query.answer(_["playlist_12"], show_alert=True)
        except Exception:
            return

    return await query.edit_message_text(text=_["playlist_23"])


@app.on_callback_query(filters.regex("add_playlist") & ~BANNED_USERS)
@language
async def add_playlist(client, query, _):
    callback_data = query.data.strip()
    videoid = callback_data.split(None, 1)[1]
    user_id = query.from_user.id
    _check = await get_playlist(user_id, videoid)
    if _check:
        try:
            return await query.answer(_["playlist_8"], show_alert=True)
        except Exception:
            return
    _count = await get_playlist_names(user_id)
    count = len(_count)
    if count == SERVER_PLAYLIST_LIMIT:
        try:
            return await query.answer(
                _["playlist_9"].format(SERVER_PLAYLIST_LIMIT),
                show_alert=True,
            )
        except Exception:
            return
    (
        title,
        duration_min,
        duration_sec,
        thumbnail,
        vidid,
    ) = await youtube.details(videoid, True)
    title = (title[:50]).title()
    plist = {
        "videoid": vidid,
        "title": title,
        "duration": duration_min,
    }
    await save_playlist(user_id, videoid, plist)
    try:
        title = (title[:30]).title()
        return await query.answer(_["playlist_10"].format(title), show_alert=True)
    except Exception:
        return


@app.on_callback_query(filters.regex("del_playlist") & ~BANNED_USERS)
@language
async def del_plistcb(client, query, _):
    callback_data = query.data.strip()
    videoid = callback_data.split(None, 1)[1]
    user_id = query.from_user.id
    deleted = await delete_playlist(query.from_user.id, videoid)
    if deleted:
        try:
            await query.answer(_["playlist_11"], show_alert=True)
        except Exception:
            pass
    else:
        try:
            return await query.answer(_["playlist_12"], show_alert=True)
        except Exception:
            return
    keyboard, count = await get_keyboard(_, user_id)
    return await query.edit_message_reply_markup(reply_markup=keyboard)


@app.on_callback_query(filters.regex("delete_whole_playlist") & ~BANNED_USERS)
@language
async def del_whole_playlist(client, query, _):
    _playlist = await get_playlist_names(query.from_user.id)
    for x in _playlist:
        await query.answer(_["playlist_25"], show_alert=True)
        await delete_playlist(query.from_user.id, x)
    return await query.edit_message_text(_["playlist_13"])


@app.on_callback_query(filters.regex("get_playlist_playmode") & ~BANNED_USERS)
@language
async def get_playlist_playmode_(client, query, _):
    try:
        await query.answer()
    except Exception:
        pass
    buttons = get_playlist_markup(_)
    await query.edit_message_reply_markup(reply_markup=InlineKeyboardMarkup(buttons))


@app.on_callback_query(filters.regex("home_play") & ~BANNED_USERS)
@language
async def home_play_(client, query, _):
    pass

    try:
        await query.answer()
    except Exception:
        pass
    buttons = botplaylist_markup(_)
    return await query.edit_message_reply_markup(
        reply_markup=InlineKeyboardMarkup(buttons)
    )


@app.on_callback_query(filters.regex("delete_warning") & ~BANNED_USERS)
@language
async def delete_warning_message(client, query, _):
    try:
        await query.answer()
    except Exception:
        pass
    upl = warning_markup(_)
    return await query.edit_message_text(_["playlist_14"], reply_markup=upl)


@app.on_callback_query(filters.regex("del_back_playlist") & ~BANNED_USERS)
@language
async def del_back_playlist(client, query, _):
    user_id = query.from_user.id
    _playlist = await get_playlist_names(user_id)
    if _playlist:
        try:
            await query.answer(_["playlist_2"], show_alert=True)
        except Exception:
            pass
    else:
        try:
            return await query.answer(_["playlist_3"], show_alert=True)
        except Exception:
            return
    keyboard, count = await get_keyboard(_, user_id)
    return await query.edit_message_text(
        _["playlist_7"].format(count), reply_markup=keyboard
    )


(
    mhelp.add(
        "en",
        (
            f"<b>{pick_commands('PLAYLIST_COMMAND')}</b> - Check your whole playlist on the bot server\n"
            f"<b>{pick_commands('DELETE_PLAYLIST_COMMAND')}</b> - Delete any song from your saved playlist\n"
            f"<b>{pick_commands('PLAY_PLAYLIST_COMMAND')}</b> - Start playing your saved playlist in voice chat."
        ),
    )
    .add(
        "ar",
        (
            f"<b>{pick_commands('PLAYLIST_COMMAND')}</b> - تحقق من قائمة التشغيل الكاملة الخاصة بك على خادم البوت\n"
            f"<b>{pick_commands('DELETE_PLAYLIST_COMMAND')}</b> - احذف أي أغنية من قائمة التشغيل المحفوظة الخاصة بك\n"
            f"<b>{pick_commands('PLAY_PLAYLIST_COMMAND')}</b> - ابدأ تشغيل قائمة التشغيل المحفوظة الخاصة بك في الدردشة الصوتية."
        ),
    )
    .add(
        "as",
        (
            f"<b>{pick_commands('PLAYLIST_COMMAND')}</b> - বট ছাৰ্ভাৰত আপোনাৰ সম্পূৰ্ণ প্লেলিষ্ট চাওক\n"
            f"<b>{pick_commands('DELETE_PLAYLIST_COMMAND')}</b> - আপোনাৰ ছেভ কৰা প্লেলিষ্টৰ পৰা যিকোনো গান মচক\n"
            f"<b>{pick_commands('PLAY_PLAYLIST_COMMAND')}</b> - ভইচ চেটত আপোনাৰ ছেভ কৰা প্লেলিষ্ট প্লে আৰম্ভ কৰক।"
        ),
    )
    .add(
        "hi",
        (
            f"<b>{pick_commands('PLAYLIST_COMMAND')}</b> - बॉट सर्वर पर अपनी पूरी प्लेलिस्ट देखें\n"
            f"<b>{pick_commands('DELETE_PLAYLIST_COMMAND')}</b> - अपनी सेव की गई प्लेलिस्ट से कोई भी गाना हटाएँ\n"
            f"<b>{pick_commands('PLAY_PLAYLIST_COMMAND')}</b> - अपनी सेव की गई प्लेलिस्ट को वॉयस चैट में प्ले करना शुरू करें।"
        ),
    )
    .add(
        "ku",
        (
            f"<b>{pick_commands('PLAYLIST_COMMAND')}</b> - لیستی گۆرانییەکانت لەسەر ڕاژەی بۆت بپشکنە\n"
            f"<b>{pick_commands('DELETE_PLAYLIST_COMMAND')}</b> - هەر گۆرانییەک لە لیستی پاشەکەوتکراوەکەت بسڕەوە\n"
            f"<b>{pick_commands('PLAY_PLAYLIST_COMMAND')}</b> - دەست بکە بە پەخشکردنی لیستی پاشەکەوتکراوەکەت لە قسەکردنی دەنگ."
        ),
    )
    .add(
        "tr",
        (
            f"<b>{pick_commands('PLAYLIST_COMMAND')}</b> - Bot sunucusundaki tüm çalma listenizi kontrol edin\n"
            f"<b>{pick_commands('DELETE_PLAYLIST_COMMAND')}</b> - Kaydedilmiş çalma listenizden herhangi bir şarkıyı silin\n"
            f"<b>{pick_commands('PLAY_PLAYLIST_COMMAND')}</b> - Kaydedilmiş çalma listenizi sesli sohbette çalmaya başlayın."
        ),
    )
)

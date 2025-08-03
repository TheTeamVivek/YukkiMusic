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
import re
import traceback

import yt_dlp
from pykeyboard import InlineKeyboard
from pyrogram import enums, filters
from pyrogram.types import (
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    InputMediaAudio,
    InputMediaVideo,
    Message,
)

from config import (
    BANNED_USERS,
    SONG_DOWNLOAD_DURATION,
    SONG_DOWNLOAD_DURATION_LIMIT,
    cookies,
)
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.platforms import youtube
from yukkimusic.utils.decorators.language import language
from yukkimusic.utils.formatters import convert_bytes
from yukkimusic.utils.inline.song import song_markup

from . import mhelp


@app.on_message(command("SONG_COMMAND") & filters.group & ~BANNED_USERS)
@language
async def song_command_group(client, message: Message, _):
    upl = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    text=_["SG_B_1"],
                    url=f"https://t.me/{app.username}?start=song",
                ),
            ]
        ]
    )

    await message.reply_text(_["song_1"], reply_markup=upl)


# Song Module


@app.on_message(command("SONG_COMMAND") & filters.private & ~BANNED_USERS)
@language
async def song_command_private(client, message: Message, _):
    await message.delete()

    url = await youtube.url(message)

    if url:
        if not await youtube.exists(url):
            return await message.reply_text(_["song_5"])

        mystic = await message.reply_text(_["play_1"])

        (
            title,
            duration_min,
            duration_sec,
            thumbnail,
            vidid,
        ) = await youtube.details(url)

        if str(duration_min) == "None":
            return await mystic.edit_text(_["song_3"])

        if int(duration_sec) > SONG_DOWNLOAD_DURATION_LIMIT:
            return await mystic.edit_text(
                _["play_4"].format(SONG_DOWNLOAD_DURATION, duration_min)
            )

        buttons = song_markup(_, vidid)

        await mystic.delete()

        return await message.reply_photo(
            thumbnail,
            caption=_["song_4"].format(title),
            reply_markup=InlineKeyboardMarkup(buttons),
        )

    else:
        if len(message.command) < 2:
            return await message.reply_text(_["song_2"])

    mystic = await message.reply_text(_["play_1"])

    query = message.text.split(None, 1)[1]

    try:
        (
            title,
            duration_min,
            duration_sec,
            thumbnail,
            vidid,
        ) = await youtube.details(query)

    except Exception:
        return await mystic.edit_text(_["play_3"])

    if str(duration_min) == "None":
        return await mystic.edit_text(_["song_3"])

    if int(duration_sec) > SONG_DOWNLOAD_DURATION_LIMIT:
        return await mystic.edit_text(
            _["play_6"].format(SONG_DOWNLOAD_DURATION, duration_min)
        )

    buttons = song_markup(_, vidid)

    await mystic.delete()

    return await message.reply_photo(
        thumbnail,
        caption=_["song_4"].format(title),
        reply_markup=InlineKeyboardMarkup(buttons),
    )


@app.on_callback_query(filters.regex(pattern=r"song_back") & ~BANNED_USERS)
@language
async def songs_back_helper(client, query, _):
    callback_data = query.data.strip()

    callback_request = callback_data.split(None, 1)[1]

    stype, vidid = callback_request.split("|")

    buttons = song_markup(_, vidid)

    return await query.edit_message_reply_markup(
        reply_markup=InlineKeyboardMarkup(buttons)
    )


@app.on_callback_query(filters.regex(pattern=r"song_helper") & ~BANNED_USERS)
@language
async def song_helper_cb(client, query, _):
    callback_data = query.data.strip()

    callback_request = callback_data.split(None, 1)[1]

    stype, vidid = callback_request.split("|")

    try:
        await query.answer(_["song_6"], show_alert=True)

    except Exception:
        pass

    if stype == "audio":
        try:
            formats_available, link = await youtube.formats(vidid, True)

        except Exception:
            return await query.edit_message_text(_["song_7"])

        keyboard = InlineKeyboard()

        done = []

        for x in formats_available:
            check = x["format"]

            if "audio" in check:
                if x["filesize"] is None:
                    continue

                form = x["format_note"].title()

                if form not in done:
                    done.append(form)

                else:
                    continue

                sz = convert_bytes(x["filesize"])

                fmt = x["format_id"]

                keyboard.row(
                    InlineKeyboardButton(
                        text=f"{form} Quality Audio = {sz}",
                        callback_data=f"song_download {stype}|{fmt}|{vidid}",
                    ),
                )

        keyboard.row(
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data=f"song_back {stype}|{vidid}",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        )

        return await query.edit_message_reply_markup(reply_markup=keyboard)

    else:
        try:
            formats_available, link = await youtube.formats(vidid, True)

        except Exception as e:
            print(e)

            return await query.edit_message_text(_["song_7"])

        keyboard = InlineKeyboard()

        # AVC Formats Only

        done = [160, 133, 134, 135, 136, 137, 298, 299, 264, 304, 266]

        for x in formats_available:
            check = x["format"]

            if x["filesize"] is None:
                continue

            if int(x["format_id"]) not in done:
                continue

            sz = convert_bytes(x["filesize"])

            ap = check.split("-")[1]

            to = f"{ap} = {sz}"

            keyboard.row(
                InlineKeyboardButton(
                    text=to,
                    callback_data=f"song_download {stype}|{x['format_id']}|{vidid}",
                )
            )

        keyboard.row(
            InlineKeyboardButton(
                text=_["BACK_BUTTON"],
                callback_data=f"song_back {stype}|{vidid}",
            ),
            InlineKeyboardButton(text=_["CLOSE_BUTTON"], callback_data="close"),
        )

        return await query.edit_message_reply_markup(reply_markup=keyboard)


# Downloading Songs Here


@app.on_callback_query(filters.regex(pattern=r"song_download") & ~BANNED_USERS)
@language
async def song_download_cb(client, query, _):
    try:
        await query.answer("ᴅᴏᴡɴʟᴏᴀᴅɪɴɢ...")

    except Exception:
        pass

    callback_data = query.data.strip()

    callback_request = callback_data.split(None, 1)[1]

    stype, format_id, vidid = callback_request.split("|")

    mystic = await query.edit_message_text(_["song_8"])

    yturl = f"https://www.youtube.com/watch?v={vidid}"

    with yt_dlp.YoutubeDL({"quiet": True, "cookiefile": f"{cookies()}"}) as ytdl:
        x = ytdl.extract_info(yturl, download=False)

    title = (x["title"]).title()

    title = re.sub(r"\W+", " ", title)

    thumb_image_path = await query.message.download()

    duration = x["duration"]

    if stype == "video":
        thumb_image_path = await query.message.download()

        width = query.message.photo.width

        height = query.message.photo.height

        try:
            file_path = await youtube.download(
                yturl,
                mystic,
                songvideo=True,
                format_id=format_id,
                title=title,
            )

        except Exception as e:
            return await mystic.edit_text(_["song_9"].format(e))

        med = InputMediaVideo(
            media=file_path,
            duration=duration,
            width=width,
            height=height,
            thumb=thumb_image_path,
            caption=title,
            supports_streaming=True,
        )

        await mystic.edit_text(_["song_11"])

        await app.send_chat_action(
            chat_id=query.message.chat.id,
            action=enums.ChatAction.UPLOAD_VIDEO,
        )

        try:
            await query.edit_message_media(media=med)

        except Exception:
            traceback.print_exc()

            return await mystic.edit_text(_["song_10"])

        os.remove(file_path)

    elif stype == "audio":
        try:
            filename = await youtube.download(
                yturl,
                mystic,
                songaudio=True,
                format_id=format_id,
                title=title,
            )

        except Exception as e:
            return await mystic.edit_text(_["song_9"].format(e))

        med = InputMediaAudio(
            media=filename,
            caption=title,
            thumb=thumb_image_path,
            title=title,
            performer=x["uploader"],
        )

        await mystic.edit_text(_["song_11"])

        await app.send_chat_action(
            chat_id=query.message.chat.id,
            action=enums.ChatAction.UPLOAD_AUDIO,
        )

        try:
            await query.edit_message_media(media=med)

        except Exception:
            traceback.print_exc()

            return await mystic.edit_text(_["song_10"])

        os.remove(filename)


(
    mhelp.add(
        "en",
        f"<b>★ {pick_commands('SONG_COMMAND')}</b> [Track Name] or [YT Link] - Download any track from YouTube in MP3 or MP4 formats.",
    )
    .add(
        "ar",
        f"<b>★ {pick_commands('SONG_COMMAND')}</b> [اسم المقطع] أو [رابط يوتيوب] - قم بتنزيل أي مقطع من يوتيوب بصيغة MP3 أو MP4.",
    )
    .add(
        "as",
        f"<b>★ {pick_commands('SONG_COMMAND')}</b> [গীতৰ নাম] বা [YT লিংক] - যিকোনো গান YouTube ৰ পৰা MP3 বা MP4 ফৰ্মেটত ডাউনল'ড কৰক।",
    )
    .add(
        "hi",
        f"<b>★ {pick_commands('SONG_COMMAND')}</b> [गाने का नाम] या [YT लिंक] - YouTube से कोई भी गाना MP3 या MP4 फ़ॉर्मेट में डाउनलोड करें।",
    )
    .add(
        "ku",
        f"<b>★ {pick_commands('SONG_COMMAND')}</b> [ناوی گۆرانی] یان [لینکی YT] - هەر گۆرانییەک لە YouTube دا بە فۆرماتی MP3 یان MP4 داگرتن بکە.",
    )
    .add(
        "tr",
        f"<b>★ {pick_commands('SONG_COMMAND')}</b> [Şarkı Adı] veya [YT Link] - YouTube'dan herhangi bir şarkıyı MP3 veya MP4 formatında indirin.",
    )
)

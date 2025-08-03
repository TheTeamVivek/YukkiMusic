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
from pyrogram.types import Message
from pytgcalls.exceptions import NoActiveGroupCall

import config
from config import BANNED_USERS
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.core.call import yukki
from yukkimusic.utils.decorators.play import PlayWrapper
from yukkimusic.utils.logger import play_logs
from yukkimusic.utils.stream.stream import stream

from . import mhelp


@app.on_message(command("STREAM_COMMAND") & filters.group & ~BANNED_USERS)
@PlayWrapper
async def stream_command(
    client,
    message: Message,
    _,
    chat_id,
    video,
    channel,
    playmode,
    url,
    fplay,
):
    if url:
        mystic = await message.reply_text(
            _["play_2"].format(channel) if channel else _["play_1"]
        )
        try:
            await yukki.stream_call(url)
        except NoActiveGroupCall:
            await mystic.edit_text(
                "There's an issue with the bot. please report it to my Owner and ask them to check logger group"
            )
            text = "Please Turn on voice chat.. Bot is unable to stream urls.."
            return await app.send_message(config.LOG_GROUP_ID, text)
        except Exception as e:
            return await mystic.edit_text(_["general_3"].format(type(e).__name__))
        await mystic.edit_text(_["str_2"])
        try:
            await stream(
                _,
                mystic,
                message.from_user.id,
                url,
                chat_id,
                message.from_user.first_name,
                message.chat.id,
                video=True,
                streamtype="index",
            )
        except Exception as e:
            ex_type = type(e).__name__
            err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
            return await mystic.edit_text(err)
        return await play_logs(message, streamtype="M3u8 or Index Link")
    else:
        await message.reply_text(_["str_1"])


(
    mhelp.add(
        "en",
        f"<b>✧ {pick_commands('STREAM_COMMAND')}</b> - Stream a URL that you believe is direct or m3u8 that can't be played by play.",
    )
    .add(
        "ar",
        f"<b>✧ {pick_commands('STREAM_COMMAND')}</b> - قم ببث رابط تعتقد أنه مباشر أو m3u8 ولا يمكن تشغيله باستخدام تشغيل.",
    )
    .add(
        "as",
        f"<b>✧ {pick_commands('STREAM_COMMAND')}</b> - এটা URL স্ট্ৰিম কৰক যাক আপুনি প্ৰত্যক্ষ বা m3u8 বুলি বিশ্বাস কৰে আৰু যাক প্লে ৰ দ্বাৰা প্লে কৰা নাযায়।",
    )
    .add(
        "hi",
        f"<b>✧ {pick_commands('STREAM_COMMAND')}</b> - ऐसा URL स्ट्रीम करें जो सीधा हो या m3u8 हो जिसे प्ले से प्ले नहीं किया जा सकता।",
    )
    .add(
        "ku",
        f"<b>✧ {pick_commands('STREAM_COMMAND')}</b> - بەستەرێک بڵاو بکەرەوە کە باوەڕت پێیە سەردەشتە یان m3u8 ـە کە ناتوانرێت بە پەخشکردن بلیژرێت.",
    )
    .add(
        "tr",
        f"<b>✧ {pick_commands('STREAM_COMMAND')}</b> - play ile çalınamayan doğrudan veya m3u8 olduğunu düşündüğünüz bir URL'yi yayınlayın.",
    )
)

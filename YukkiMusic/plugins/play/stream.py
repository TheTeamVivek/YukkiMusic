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
from strings import command
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.decorators.play import play_wrapper
from YukkiMusic.utils.logger import play_logs
from YukkiMusic.utils.stream.stream import stream


@app.on_message(command("STREAM_COMMAND") & filters.group & ~BANNED_USERS)
@play_wrapper
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
        mystic = await event.reply(
            _["play_2"].format(channel) if channel else _["play_1"]
        )
        try:
            await Yukki.stream_call(url)
        except NoActiveGroupCall:
            await mystic.edit(
                "There's an issue with the bot. please report it to my Owner and ask them to check logger group"
            )
            text = "Please Turn on voice chat.. Bot is unable to stream urls.."
            return await app.send_message(config.LOG_GROUP_ID, text)
        except Exception as e:
            return await mystic.edit(_["ERROR_OCCURRED_MSG"].format(type(e).__name__))
        await mystic.edit(_["str_2"])
        try:
            await stream(
                _,
                mystic,
                event.sender_id,
                url,
                chat_id,
                message.from_user.first_name,
                event.chat_id,
                video=True,
                streamtype="index",
            )
        except Exception as e:
            ex_type = type(e).__name__
            err = (
                e
                if ex_type == "AssistantErr"
                else _["ERROR_OCCURRED_MSG"].format(ex_type)
            )
            return await mystic.edit(err)
        return await play_logs(message, streamtype="M3u8 or Index Link")
    else:
        await event.reply(_["str_1"])

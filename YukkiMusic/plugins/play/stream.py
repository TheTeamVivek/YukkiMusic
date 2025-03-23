#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import logging
from pyrogram import filters
from pyrogram.types import Message
from pytgcalls.exceptions import NoActiveGroupCall

import config
from config import BANNED_USERS
from strings import command
from YukkiMusic import app, tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.decorators.play import play_wrapper
from YukkiMusic.utils.logger import play_logs
from YukkiMusic.utils.stream.stream import stream

from YukkiMusic.core.youtube import Track
from YukkiMusic.core.enum import SourceType

logger = logging.getLogger(__name__)

@tbot.on_message(
    flt.command("STREAM_COMMAND", True) & flt.group & ~BANNED_USERS)
)
@play_wrapper
async def stream_command(
    event,
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
        track = Track(
            title="M3U8 or index Urls",
            link=url,
            thumb=config.STREAM_IMG_URL,
            duration=0,
            streamtype=SourceType.M3U8,
            video=video,
        )
        try:
            await stream(
                chat_id=chat_id,
                original_chat_id=event.chat_id,
                track=track,
                user_id=event.sender_id,
                forceplay=fplay,
            )
        except Exception as e:
            ex_type = type(e).__name__
            if ex_type == "AssistantErr":
                err = e
            else:
                err = _["general_3"].format(ex_type)
                logger.error("\n", exc_info=True)
            return await mystic.edit(err)
        return await play_logs(event, streamtype=SourceType.M3U8)
    else:
        await event.reply(_["str_1"])

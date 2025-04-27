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

import config
from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.core import SourceType, Track, filters
from YukkiMusic.utils import play_logs, play_wrapper, stream

logger = logging.getLogger(__name__)


@tbot.on_message(
    filters.command("STREAM_COMMAND", True) & filters.group & ~BANNED_USERS
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

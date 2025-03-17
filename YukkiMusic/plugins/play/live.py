#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from telethon import events

from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.platforms import youtube
from YukkiMusic.utils.channelplay import get_channeplay_cb
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.stream.stream import stream


@tbot.on(events.CallbackQuery("LiveStream", func=~BANNED_USERS))
@language
async def play_live_stream(event, _):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    vidid, user_id, mode, cplay, fplay = callback_request.split("|")
    if event.sender_id != int(user_id):
        try:
            return await event.answer(_["playcb_1"], alert=True)
        except Exception:
            return
    try:
        chat_id, channel = await get_channeplay_cb(_, cplay, event)
    except Exception:
        return
        
    await event.delete()
    try:
        await event.answer()
    except Exception:
        pass
    mystic = await event.reply(
        _["play_2"].format(channel) if channel else _["play_1"]
    )
    try:
        url = youtube.base + vidid
        track = await youtube.track(url, mode == "v")
    except Exception:
        return await mystic.edit(_["play_3"])
    if track.is_live:
        try:
            await stream(
                chat_id=chat_id,
                original_chat_id=event.chat_id,
                track=track,
                user_id=int(user_id),
                forceplay=fplay == "f",
            )
        except Exception as e:
            ex_type = type(e).__name__
            err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
            await tbot.handle_error(e)
            return await mystic.edit(err)
    else:
        return await mystic.edit("Not a live stream")
    await mystic.delete()

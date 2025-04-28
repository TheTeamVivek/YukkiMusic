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

from telethon import events

from YukkiMusic.misc import BANNED_USERS
from YukkiMusic.utils import (
    asyncify,
    botplaylist_markup,
    failed_top_markup,
    get_global_tops,
    get_particulars,
    get_userss,
    language,
    stream,
    top_play_markup,
)

logger = logging.getLogger(__name__)


@tbot.on(events.CallbackQuery("get_playmarkup", func=~BANNED_USERS))
@tbot.on(events.CallbackQuery("get_top_playlists", func=~BANNED_USERS))
@language
async def get_play_markup(event, _):
    try:
        await event.answer()
    except Exception:
        pass
    data = event.data.decode("utf-8")

    if data.startswith("get_playmarkup"):
        buttons = botplaylist_markup(_)
    elif data.startswith("get_top_playlists"):
        buttons = top_play_markup(_)

    return await event.edit(buttons=buttons)


@tbot.on(events.CallbackQuery("SERVERTOP", func=~BANNED_USERS))
@language
async def server_to_play(event, _):
    chat_id = event.chat_id
    user_id = event.sender_id
    user_name = (await event.get_sender).first_name
    try:
        await event.answer()
    except Exception:
        pass
    callback_data = event.data.decode("utf-8").strip()
    what = callback_data.split(None, 1)[1]
    mystic = await event.edit(
        _["tracks_1"].format(
            what,
            user_name,
        )
    )
    upl = failed_top_markup(_)
    if what == "Global":
        stats = await get_global_tops()
    elif what == "Group":
        stats = await get_particulars(chat_id)
    elif what == "Personal":
        stats = await get_userss(user_id)
    if not stats:
        return await mystic.edit(_["tracks_2"].format(what), buttons=upl)

    @asyncify
    def get_stats():
        results = {}
        for i in stats:
            top_list = stats[i]["spot"]
            results[str(i)] = top_list
            list_arranged = dict(
                sorted(
                    results.items(),
                    key=lambda item: item[1],
                    reverse=True,
                )
            )
        if not results:
            return mystic.edit(_["tracks_2"].format(what), buttons=upl)
        details = []
        limit = 0
        for vidid, count in list_arranged.items():
            if vidid == "telegram":
                continue
            if limit == 10:
                break
            limit += 1
            details.append(vidid)
        if not details:
            return mystic.edit(_["tracks_2"].format(what), buttons=upl)
        return details

    try:
        details = await get_stats()
    except Exception:
        logger.error("", exc_info=True)
        return
    try:
        await stream(
            chat_id=chat_id,
            original_chat_id=chat_id,
            track=details,  # TODO: fix it
            user_id=user_id,
        )
    except Exception as e:
        ex_type = type(e).__name__
        if ex_type == "AssistantErr":
            err = e
        else:
            err = _["general_3"].format(ex_type)
        logger.error("\n", exc_info=True)

        return await mystic.edit(err)
    return await mystic.delete()

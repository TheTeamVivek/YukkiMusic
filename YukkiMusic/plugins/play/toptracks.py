#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio

from pyrogram import filters
from pyrogram.types import InlineKeyboardMarkup

from config import BANNED_USERS
from YukkiMusic import app
from YukkiMusic.utils.database import (
    get_global_tops,
    get_particulars,
    get_userss,
)
from YukkiMusic.utils.decorators import asyncify, language
from YukkiMusic.utils.inline.playlist import (
    botplaylist_markup,
    failed_top_markup,
    top_play_markup,
)
from YukkiMusic.utils.stream.stream import stream

loop = asyncio.get_running_loop()


@app.on_callback_query(filters.regex("get_playmarkup") & ~BANNED_USERS)
@language
async def get_play_markup(client, query, _):
    try:
        await query.answer()
    except Exception:
        pass
    buttons = botplaylist_markup(_)
    return await query.edit_message_reply_markup(
        reply_markup=InlineKeyboardMarkup(buttons)
    )


@app.on_callback_query(filters.regex("get_top_playlists") & ~BANNED_USERS)
@language
async def get_topz_playlists(client, query, _):
    try:
        await query.answer()
    except Exception:
        pass
    buttons = top_play_markup(_)
    return await query.edit_message_reply_markup(
        reply_markup=InlineKeyboardMarkup(buttons)
    )


@app.on_callback_query(filters.regex("SERVERTOP") & ~BANNED_USERS)
@language
async def server_to_play(client, query, _):
    chat_id = query.message.chat.id
    user_name = query.from_user.first_name
    try:
        await query.answer()
    except Exception:
        pass
    callback_data = query.data.strip()
    what = callback_data.split(None, 1)[1]
    mystic = await query.edit_message_text(
        _["tracks_1"].format(
            what,
            query.from_user.first_name,
        )
    )
    upl = failed_top_markup(_)
    if what == "Global":
        stats = await get_global_tops()
    elif what == "Group":
        stats = await get_particulars(chat_id)
    elif what == "Personal":
        stats = await get_userss(query.from_user.id)
    if not stats:
        return await mystic.edit(_["tracks_2"].format(what), reply_markup=upl)

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
            return mystic.edit(_["tracks_2"].format(what), reply_markup=upl)
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
            return mystic.edit(_["tracks_2"].format(what), reply_markup=upl)
        return details

    try:
        details = await get_stats()
    except Exception as e:
        print(e)
        return
    try:
        await stream(
            _,
            mystic,
            query.from_user.id,
            details,
            chat_id,
            user_name,
            query.message.chat.id,
            video=False,
            streamtype="playlist",
        )
    except Exception as e:
        ex_type = type(e).__name__
        err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
        return await mystic.edit_text(err)
    return await mystic.delete()

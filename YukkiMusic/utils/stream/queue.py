#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from config import autoclean, chatstats, time_to_seconds, userstats
from YukkiMusic.core.youtube import Track 
from YukkiMusic.misc import db


async def put_queue(
    chat_id,
    original_chat_id,
    user_id,
    track: Track,
    video: bool = False,
    is_index: bool = False,
    forceplay: bool | None = None,
):

    track.title = track.title.title()
    put = {
        "chat_id": original_chat_id,
        "track": track,
        "played": 0,
    }
    if forceplay:
        if check := db.get(chat_id):
            check.insert(0, put)
        else:
            db[chat_id] = []
            db[chat_id].append(put)
    else:
        db[chat_id].append(put)
    autoclean.append(file)

    if is_index is None:
        vidid = track.vidid if track.is_youtube else track.vidid.value

        to_append = {"vidid": vidid, "title": title}
        if chat_id not in chatstats:
            chatstats[chat_id] = []
        chatstats[chat_id].append(to_append)
        if user_id not in userstats:
            userstats[user_id] = []
        userstats[user_id].append(to_append)
        return

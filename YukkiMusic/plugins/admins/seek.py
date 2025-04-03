#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from YukkiMusic.core import filters as flt
from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import db
from YukkiMusic.platforms import youtube
from YukkiMusic.utils import admin_rights_check, seconds_to_min


@tbot.on_message(
    flt.command(["SEEK_COMMAND", "SEEK_BACK_COMMAND"], True)
    & flt.group
    & ~flt.user(BANNED_USERS)
)
@admin_rights_check
async def seek_comm(event, _, chat_id):
    comm = event.text.split()
    if len(comm) == 1:
        return await event.reply(_["admin_28"])
    query = event.text.split(None, 1)[1].strip()
    if not query.isnumeric():
        return await event.reply(_["admin_29"])
    playing = db.get(chat_id)
    if not playing:
        return await event.reply(_["queue_2"])
    duration_seconds = int(playing[0]["seconds"])
    if duration_seconds == 0:
        return await event.reply(_["admin_30"])
    file_path = playing[0]["file"]
    if "index_" in file_path or "live_" in file_path:
        return await event.reply(_["admin_30"])
    duration_played = int(playing[0]["played"])
    duration_to_skip = int(query)
    duration = playing[0]["dur"]
    if comm[0][-2] == "c":
        if (duration_played - duration_to_skip) <= 10:
            return await event.reply(
                _["admin_31"].format(seconds_to_min(duration_played), duration)
            )
        to_seek = duration_played - duration_to_skip + 1
    else:
        if (duration_seconds - (duration_played + duration_to_skip)) <= 10:
            return await event.reply(
                _["admin_31"].format(seconds_to_min(duration_played), duration)
            )
        to_seek = duration_played + duration_to_skip + 1
    mystic = await event.reply(_["admin_32"])
    if "vid_" in file_path:
        n, file_path = await youtube.video(playing[0]["vidid"], True)
        if n == 0:
            return await event.reply(_["admin_30"])
    try:
        await Yukki.seek_stream(
            chat_id,
            file_path,
            seconds_to_min(to_seek),
            duration,
            playing[0]["streamtype"],
        )
    except Exception:
        return await mystic.edit(_["admin_34"])
    if message.command[0][-2] == "c": #TODO: replace with patse_flags
        db[chat_id][0]["played"] -= duration_to_skip
    else:
        db[chat_id][0]["played"] += duration_to_skip
    await mystic.edit(_["admin_33"].format(seconds_to_min(to_seek)))

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from strings import get_command
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import BANNED_USERS, db
from YukkiMusic.utils import admin_rights_check, seconds_to_min

SEEK_COMMAND = get_command("SEEK_COMMAND")
SEEK_BACK_COMMAND = get_command("SEEK_BACK_COMMAND")


@tbot.on_message(
    flt.command(SEEK_COMMAND + SEEK_BACK_COMMAND) & flt.group & ~BANNED_USERS
)
@admin_rights_check
async def seek_comm(event, _, chat_id):
    text = event.text
    comm = text.lstrip("/").lower.split()
    is_seekback = any(comm[0] == key for key in SEEK_BACK_COMMAND)
    if len(comm) == 1:
        return await event.reply(_["admin_29"])

    query = text.split(None, 1)[1].strip()

    if not query.isnumeric():
        return await event.reply(_["admin_30"])

    playing = db.get(chat_id)

    if not playing:
        return await event.reply(_["queue_2"])

    duration_seconds = int(playing[0]["seconds"])

    if duration_seconds == 0:
        return await event.reply(_["admin_31"])

    track = playing[0]["track"]

    if track.is_m3u8 or track.is_live:
        return await event.reply(_["admin_31"])

    duration_played = int(playing[0]["played"])
    duration_to_skip = int(query)
    duration = playing[0]["dur"]

    if is_seekback:
        if (duration_played - duration_to_skip) <= 10:
            return await event.reply(
                _["admin_32"].format(seconds_to_min(duration_played), duration)
            )
        to_seek = duration_played - duration_to_skip + 1

    else:
        if (duration_seconds - (duration_played + duration_to_skip)) <= 10:
            return await event.reply(
                _["admin_32"].format(seconds_to_min(duration_played), duration)
            )
        to_seek = duration_played + duration_to_skip + 1

    mystic = await event.reply(_["admin_33"])

    try:
        file_path = await track.download()
    except Exception as e:
        await mystic.edit(_["admin_31"])
        return await tbot.handle_error(e, event)

    try:
        await Yukki.seek_stream(
            chat_id,
            file_path,
            seconds_to_min(to_seek),
            duration,
            playing[0]["streamtype"],
        )
    except Exception:
        return await mystic.edit(_["admin_35"])

    if is_seekback:
        db[chat_id][0]["played"] -= duration_to_skip
    else:
        db[chat_id][0]["played"] += duration_to_skip
    await mystic.edit(_["admin_34"].format(seconds_to_min(to_seek)))

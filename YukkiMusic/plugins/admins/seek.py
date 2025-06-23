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

from config import BANNED_USERS
from strings import command, pick_commands
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import db
from YukkiMusic.platforms import youtube
from YukkiMusic.utils import AdminRightsCheck, seconds_to_min

from . import mhelp


@app.on_message(
    command(["SEEK_COMMAND", "SEEK_BACK_COMMAND"]) & filters.group & ~BANNED_USERS
)
@AdminRightsCheck
async def seek_comm(cli, message: Message, _, chat_id):
    if len(message.command) == 1:
        return await message.reply_text(_["seek_1"])
    query = message.text.split(None, 1)[1].strip()
    if not query.isnumeric():
        return await message.reply_text(_["seek_2"])
    playing = db.get(chat_id)
    if not playing:
        return await message.reply_text(_["queue_2"])
    duration_seconds = int(playing[0]["seconds"])
    if duration_seconds == 0:
        return await message.reply_text(_["seek_3"])
    file_path = playing[0]["file"]
    if "index_" in file_path or "live_" in file_path:
        return await message.reply_text(_["seek_3"])
    duration_played = int(playing[0]["played"])
    duration_to_skip = int(query)
    duration = playing[0]["dur"]
    if message.command[0][-2] == "c":
        if (duration_played - duration_to_skip) <= 10:
            return await message.reply_text(
                _["seek_4"].format(seconds_to_min(duration_played), duration)
            )
        to_seek = duration_played - duration_to_skip + 1
    else:
        if (duration_seconds - (duration_played + duration_to_skip)) <= 10:
            return await message.reply_text(
                _["seek_4"].format(seconds_to_min(duration_played), duration)
            )
        to_seek = duration_played + duration_to_skip + 1
    mystic = await message.reply_text(_["seek_5"])
    if "vid_" in file_path:
        n, file_path = await youtube.video(playing[0]["vidid"], True)
        if n == 0:
            return await message.reply_text(_["seek_3"])
    try:
        await Yukki.seek_stream(
            chat_id,
            file_path,
            seconds_to_min(to_seek),
            duration,
            playing[0]["streamtype"],
        )
    except Exception:
        return await mystic.edit_text(_["seek_7"])
    if message.command[0][-2] == "c":
        db[chat_id][0]["played"] -= duration_to_skip
    else:
        db[chat_id][0]["played"] += duration_to_skip
    await mystic.edit_text(_["seek_6"].format(seconds_to_min(to_seek)))


(
    mhelp.add(
        "en",
        f"<b>✧ {pick_commands('SEEK_COMMAND', 'en')}</b> - Forward seek the current track.\n"
        f"<b>✧ {pick_commands('SEEK_BACK_COMMAND', 'en')}</b> - Rewind the current track to a previous point.",
        priority=14,
    )
    .add(
        "ar",
        f"<b>✧ {pick_commands('SEEK_COMMAND', 'ar')}</b> - تقديم المسار الحالي.\n"
        f"<b>✧ {pick_commands('SEEK_BACK_COMMAND', 'ar')}</b> - إرجاع المسار الحالي إلى نقطة سابقة.",
        priority=14,
    )
    .add(
        "as",
        f"<b>✧ {pick_commands('SEEK_COMMAND', 'as')}</b> - বৰ্তমান সংগীত আগবঢ়াওক।\n"
        f"<b>✧ {pick_commands('SEEK_BACK_COMMAND', 'as')}</b> - সংগীতক পুৰণি সময়লৈ পিচলৈ আনক।",
        priority=14,
    )
    .add(
        "hi",
        f"<b>✧ {pick_commands('SEEK_COMMAND', 'hi')}</b> - वर्तमान ट्रैक को आगे बढ़ाएं।\n"
        f"<b>✧ {pick_commands('SEEK_BACK_COMMAND', 'hi')}</b> - ट्रैक को पहले की स्थिति पर पीछे करें।",
        priority=14,
    )
    .add(
        "ku",
        f"<b>✧ {pick_commands('SEEK_COMMAND', 'ku')}</b> - گۆرانییەکە بەرامبەر بگەڕێنەوە.\n"
        f"<b>✧ {pick_commands('SEEK_BACK_COMMAND', 'ku')}</b> - گۆرانییەکە بگەڕێنەوە بۆ خاڵێکی پێشوو.",
        priority=14,
    )
    .add(
        "tr",
        f"<b>✧ {pick_commands('SEEK_COMMAND', 'tr')}</b> - Çalan şarkıyı ileri sar.\n"
        f"<b>✧ {pick_commands('SEEK_BACK_COMMAND', 'tr')}</b> - Şarkıyı geri sar.",
        priority=14,
    )
)

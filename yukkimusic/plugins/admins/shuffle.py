#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#

import random

from pyrogram import filters
from pyrogram.types import Message

from config import BANNED_USERS
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import db
from yukkimusic.utils.decorators import AdminRightsCheck

from . import mhelp


@app.on_message(command("SHUFFLE_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def admins(Client, message: Message, _, chat_id):
    if not len(message.command) == 1:
        return await message.reply_text(_["general_2"])
    check = db.get(chat_id)
    if not check:
        return await message.reply_text(_["shuffle_1"])
    try:
        popped = check.pop(0)
    except Exception:
        return await message.reply_text(_["shuffle_2"])
    check = db.get(chat_id)
    if not check:
        check.insert(0, popped)
        return await message.reply_text(_["shuffle_2"])
    random.shuffle(check)
    check.insert(0, popped)
    await message.reply_text(_["shuffle_3"].format(message.from_user.mention))


(
    mhelp.add(
        "en",
        f"<b>✧ {pick_commands('SHUFFLE_COMMAND', 'en')}</b> - Randomly shuffle the queued playlist or songs.",
        priority=15,
    )
    .add(
        "ar",
        f"<b>✧ {pick_commands('SHUFFLE_COMMAND', 'ar')}</b> - خلط قائمة التشغيل أو الأغاني المنتظرة عشوائيًا.",
        priority=15,
    )
    .add(
        "as",
        f"<b>✧ {pick_commands('SHUFFLE_COMMAND', 'as')}</b> - কিউত থকা প্লেলিষ্ট বা গীতবোৰ এলোমেলো কৰক।",
        priority=15,
    )
    .add(
        "hi",
        f"<b>✧ {pick_commands('SHUFFLE_COMMAND', 'hi')}</b> - कतार में प्लेलिस्ट या गानों को यादृच्छिक रूप से फेरबदल करें।",
        priority=15,
    )
    .add(
        "ku",
        f"<b>✧ {pick_commands('SHUFFLE_COMMAND', 'ku')}</b> - لیستی گۆرانیەکان بە شێوەیەکی خاوەن بەخت داگرێنەوە.",
        priority=15,
    )
    .add(
        "tr",
        f"<b>✧ {pick_commands('SHUFFLE_COMMAND', 'tr')}</b> - Kuyruktaki çalma listesini veya şarkıları rastgele karıştır.",
        priority=15,
    )
)

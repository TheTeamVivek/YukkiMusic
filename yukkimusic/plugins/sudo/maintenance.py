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

import config
from strings import command, get_string, pick_commands
from yukkimusic import app
from yukkimusic.utils.database import (
    get_lang,
    is_maintenance,
    maintenance_off,
    maintenance_on,
)

from . import ohelp


@app.on_message(command("MAINTENANCE_COMMAND") & filters.user(config.OWNER_ID))
async def maintenance(client, message: Message):
    try:
        language = await get_lang(message.chat.id)
        _ = get_string(language)
    except Exception:
        _ = get_string("en")
    usage = _["maint_1"]
    if len(message.command) != 2:
        return await message.reply_text(usage)
    message.chat.id
    state = message.text.split(None, 1)[1].strip()
    state = state.lower()
    if state == "enable":
        if await is_maintenance():
            await message.reply_text(_["maint_6"])
        else:
            await maintenance_on()
            await message.reply_text(_["maint_2"])
    elif state == "disable":
        if await is_maintenance():
            await maintenance_off()
            await message.reply_text(_["maint_3"])
        else:
            await message.reply_text(_["maint_5"])
    else:
        await message.reply_text(usage)


(
    ohelp.add(
        "en",
        f"<b>{pick_commands('MAINTENANCE_COMMAND')}</b> [enable / disable] - Toggle bot maintenance mode",
    )
    .add(
        "ar",
        f"<b>{pick_commands('MAINTENANCE_COMMAND')}</b> [تفعيل / تعطيل] - تبديل وضع صيانة البوت",
    )
    .add(
        "as",
        f"<b>{pick_commands('MAINTENANCE_COMMAND')}</b> [সক্ৰিয় কৰক / নিষ্ক্ৰিয় কৰক] - বটৰ মেইনটেনেন্স অৱস্থা টগল কৰক",
    )
    .add(
        "hi",
        f"<b>{pick_commands('MAINTENANCE_COMMAND')}</b> [सक्रिय / निष्क्रिय] - बॉट रखरखाव मोड टॉगल करें",
    )
    .add(
        "ku",
        f"<b>{pick_commands('MAINTENANCE_COMMAND')}</b> [چالاک / ناچالاک] - دۆخی چاکسازی بۆت بگۆڕە",
    )
    .add(
        "tr",
        f"<b>{pick_commands('MAINTENANCE_COMMAND')}</b> [etkinleştir / devre dışı bırak] - Bot bakım modunu değiştir",
    )
)

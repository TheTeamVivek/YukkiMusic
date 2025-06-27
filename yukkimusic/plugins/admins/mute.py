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
from yukkimusic import app
from yukkimusic.core.call import ModuleHelp, yukki
from yukkimusic.core.help import ModuleHelp
from yukkimusic.utils.database import is_muted, mute_off, mute_on
from yukkimusic.utils.decorators import AdminRightsCheck


@app.on_message(command("MUTE_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def mute_admin(cli, message: Message, _, chat_id):
    if not len(message.command) == 1 or message.reply_to_message:
        return
    if await is_muted(chat_id):
        return await message.reply_text(_["mute_1"])
    await mute_on(chat_id)
    await yukki.mute_stream(chat_id)
    await message.reply_text(_["mute_2"].format(message.from_user.mention))


@app.on_message(command("UNMUTE_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def unmute_admin(Client, message: Message, _, chat_id):
    if not len(message.command) == 1 or message.reply_to_message:
        return
    if not await is_muted(chat_id):
        return await message.reply_text(
            _["unmute_1"],
        )
    await mute_off(chat_id)
    await yukki.unmute_stream(chat_id)
    await message.reply_text(
        _["unmute_2"].format(message.from_user.mention),
    )


(
    ModuleHelp("Admins")
    .name("en", "Admins")
    .add(
        "en",
        f"<b>✧ {pick_commands('MUTE_COMMAND', 'en')}</b> - Mute the playing music.\n"
        f"<b>✧ {pick_commands('UNMUTE_COMMAND', 'en')}</b> - Unmute the muted music.",
        priority=19,
    )
    .name("ar", "المسؤولين")
    .add(
        "ar",
        f"<b>✧ {pick_commands('MUTE_COMMAND', 'ar')}</b> - كتم الموسيقى الجارية.\n"
        f"<b>✧ {pick_commands('UNMUTE_COMMAND', 'ar')}</b> - إلغاء كتم الموسيقى المكتومة.",
        priority=19,
    )
    .name("as", "প্ৰশাসক")
    .add(
        "as",
        f"<b>✧ {pick_commands('MUTE_COMMAND', 'as')}</b> - বজাই থকা সংগীত চুপ কৰক।\n"
        f"<b>✧ {pick_commands('UNMUTE_COMMAND', 'as')}</b> - চুপ কৰা সংগীতটো শব্দসহ কৰক।",
        priority=19,
    )
    .name("hi", "प्रशासक")
    .add(
        "hi",
        f"<b>✧ {pick_commands('MUTE_COMMAND', 'hi')}</b> - चल रहे संगीत को म्यूट करें।\n"
        f"<b>✧ {pick_commands('UNMUTE_COMMAND', 'hi')}</b> - म्यूट किए गए संगीत को अनम्यूट करें।",
        priority=19,
    )
    .name("ku", "بەڕێوەبەرەکان")
    .add(
        "ku",
        f"<b>✧ {pick_commands('MUTE_COMMAND', 'ku')}</b> - گۆرانییەکە بێ دەنگ بکە.\n"
        f"<b>✧ {pick_commands('UNMUTE_COMMAND', 'ku')}</b> - گۆرانییەکە دەنگدار بکە دووبارە.",
        priority=19,
    )
    .name("tr", "Yöneticiler")
    .add(
        "tr",
        f"<b>✧ {pick_commands('MUTE_COMMAND', 'tr')}</b> - Çalan müziği sessize al.\n"
        f"<b>✧ {pick_commands('UNMUTE_COMMAND', 'tr')}</b> - Sessize alınmış müziğin sesini aç.",
        priority=19,
    )
)

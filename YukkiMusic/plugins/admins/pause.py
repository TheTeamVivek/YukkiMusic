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
from YukkiMusic.core.help import ModuleHelp
from YukkiMusic.utils.database import is_music_playing, music_off, music_on
from YukkiMusic.utils.decorators import AdminRightsCheck


@app.on_message(command("PAUSE_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def pause_admin(cli, message: Message, _, chat_id):
    if not len(message.command) == 1:
        return await message.reply_text(_["general_2"])
    if not await is_music_playing(chat_id):
        return await message.reply_text(_["pause_1"])
    await music_off(chat_id)
    await Yukki.pause_stream(chat_id)
    await message.reply_text(_["pause_2"].format(message.from_user.mention))


@app.on_message(command("RESUME_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def resume_com(cli, message: Message, _, chat_id):
    if not len(message.command) == 1:
        return await message.reply_text(_["general_2"])
    if await is_music_playing(chat_id):
        return await message.reply_text(_["resume_1"])
    await music_on(chat_id)
    await Yukki.resume_stream(chat_id)
    await message.reply_text(_["resume_2"].format(message.from_user.mention))


(
    ModuleHelp("Admins")
    .name("en", "Admins")
    .add(
        "en",
        f"<b>✧ {pick_commands('RESUME_COMMAND', 'en')}</b> - Resume the paused music.\n"
        f"<b>✧ {pick_commands('PAUSE_COMMAND', 'en')}</b> - Pause the playing music.",
        priority=18,
    )
    .name("ar", "المسؤولين")
    .add(
        "ar",
        f"<b>✧ {pick_commands('RESUME_COMMAND', 'ar')}</b> - استئناف الموسيقى المتوقفة مؤقتًا.\n"
        f"<b>✧ {pick_commands('PAUSE_COMMAND', 'ar')}</b> - إيقاف الموسيقى مؤقتًا.",
        priority=18,
    )
    .name("as", "প্ৰশাসক")
    .add(
        "as",
        f"<b>✧ {pick_commands('RESUME_COMMAND', 'as')}</b> - ৰখা সংগীতটো পুনৰ বজাওক।\n"
        f"<b>✧ {pick_commands('PAUSE_COMMAND', 'as')}</b> - বজাই থকা সংগীতটো ৰখাওক।",
        priority=18,
    )
    .name("hi", "प्रशासक")
    .add(
        "hi",
        f"<b>✧ {pick_commands('RESUME_COMMAND', 'hi')}</b> - रुके हुए संगीत को फिर से चलाएं।\n"
        f"<b>✧ {pick_commands('PAUSE_COMMAND', 'hi')}</b> - चल रहे संगीत को रोकें।",
        priority=18,
    )
    .name("ku", "بەڕێوەبەرەکان")
    .add(
        "ku",
        f"<b>✧ {pick_commands('RESUME_COMMAND', 'ku')}</b> - گۆرانییە ناچالاکەکە دووبارە دەستپێبکە.\n"
        f"<b>✧ {pick_commands('PAUSE_COMMAND', 'ku')}</b> - گۆرانییە چالاکەکە وەستا.",
        priority=18,
    )
    .name("tr", "Yöneticiler")
    .add(
        "tr",
        f"<b>✧ {pick_commands('RESUME_COMMAND', 'tr')}</b> - Durdurulan müziği devam ettir.\n"
        f"<b>✧ {pick_commands('PAUSE_COMMAND', 'tr')}</b> - Çalan müziği duraklat.",
        priority=18,
    )
)

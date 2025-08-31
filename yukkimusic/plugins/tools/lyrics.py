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
from functools import partial

import lyricsgenius as lg
from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from config import BANNED_USERS
from strings import get_command, pick_commands
from yukkimusic import app
from yukkimusic.utils.decorators.language import language

from . import mhelp

###Commands
LYRICS_COMMAND = get_command("LYRICS_COMMAND")
api_key = "BISewzoVNdCgo26Vf7SGi7D2pCfGdrUdQ5kpIZ1Iz6YhhnaU2ab9T4Rv0r06r3xj"

y = lg.Genius(
    api_key,
    sleep_time=0,
    skip_non_songs=True,
    excluded_terms=["(Remix)", "(Live)"],
    # remove_section_headers=True,
)
MAX_LENGTH = 4096
CACHE = {}


def format_result(item, bot_username):
    r = item["result"]
    title = r.get("title", "Unknown Title")
    artist = r.get("primary_artist", {}).get("name", "Unknown Artist")
    artist_url = r.get("primary_artist", {}).get("url", "")
    song_id = r.get("id", 0)
    return (
        f"<b>{title}</b>\n"
        f'👤 <a href="{artist_url}">{artist}</a>\n'
        f'🔗 <a href="https://t.me/{bot_username}?start=lyr_{song_id}">Get Lyrics</a>\n\n'
    )


def build_pages(query, data, bot_username):
    if query in CACHE:
        return CACHE[query]
    hits = data.get("hits", [])
    if not hits:
        return None, None
    pages, current = [], ""
    for hit in hits:
        text = format_result(hit, bot_username)
        if len(current) + len(text) > MAX_LENGTH:
            pages.append(current)
            current = ""
        current += text
    if current:
        pages.append(current)
    markups = []
    for i in range(len(pages)):
        buttons = []
        if i > 0:
            buttons.append(
                InlineKeyboardButton(
                    "⬅️ Back", callback_data=f"lyrics:page_{i-1}_{query}"
                )
            )
        if i < len(pages) - 1:
            buttons.append(
                InlineKeyboardButton(
                    "Next ➡️", callback_data=f"lyrics:page_{i+1}_{query}"
                )
            )
        markups.append(InlineKeyboardMarkup([buttons]) if buttons else None)
    CACHE[query] = (pages, markups)
    return pages, markups


@app.on_message(filters.command(LYRICS_COMMAND) & ~BANNED_USERS)
@language
async def lrsearch(client, message: Message, _):
    if len(message.command) < 2:
        return await message.reply_text(_["lyrics_1"])
    title = message.text.split(None, 1)[1]
    m = await message.reply_text(_["lyrics_2"])
    try:
        data = await asyncio.to_thread(partial(y.search_songs, title))
    except Exception:
        return await m.edit(_["lyrics_3"].format(title))
    pages, markups = build_pages(title, data, app.username)
    if not pages:
        return await m.edit(_["lyrics_3"].format(title))
    await m.edit(pages[0], reply_markup=markups[0])


@app.on_callback_query(filters.regex(r"^lyrics:page_(\d+)_(.+)$"))
async def lyrics_page_nav(client, cq: CallbackQuery):
    index = int(cq.matches[0].group(1))
    query = cq.matches[0].group(2)
    if query not in CACHE:
        return await cq.answer("Cache expired, please search again.", show_alert=True)
    pages, markups = CACHE[query]
    if index < 0 or index >= len(pages):
        return await cq.answer("Invalid page.", show_alert=True)
    await cq.message.edit_text(
        pages[index],
        reply_markup=markups[index],
        disable_web_page_preview=True,
    )
    await cq.answer()


#   S = y.search_song(title, per_page=5, get_full_info=False)

(
    mhelp.add(
        "en",
        f"<b>★ {pick_commands('LYRICS_COMMAND')}</b> [Music Name] - Search lyrics for the particular music on the web.",
    )
    .add(
        "ar",
        f"<b>★ {pick_commands('LYRICS_COMMAND')}</b> [اسم الموسيقى] - ابحث عن كلمات الأغنية المحددة على الويب.",
    )
    .add(
        "as",
        f"<b>★ {pick_commands('LYRICS_COMMAND')}</b> [গীতৰ নাম] - ৱেবত বিশেষ গীতৰ লিৰিকচ সন্ধান কৰক।",
    )
    .add(
        "hi",
        f"<b>★ {pick_commands('LYRICS_COMMAND')}</b> [संगीत का नाम] - वेब पर विशेष गीत के बोल खोजें।",
    )
    .add(
        "ku",
        f"<b>★ {pick_commands('LYRICS_COMMAND')}</b> [ناوی میوزیک] - لەسەر وێب بگەڕێ بۆ هۆنراوی گۆرانییە دیاریکراوەکان.",
    )
    .add(
        "tr",
        f"<b>★ {pick_commands('LYRICS_COMMAND')}</b> [Müzik Adı] - Web'de belirli bir müziğin sözlerini arayın.",
    )
)

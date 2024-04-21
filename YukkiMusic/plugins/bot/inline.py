#
# Copyright (C) 2021-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

import traceback
from pyrogram import filters
from pykeyboard import InlineKeyboard

from YukkiMusic import app
from YukkiMusic.utils.inline import *

__MODULE__ = "Inline"
__HELP__ = f"""

Send /inline for help related to inline.

**Example Usage:**
`@{app.username} google github`
"""


@app.on_message(filters.command("inline"))
async def inline(_, message):
    buttons = InlineKeyboard(row_width=4)
    buttons.add(
        *[
            (InlineKeyboardButton(text=i, switch_inline_query_current_chat=i))
            for i in keywords_list
        ]
    )
    text = "**ʜᴇʀᴇ ᴀʀᴇ sᴏᴍᴇ ᴀᴠᴀɪʟᴀʙʟᴇ ɪɴʟɪɴᴇ ᴄᴏᴍᴍᴀɴᴅs ғᴏʀ ɪɴʟɪɴᴇ**"
    await message.reply_text(text=text, reply_markup=buttons)


@app.on_inline_query()
async def inline_query_handler(client, query):
    try:
        text = query.query.strip().lower()
        answers = []
        if text.strip() == "":
            answerss = await inline_help_func(__HELP__)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=10
            )
        elif text.split()[0] == "tr":
            if len(text.split()) < 3:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Translator | tr [LANG] [TEXT]",
                    switch_pm_parameter="inline",
                )
            lang = text.split()[1]
            tex = text.split(None, 2)[2].strip()
            answerss = await translate_func(answers, lang, tex)
            await client.answer_inline_query(
                query.id,
                results=answerss,
            )
        elif text.split()[0] == "ud":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Urban Dictionary | ud [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await urban_func(answers, tex)
            await client.answer_inline_query(
                query.id,
                results=answerss,
            )
        elif text.split()[0] == "google":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Google Search | google [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await google_search_func(answers, tex)
            await client.answer_inline_query(
                query.id,
                results=answerss,
            )

        elif text.split()[0] == "wall":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    is_gallery=True,
                    switch_pm_text="Wallpapers Search | wall [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await wall_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "torrent":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Torrent Search | torrent [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await torrent_func(answers, tex)
            await client.answer_inline_query(
                query.id,
                results=answerss,
            )

        elif text.split()[0] == "yt":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="YouTube Search | yt [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await youtube_func(answers, tex)
            await client.answer_inline_query(query.id, results=answerss)

        elif text.split()[0] == "lyrics":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Lyrics Search | lyrics [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await lyrics_func(answers, tex)
            await client.answer_inline_query(query.id, results=answerss)

        elif text.split()[0] == "search":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Global Message Search. | search [QUERY]",
                    switch_pm_parameter="inline",
                )
            user_id = query.from_user.id
            tex = text.split(None, 1)[1].strip()
            answerss = await tg_search_func(answers, tex, user_id)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "music":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Music Search | music [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await music_inline_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "wiki":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="Wikipedia | wiki [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await wiki_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "ytmusic":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="YT Music | ytmusic [url]",
                    switch_pm_parameter="inline",
                )
            tex = query.query.split(None, 1)[1].strip()
            answerss = await yt_music_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "info":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    switch_pm_text="User Info | info [USERNAME|ID]",
                    switch_pm_parameter="inline",
                )
            tex = text.split()[1].strip()
            answerss = await info_inline_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "tmdb":
            if len(text.split()) < 2:
                answerss = await tmdb_func(answers, "")
                return await client.answer_inline_query(
                    query.id,
                    results=answerss,
                    switch_pm_text="TMDB Search | tmdb [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split()[1].strip()
            answerss = await tmdb_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=2
            )

        elif text.split()[0] == "image":
            if len(text.split()) < 2:
                return await client.answer_inline_query(
                    query.id,
                    results=answers,
                    is_gallery=True,
                    switch_pm_text="Image Search | image [QUERY]",
                    switch_pm_parameter="inline",
                )
            tex = text.split(None, 1)[1].strip()
            answerss = await image_func(answers, tex)
            await client.answer_inline_query(
                query.id, results=answerss, cache_time=3600
            )

        elif text.split()[0] == "exec":
            await execute_code(query)

        
    except Exception as e:
        e = traceback.format_exc()
        print(e, " InLine") 
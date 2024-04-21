import asyncio
import os
from contextlib import suppress
from html import escape
from re import sub as re_sub
from time import ctime, time
from fuzzysearch import find_near_matches

from pykeyboard import InlineKeyboard
from pyrogram import enums, filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineQueryResultArticle,
    InlineQueryResultCachedDocument,
    InlineQueryResultPhoto,
    InputTextMessageContent,
)
from search_engine_parser import GoogleSearch

from config import LOG_GROUP_ID
from YukkiMusic.misc import SUDOERS
from YukkiMusic import arq, app, aiohttpsession
from YukkiMusic.utils.keyboard import ikb
from YukkiMusic.plugins.tools.info import get_chat_info, get_user_info
from YukkiMusic.plugins.tools.music import download_youtube_audio
from YukkiMusic.utils.pastebin import Yukkibin

keywords_list = [
    "image",
    "wall",
    "tmdb",
    "lyrics",
    "exec",
    "search",
    "tr",
    "ud",
    "yt",
    "info",
    "google",
    "torrent",
    "wiki",
    "music",
    "ytmusic",
]


async def inline_help_func(__HELP__):
    buttons = InlineKeyboard(row_width=4)
    buttons.add(
        *[
            (InlineKeyboardButton(text=i, switch_inline_query_current_chat=i))
            for i in keywords_list
        ]
    )
    answerss = [
        InlineQueryResultArticle(
            title="ɪɴʟɪɴᴇ ᴄᴏᴍᴍᴀɴᴅs",
            description="ɪɴʟɪɴᴇ ᴜsᴀsɢᴇ ʀᴇʟᴀᴛᴇᴅ ʜᴇʟᴘ",
            input_message_content=InputTextMessageContent(
                "ᴄʟɪᴄᴋ ᴀ ʙᴜᴛᴛᴏɴ ᴀɴᴅ sᴇᴇ ᴍᴀɢɪᴄ."
            ),
            thumb_url="https://hamker.me/cy00x5x.png",
            reply_markup=buttons,
        ),
    ]
    answerss = await alive_function(answerss)
    return answerss



async def translate_func(answers, lang, tex):
    result = await arq.translate(tex, lang)
    if not result.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=result.result,
                input_message_content=InputTextMessageContent(result.result),
            )
        )
        return answers
    result = result.result
    msg = f"""
__**Translated from {result.src} to {result.dest}**__

**INPUT:**
{tex}

**OUTPUT:**
{result.translatedText}"""
    answers.extend(
        [
            InlineQueryResultArticle(
                title=f"Translated from {result.src} to {result.dest}.",
                description=result.translatedText,
                input_message_content=InputTextMessageContent(msg),
            ),
            InlineQueryResultArticle(
                title=result.translatedText,
                input_message_content=InputTextMessageContent(
                    result.translatedText
                ),
            ),
        ]
    )
    return answers


async def urban_func(answers, text):
    results = await arq.urbandict(text)
    if not results.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=results.result,
                input_message_content=InputTextMessageContent(results.result),
            )
        )
        return answers
    results = results.result[0:48]
    for i in results:
        clean = lambda x: re_sub(r"[\[\]]", "", x)
        msg = f"""
**Query:** {text}

**Definition:** __{clean(i.definition)}__

**Example:** __{clean(i.example)}__"""

        answers.append(
            InlineQueryResultArticle(
                title=i.word,
                description=clean(i.definition),
                input_message_content=InputTextMessageContent(msg),
            )
        )
    return answers


async def google_search_func(answers, text):
    gresults = await GoogleSearch().async_search(text)
    limit = 0
    for i in gresults:
        if limit > 48:
            break
        limit += 1

        with suppress(KeyError):
            msg = f"""
[{i['titles']}]({i['links']})
{i['descriptions']}"""

            answers.append(
                InlineQueryResultArticle(
                    title=i["titles"],
                    description=i["descriptions"],
                    input_message_content=InputTextMessageContent(
                        msg, disable_web_page_preview=True
                    ),
                )
            )
    return answers


async def wall_func(answers, text):
    results = await arq.wall(text)
    if not results.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=results.result,
                input_message_content=InputTextMessageContent(results.result),
            )
        )
        return answers
    results = results.result[0:48]
    for i in results:
        answers.append(
            InlineQueryResultPhoto(
                photo_url=i.url_image,
                thumb_url=i.url_thumb,
                caption=f"[Source]({i.url_image})",
            )
        )
    return answers


async def torrent_func(answers, text):
    results = await arq.torrent(text)
    if not results.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=results.result,
                input_message_content=InputTextMessageContent(results.result),
            )
        )
        return answers
    results = results.result[0:48]
    for i in results:
        title = i.name
        size = i.size
        seeds = i.seeds
        leechs = i.leechs
        upload_date = i.uploaded
        magnet = i.magnet
        caption = f"""
**Title:** __{title}__
**Size:** __{size}__
**Seeds:** __{seeds}__
**Leechs:** __{leechs}__
**Uploaded:** __{upload_date}__
**Magnet:** `{magnet}`"""

        description = f"{size} | {upload_date} | Seeds: {seeds}"
        answers.append(
            InlineQueryResultArticle(
                title=title,
                description=description,
                input_message_content=InputTextMessageContent(
                    caption, disable_web_page_preview=True
                ),
            )
        )
    return answers


async def youtube_func(answers, text):
    results = await arq.youtube(text)
    if not results.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=results.result,
                input_message_content=InputTextMessageContent(results.result),
            )
        )
        return answers
    results = results.result[0:48]
    for i in results:
        buttons = InlineKeyboard(row_width=1)
        video_url = f"https://youtube.com{i.url_suffix}"
        buttons.add(InlineKeyboardButton("Watch", url=video_url))
        caption = f"""
**Title:** {i.title}
**Views:** {i.views}
**Channel:** {i.channel}
**Duration:** {i.duration}
**Uploaded:** {i.publish_time}
**Description:** {i.long_desc}"""
        description = (
            f"{i.views} | {i.channel} | {i.duration} | {i.publish_time}"
        )
        answers.append(
            InlineQueryResultArticle(
                title=i.title,
                thumb_url=i.thumbnails[0],
                description=description,
                input_message_content=InputTextMessageContent(
                    caption, disable_web_page_preview=True
                ),
                reply_markup=buttons,
            )
        )
    return answers


async def lyrics_func(answers, text):
    resp = await arq.lyrics(text)
    if not resp.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=resp.result,
                input_message_content=InputTextMessageContent(resp.result),
            )
        )
        return answers
    songs = resp.result
    for song in songs:
        song_name = song["song"]
        artist = song["artist"]
        lyrics = song["lyrics"]
        msg = f"**{song_name}** | **{artist}**\n\n__{lyrics}__"

        if len(msg) > 4095:
            msg = await Yukkibin(msg)
            msg = f"**LYRICS_TOO_LONG:** [URL]({msg})"

        answers.append(
            InlineQueryResultArticle(
                title=song_name,
                description=artist,
                input_message_content=InputTextMessageContent(msg),
            )
        )
    return answers


async def tg_search_func(answers, text, user_id):
    if user_id not in SUDOERS:
        msg = "**ERROR**\n__THIS FEATURE IS ONLY FOR SUDO USERS__"
        answers.append(
            InlineQueryResultArticle(
                title="ERROR",
                description="THIS FEATURE IS ONLY FOR SUDO USERS",
                input_message_content=InputTextMessageContent(msg),
            )
        )
        return answers
    if str(text)[-1] != ":":
        msg = "**ERROR**\n__Put A ':' After The Text To Search__"
        answers.append(
            InlineQueryResultArticle(
                title="ERROR",
                description="Put A ':' After The Text To Search",
                input_message_content=InputTextMessageContent(msg),
            )
        )

        return answers
    text = text[0:-1]
    app2 = await get_client(1)
    async for message in app2.search_global(text, limit=49):
        buttons = InlineKeyboard(row_width=2)
        buttons.add(
            InlineKeyboardButton(
                text="Origin",
                url=message.link if message.link else "https://t.me/telegram",
            ),
            InlineKeyboardButton(
                text="Search again",
                switch_inline_query_current_chat="search",
            ),
        )
        name = (
            message.from_user.first_name
            if message.from_user.first_name
            else "NO NAME"
        )
        caption = f"""
**Query:** {text}
**Name:** {str(name)} [`{message.from_user.id}`]
**Chat:** {str(message.chat.title)} [`{message.chat.id}`]
**Date:** {ctime(message.date)}
**Text:** >>

{message.text.markdown if message.text else message.caption if message.caption else '[NO_TEXT]'}
"""
        result = InlineQueryResultArticle(
            title=name,
            description=message.text if message.text else "[NO_TEXT]",
            reply_markup=buttons,
            input_message_content=InputTextMessageContent(
                caption, disable_web_page_preview=True
            ),
        )
        answers.append(result)
    return answers


async def music_inline_func(answers, query):
    chat_id = -1001445180719
    app2 = await get_client(1)
    group_invite = "https://t.me/joinchat/vSDE2DuGK4Y4Nzll"
    try:
        messages = [
            m
            async for m in app2.search_messages(
                chat_id, query, filter=enums.MessagesFilter.AUDIO, limit=100
            )
        ]
    except Exception as e:
        print(e)
        msg = f"You Need To Join Here With Your Bot And Userbot To Get Cached Music.\n{group_invite}"
        answers.append(
            InlineQueryResultArticle(
                title="ERROR",
                description="Click Here To Know More.",
                input_message_content=InputTextMessageContent(
                    msg, disable_web_page_preview=True
                ),
            )
        )
        return answers
    messages_ids_and_duration = []
    for f_ in messages:
        messages_ids_and_duration.append(
            {
                "id": f_.id,
                "duration": f_.audio.duration if f_.audio.duration else 0,
            }
        )
    messages = list(
        {v["duration"]: v for v in messages_ids_and_duration}.values()
    )
    messages_ids = [ff_["id"] for ff_ in messages]
    messages = await app.get_messages(chat_id, messages_ids[0:48])
    return [
        InlineQueryResultCachedDocument(
            document_file_id=message_.audio.file_id,
            title=message_.audio.title,
        )
        for message_ in messages
    ]


async def wiki_func(answers, text):
    data = await arq.wiki(text)
    if not data.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=data.result,
                input_message_content=InputTextMessageContent(data.result),
            )
        )
        return answers
    data = data.result
    msg = f"""
**QUERY:**
{data.title}

**ANSWER:**
__{data.answer}__"""
    answers.append(
        InlineQueryResultArticle(
            title=data.title,
            description=data.answer,
            input_message_content=InputTextMessageContent(msg),
        )
    )
    return answers




async def yt_music_func(answers, url):
    arq_resp = await arq.youtube(url)
    loop = asyncio.get_running_loop()
    music = await loop.run_in_executor(None, download_youtube_audio, arq_resp)
    if not music:
        msg = "**ERROR**\n__MUSIC TOO LONG__"
        answers.append(
            InlineQueryResultArticle(
                title="ERROR",
                description="MUSIC TOO LONG",
                input_message_content=InputTextMessageContent(msg),
            )
        )
        return answers
    (
        title,
        performer,
        duration,
        audio,
        thumbnail,
    ) = music
    m = await app.send_audio(
        LOG_GROUP_ID,
        audio,
        title=title,
        duration=duration,
        performer=performer,
        thumb=thumbnail,
    )
    os.remove(audio)
    os.remove(thumbnail)
    answers.append(
        InlineQueryResultCachedDocument(
            title=title, document_file_id=m.audio.file_id
        )
    )
    return answers


async def info_inline_func(answers, peer):
    not_found = InlineQueryResultArticle(
        title="PEER NOT FOUND",
        input_message_content=InputTextMessageContent("PEER NOT FOUND"),
    )
    try:
        user = await app.get_users(peer)
        caption, _ = await get_user_info(user, True)
    except IndexError:
        try:
            chat = await app.get_chat(peer)
            caption, _ = await get_chat_info(chat, True)
        except Exception:
            return [not_found]
    except Exception:
        return [not_found]

    answers.append(
        InlineQueryResultArticle(
            title="Found Peer.",
            input_message_content=InputTextMessageContent(
                caption, disable_web_page_preview=True
            ),
        )
    )
    return answers


async def tmdb_func(answers, query):
    response = await arq.tmdb(query)
    if not response.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=response.result,
                input_message_content=InputTextMessageContent(response.result),
            )
        )
        return answers
    results = response.result[:49]
    for result in results:
        if not result.poster and not result.backdrop:
            continue
        if not result.genre:
            genre = None
        else:
            genre = " | ".join(result.genre)
        description = result.overview[0:900] if result.overview else "None"
        caption = f"""
**{result.title}**
**Type:** {result.type}
**Rating:** {result.rating}
**Genre:** {genre}
**Release Date:** {result.releaseDate}
**Description:** __{description}__
"""
        buttons = InlineKeyboard(row_width=1)
        buttons.add(
            InlineKeyboardButton(
                "Search Again",
                switch_inline_query_current_chat="tmdb",
            )
        )
        answers.append(
            InlineQueryResultPhoto(
                photo_url=result.backdrop
                if result.backdrop
                else result.poster,
                caption=caption,
                title=result.title,
                description=f"{genre} • {result.releaseDate} • {result.rating} • {description}",
                reply_markup=buttons,
            )
        )
    return answers


async def image_func(answers, query):
    results = await arq.image(query)
    if not results.ok:
        answers.append(
            InlineQueryResultArticle(
                title="Error",
                description=results.result,
                input_message_content=InputTextMessageContent(results.result),
            )
        )
        return answers
    results = results.result[:49]
    buttons = InlineKeyboard(row_width=2)
    buttons.add(
        InlineKeyboardButton(
            text="Search again",
            switch_inline_query_current_chat="image",
        ),
    )
    for i in results:
        answers.append(
            InlineQueryResultPhoto(
                title=i.title,
                photo_url=i.url,
                thumb_url=i.url,
                reply_markup=buttons,
            )
        )
    return answers


async def execute_code(query):
    text = query.query.strip()
    offset = int((query.offset or 0))
    answers = []
    languages = (await arq.execute()).result
    if len(text.split()) == 1:
        answers = [
            InlineQueryResultArticle(
                title=lang,
                input_message_content=InputTextMessageContent(lang),
            )
            for lang in languages
        ][offset : offset + 25]
        await query.answer(
            next_offset=str(offset + 25),
            results=answers,
            cache_time=1,
        )
    elif len(text.split()) == 2:
        text = text.split()[1].strip()
        languages = list(
            filter(
                lambda x: find_near_matches(text, x, max_l_dist=1),
                languages,
            )
        )
        answers.extend(
            [
                InlineQueryResultArticle(
                    title=lang,
                    input_message_content=InputTextMessageContent(lang),
                )
                for lang in languages
            ][:49]
        )
    else:
        lang = text.split()[1]
        code = text.split(None, 2)[2]
        response = await arq.execute(lang, code)
        if not response.ok:
            answers.append(
                InlineQueryResultArticle(
                    title="Error",
                    input_message_content=InputTextMessageContent(
                        response.result
                    ),
                )
            )
        else:
            res = response.result
            stdout, stderr = escape(res.stdout), escape(res.stderr)
            output = stdout or stderr
            out = "STDOUT" if stdout else ("STDERR" if stderr else "No output")

            msg = f"""
**{lang.capitalize()}:**
```{code}```

**{out}:**
```{output}```
            """
            answers.append(
                InlineQueryResultArticle(
                    title="Executed",
                    description=output[:20],
                    input_message_content=InputTextMessageContent(msg),
                )
            )
    await query.answer(results=answers, cache_time=1)

 
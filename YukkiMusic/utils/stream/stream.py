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
import os
from random import randint

from pyrogram.types import InlineKeyboardMarkup, Message, User

import config
from YukkiMusic import Platform, app
from YukkiMusic.core.call import Yukki
from YukkiMusic.core.enum import SongType
from YukkiMusic.core.youtube import Track
from YukkiMusic.misc import db
from YukkiMusic.utils.database import (
    add_active_video_chat,
    is_active_chat,
    is_video_allowed,
)
from YukkiMusic.utils.exceptions import AssistantErr
from YukkiMusic.utils.inline.play import stream_markup, telegram_markup
from YukkiMusic.utils.inline.playlist import close_markup
from YukkiMusic.utils.pastebin import paste
from YukkiMusic.utils.stream.queue import put_queue, put_queue_index
from YukkiMusic.utils.thumbnails import gen_qthumb, gen_thumb


async def stream(
    _,  # TODO remove this from argument and get the lang of the chatid under this function
    mystic: Message,
    user_id: User,
    #  result,
    tracks: Track | list[Track],  # Replacement of result
    chat_id: int,
    user_name,  # TODO Remove user_name if user_id belongs to same user
    original_chat_id,
    #  video: bool | str = None,
    type: SongType = SongType.AUDIO,  # TODO rename video with  this type
    #  streamtype: bool | str = None, #TODO remove this Beacuse Track didn't want all tracks instance are same
    # spotify: bool | str = None, # Since the tracks are is already and instance of Track so we don't need to this Beacuse the Tracks is already Contains all result of Song
    forceplay: bool | None = None,
):
    if not result:
        return
    if video:
        if not await is_video_allowed(chat_id):
            raise AssistantErr(_["play_7"])
    if forceplay:
        await Yukki.force_stop_stream(chat_id)
    # if streamtype == "playlist":
    if isinstance(
        tracks, list
    ):  # TODO If YouTube Playlist returns list of vidid but other returns list of song name and can be list of Track

        msg = f"{_['playlist_16']}\n\n"
        count = 0
        r = await asyncio.gather(
            *[track.download() for track in tracks[: config.PLAYLIST_FETCH_LIMIT]],
            return_exceptions=True,
        )  # TODO: We Need to make the track.download compatible with m3u8 support
        for res in r:
            if isinstance(res, Exception):
                pass  # TODO use app.report_error for reporting for logger group or all owners and logs all log needed to create that function
        for search in result:
            # try:
            #     (
            #         title,
            #         duration_min,
            #         duration_sec,
            #         thumbnail, #TDOO REMOVE THIS ALL AND USE tracks
            #         vidid,
            #     ) = await Platform.youtube.track(search, False if spotify else True)
            # except Exception:
            #     continue
            if str(duration_min) == "None":
                continue
            if duration_sec > config.DURATION_LIMIT:
                continue
            if await is_active_chat(chat_id):
                await put_queue(
                    chat_id,
                    original_chat_id,
                    f"vid_{vidid}",
                    title,  # TODO put all Track insted track name
                    duration_min,
                    user_name,
                    vidid,
                    user_id,
                    "video" if video else "audio",
                )
                position = len(db.get(chat_id)) - 1
                count += 1
                msg += f"{count}- {title[:70]}\n"
                msg += f"{_['playlist_17']} {position}\n\n"
            else:
                if not forceplay:
                    db[chat_id] = []
                status = True if video else None
                try:
                    file_path, direct = await Platform.youtube.download(
                        vidid, mystic, video=status, videoid=True
                    )
                except Exception:
                    raise AssistantErr(_["play_16"])
                await Yukki.join_call(  # MAYBE: The join_call didn't require the original_chat_id remive it
                    chat_id, original_chat_id, file_path, video=status, image=thumbnail
                )
                await put_queue(
                    chat_id,
                    original_chat_id,
                    file_path if direct else f"vid_{vidid}",
                    title,
                    duration_min,
                    user_name,
                    vidid,
                    user_id,
                    "video" if video else "audio",
                    forceplay=forceplay,
                )
                img = await gen_thumb(
                    vidid
                )  # TODO Remove Thumbnail Support Or Add multiple Theme support can be off or changed by any command
                button = stream_markup(_, vidid, chat_id)
                run = await app.send_photo(  # TDDO Put this at the end and outside function
                    original_chat_id,
                    photo=img,
                    caption=_["stream_1"].format(
                        title[:27],
                        f"https://t.me/{app.username}?start=info_{vidid}",
                        duration_min,
                        user_name,
                    ),
                    reply_markup=InlineKeyboardMarkup(button),
                )
                db[chat_id][0]["mystic"] = run
                db[chat_id][0]["markup"] = "stream"
        if count == 0:
            return
        else:
            link = await paste(msg)
            lines = msg.count("\n")
            if lines >= 17:
                car = os.linesep.join(msg.split(os.linesep)[:17])
            else:
                car = msg
            carbon = await Platform.carbon.generate(
                car, randint(100, 10000000)
            )  # GUESS WHAT: Remove Carbon from platforms
            upl = close_markup(_)
            return await app.send_photo(
                original_chat_id,
                photo=carbon,
                caption=_["playlist_18"].format(link, position),
                reply_markup=upl,
            )

    elif (
        streamtype == "youtube"
    ):  # No Need for streamttpe Beacuse all tracks of every platform return same type and has all same attr
        link = result["link"]
        vidid = result["vidid"]
        title = (result["title"]).title()
        duration_min = result["duration_min"]
        thumbnail = result["thumb"]
        status = True if video else None
        try:
            file_path, direct = await Platform.youtube.download(
                vidid, mystic, videoid=True, video=status
            )
        except Exception:
            raise AssistantErr(_["play_16"])
        if await is_active_chat(chat_id):
            await put_queue(
                chat_id,
                original_chat_id,
                file_path if direct else f"vid_{vidid}",
                title,
                duration_min,
                user_name,
                vidid,
                user_id,
                "video" if video else "audio",
            )
            position = len(db.get(chat_id)) - 1
            qimg = await gen_qthumb(vidid)
            run = await app.send_photo(
                original_chat_id,
                photo=qimg,
                caption=_["queue_4"].format(
                    position, title[:27], duration_min, user_name
                ),
                reply_markup=close_markup(_),
            )
        else:
            if not forceplay:
                db[chat_id] = []
            await Yukki.join_call(
                chat_id, original_chat_id, file_path, video=status, image=thumbnail
            )
            await put_queue(  # We Can simplify the put_queue or the db[chat_id] with a Queue Class
                chat_id,
                original_chat_id,
                file_path if direct else f"vid_{vidid}",
                title,
                duration_min,
                user_name,
                vidid,
                user_id,
                "video" if video else "audio",
                forceplay=forceplay,
            )
            img = await gen_thumb(vidid)
            button = stream_markup(_, vidid, chat_id)
            run = await app.send_photo(
                original_chat_id,
                photo=img,
                caption=_["stream_1"].format(
                    title[:27],
                    f"https://t.me/{app.username}?start=info_{vidid}",
                    duration_min,
                    user_name,
                ),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "stream"

    elif "saavn" in streamtype:
        if streamtype == "saavn_track":
            if result["duration_sec"] == 0:
                return
            file_path = result["filepath"]
            title = result["title"]
            duration_min = result["duration_min"]
            link = result["url"]
            thumb = result["thumb"]
            if await is_active_chat(chat_id):
                await put_queue(
                    chat_id,
                    original_chat_id,
                    file_path,
                    title,
                    duration_min,
                    user_name,
                    streamtype,
                    user_id,
                    "audio",
                    url=link,
                )
                position = len(db.get(chat_id)) - 1
                await app.send_photo(
                    original_chat_id,
                    photo=thumb or "https://envs.sh/Ii_.jpg",
                    caption=_["queue_4"].format(
                        position, title[:30], duration_min, user_name
                    ),
                    reply_markup=close_markup(_),
                )
            else:
                if not forceplay:
                    db[chat_id] = []
                await Yukki.join_call(chat_id, original_chat_id, file_path, video=None)
                await put_queue(
                    chat_id,
                    original_chat_id,
                    file_path,
                    title,
                    duration_min,
                    user_name,
                    streamtype,
                    user_id,
                    "audio",
                    forceplay=forceplay,
                    url=link,
                )
                button = telegram_markup(_, chat_id)
                run = await app.send_photo(
                    original_chat_id,
                    photo=thumb,
                    caption=_["stream_1"].format(
                        title, config.SUPPORT_GROUP, duration_min, user_name
                    ),
                    reply_markup=InlineKeyboardMarkup(button),
                )
                db[chat_id][0]["mystic"] = run
                db[chat_id][0]["markup"] = "tg"

        elif streamtype == "saavn_playlist":
            msg = f"{_['playlist_16']}\n\n"
            count = 0
            for search in result:
                if search["duration_sec"] == 0:
                    continue
                title = search["title"]
                duration_min = search["duration_min"]
                search["duration_sec"]
                link = search["url"]
                thumb = search["thumb"]
                file_path, n = await Platform.saavn.download(link)
                if await is_active_chat(chat_id):
                    await put_queue(
                        chat_id,
                        original_chat_id,
                        file_path,
                        title,
                        duration_min,
                        user_name,
                        streamtype,
                        user_id,
                        "audio",
                        url=link,
                    )
                    position = len(db.get(chat_id)) - 1
                    count += 1
                    msg += f"{count}- {title[:70]}\n"
                    msg += f"{_['playlist_17']} {position}\n\n"

                else:

                    if not forceplay:
                        db[chat_id] = []
                    await Yukki.join_call(
                        chat_id, original_chat_id, file_path, video=None
                    )
                    await put_queue(
                        chat_id,
                        original_chat_id,
                        file_path,
                        title,
                        duration_min,
                        user_name,
                        streamtype,
                        user_id,
                        "audio",
                        forceplay=forceplay,
                        url=link,
                    )
                    button = telegram_markup(_, chat_id)
                    run = await app.send_photo(
                        original_chat_id,
                        photo=thumb,
                        caption=_["stream_1"].format(
                            title, link, duration_min, user_name
                        ),
                        reply_markup=InlineKeyboardMarkup(button),
                    )
                    db[chat_id][0]["mystic"] = run
                    db[chat_id][0]["markup"] = "tg"
            if count == 0:
                return
            else:
                link = await paste(msg)
                lines = msg.count("\n")
                if lines >= 17:
                    car = os.linesep.join(msg.split(os.linesep)[:17])
                else:
                    car = msg
                carbon = await Platform.carbon.generate(car, randint(100, 10000000))
                upl = close_markup(_)
                return await app.send_photo(
                    original_chat_id,
                    photo=carbon,
                    caption=_["playlist_18"].format(link, position),
                    reply_markup=upl,
                )

    elif streamtype == "soundcloud":
        file_path = result["filepath"]
        title = result["title"]
        duration_min = result["duration_min"]
        if await is_active_chat(chat_id):
            await put_queue(
                chat_id,
                original_chat_id,
                file_path,
                title,
                duration_min,
                user_name,
                streamtype,
                user_id,
                "audio",
            )
            position = len(db.get(chat_id)) - 1
            await app.send_message(
                original_chat_id,
                _["queue_4"].format(position, title[:30], duration_min, user_name),
            )
        else:
            if not forceplay:
                db[chat_id] = []
            await Yukki.join_call(chat_id, original_chat_id, file_path, video=None)
            await put_queue(
                chat_id,
                original_chat_id,
                file_path,
                title,
                duration_min,
                user_name,
                streamtype,
                user_id,
                "audio",
                forceplay=forceplay,
            )
            button = telegram_markup(_, chat_id)
            run = await app.send_photo(
                original_chat_id,
                photo=config.SOUNCLOUD_IMG_URL,
                caption=_["stream_1"].format(
                    title, config.SUPPORT_GROUP, duration_min, user_name
                ),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
    elif streamtype == "telegram":
        file_path = result["path"]
        link = result["link"]
        title = (result["title"]).title()
        duration_min = result["dur"]
        status = True if video else None
        if await is_active_chat(chat_id):
            await put_queue(
                chat_id,
                original_chat_id,
                file_path,
                title,
                duration_min,
                user_name,
                streamtype,
                user_id,
                "video" if video else "audio",
            )
            position = len(db.get(chat_id)) - 1
            await app.send_message(
                original_chat_id,
                _["queue_4"].format(position, title[:30], duration_min, user_name),
            )
        else:
            if not forceplay:
                db[chat_id] = []
            await Yukki.join_call(chat_id, original_chat_id, file_path, video=status)
            await put_queue(
                chat_id,
                original_chat_id,
                file_path,
                title,
                duration_min,
                user_name,
                streamtype,
                user_id,
                "video" if video else "audio",
                forceplay=forceplay,
            )
            if video:
                await add_active_video_chat(chat_id)
            button = telegram_markup(_, chat_id)
            run = await app.send_photo(
                original_chat_id,
                photo=config.TELEGRAM_VIDEO_URL if video else config.TELEGRAM_AUDIO_URL,
                caption=_["stream_1"].format(title, link, duration_min, user_name),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
    elif streamtype == "live":
        link = result["link"]
        vidid = result["vidid"]
        title = (result["title"]).title()
        thumbnail = result["thumb"]
        duration_min = "00:00"
        status = True if video else None
        if await is_active_chat(chat_id):
            await put_queue(
                chat_id,
                original_chat_id,
                f"live_{vidid}",
                title,
                duration_min,
                user_name,
                vidid,
                user_id,
                "video" if video else "audio",
            )
            position = len(db.get(chat_id)) - 1
            await app.send_message(
                original_chat_id,
                _["queue_4"].format(position, title[:30], duration_min, user_name),
            )
        else:
            if not forceplay:
                db[chat_id] = []
            n, file_path = await Platform.youtube.video(link)
            if n == 0:
                raise AssistantErr(_["str_3"])
            await Yukki.join_call(
                chat_id,
                original_chat_id,
                file_path,
                video=status,
                image=thumbnail if thumbnail else None,
            )
            await put_queue(
                chat_id,
                original_chat_id,
                f"live_{vidid}",
                title,
                duration_min,
                user_name,
                vidid,
                user_id,
                "video" if video else "audio",
                forceplay=forceplay,
            )
            img = await gen_thumb(vidid)
            button = telegram_markup(_, chat_id)
            run = await app.send_photo(
                original_chat_id,
                photo=img,
                caption=_["stream_1"].format(
                    title[:27],
                    f"https://t.me/{app.username}?start=info_{vidid}",
                    duration_min,
                    user_name,
                ),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
    elif streamtype == "index":
        link = result
        title = "Index or M3u8 Link"
        duration_min = "URL stream"
        if await is_active_chat(chat_id):
            await put_queue_index(
                chat_id,
                original_chat_id,
                "index_url",
                title,
                duration_min,
                user_name,
                link,
                "video" if video else "audio",
            )
            position = len(db.get(chat_id)) - 1
            await mystic.edit_text(
                _["queue_4"].format(position, title[:30], duration_min, user_name)
            )
        else:
            if not forceplay:
                db[chat_id] = []
            await Yukki.join_call(
                chat_id,
                original_chat_id,
                link,
                video=True if video else None,
            )
            await put_queue_index(
                chat_id,
                original_chat_id,
                "index_url",
                title,
                duration_min,
                user_name,
                link,
                "video" if video else "audio",
                forceplay=forceplay,
            )
            button = telegram_markup(_, chat_id)
            run = await app.send_photo(
                original_chat_id,
                photo=config.STREAM_IMG_URL,
                caption=_["stream_2"].format(user_name),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
            await mystic.delete()


# AND LAST: The if await is_active_chat was repeating many times in many conditions Short the usage and the branch or the stream

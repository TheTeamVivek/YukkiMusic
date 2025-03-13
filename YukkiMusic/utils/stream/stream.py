#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os
from random import randint

import config
from strings import get_string
from YukkiMusic import tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.core.youtube import Track
from YukkiMusic.misc import db
from YukkiMusic.platforms import carbon
from YukkiMusic.utils.database import get_lang, is_active_chat, is_video_allowed
from YukkiMusic.utils.exceptions import AssistantErr
from YukkiMusic.utils.formatters import seconds_to_min
from YukkiMusic.utils.inline.play import play_markup
from YukkiMusic.utils.inline.playlist import close_markup
from YukkiMusic.utils.pastebin import paste
from YukkiMusic.utils.stream.queue import put_queue
from YukkiMusic.utils.thumbnails import gen_qthumb, gen_thumb


async def stream(
    chat_id: int,
    original_chat_id,
    track: Track | list[Track],
    user_id: int,
    forceplay: bool | None = None,
):
    language = await get_lang(original_chat_id)
    _ = get_string(language)
    user_mention = await tbot.create_mention(user_id)
    is_queue_ = False

    if not result:
        return
    if video:
        if not await is_video_allowed(chat_id):
            raise AssistantErr(_["play_7"])
    if forceplay:
        await Yukki.force_stop_stream(chat_id)
    if isinstance(track, list):

        msg = f"{_['playlist_16']}\n\n"
        count = 0
        track = track[: config.PLAYLIST_FETCH_LIMIT]

        for song in track:
            if not song.duration or song.duration > config.DURATION_LIMIT:
                continue
            if await is_active_chat(chat_id):
                await put_queue(
                    chat_id=chat_id,
                    original_chat_id=original_chat_id,
                    user_id=user_id,
                    track=song,
                )
                position = len(db.get(chat_id)) - 1
                count += 1
                msg += f"{count}- {title[:70]}\n"
                msg += f"{_['playlist_17']} {position}\n\n"
            else:
                if not forceplay:
                    db[chat_id] = []

                try:
                    file_path = await song.download()
                except Exception as e:
                    await tbot.handle_error(e)
                    raise AssistantErr(_["play_16"])
                await Yukki.join_call(
                    chat_id, file_path, video=song.video, image=song.thumb
                )
                await put_queue(
                    chat_id=chat_id,
                    original_chat_id=original_chat_id,
                    user_id=user_id,
                    track=song,
                    forceplay=forceplay,
                )
                thumb = await gen_thumb(track.vidid, track.thumb)
                what, button = play_markup(_, chat_id, track)
                run = await tbot.send_file(
                    original_chat_id,
                    file=img,
                    caption=_["stream_1"].format(
                        song.title[:27],
                        f"https://t.me/{tbot.username}?start=info_{song.vidid}",
                        seconds_to_min(song.duration),
                        user_mention,
                    ),
                    buttons=button,
                )
                db[chat_id][0]["mystic"] = run
                db[chat_id][0]["markup"] = what
        if count == 0:
            return
        else:
            link = await paste(msg)
            lines = msg.count("\n")
            if lines >= 17:
                car = os.linesep.join(msg.split(os.linesep)[:17])
            else:
                car = msg
            carbon = await carbon.generate(car, randint(100, 10000000))
            upl = close_markup(_)
            return await tbot.send_file(
                original_chat_id,
                file=carbon,
                caption=_["playlist_18"].format(link, position),
                buttons=upl,
            )

    elif not (track.is_live or track.is_m3u8):
        if not track.duration:
            return
        if await is_active_chat(chat_id):
            await put_queue(
                chat_id=chat_id,
                original_chat_id=original_chat_id,
                user_id=user_id,
                track=track,
            )
            is_queue_ = True
        else:
            if not forceplay:
                db[chat_id] = []
            try:
                file_path = await track.download()
            except Exception as e:
                await tbot.handle_error(e)
                raise AssistantErr(_["play_16"])
            await Yukki.join_call(chat_id, file_path, video=video, image=track.thumb)
            await put_queue(
                chat_id=chat_id,
                original_chat_id=original_chat_id,
                user_id=user_id,
                track=track,
                forceplay=forceplay,
            )
    elif track.is_live or track.is_m3u8:
        if await is_active_chat(chat_id):
            await put_queue(
                chat_id=chat_id,
                original_chat_id=original_chat_id,
                user_id=user_id,
                track=track,
            )
            is_queue_ = True
        else:
            if not forceplay:
                db[chat_id] = []
            try:
                file_path = await track.download()
            except Exception as e:
                await tbot.handle_error(e)
                raise AssistantErr(_["str_3"]) from e
            await Yukki.join_call(
                chat_id,
                file_path,
                video=video,
                image=track.thumb,
            )
            await put_queue(
                chat_id=chat_id,
                original_chat_id=original_chat_id,
                user_id=user_id,
                track=track,
                forceplay=forceplay,
            )
    title = track.title or "Index or M3u8 Link"
    link = (
            f"https://t.me/{tbot.username}?start=info_{track.vidid}"
            if track.vidid
            else track.link
        )
    duration = seconds_to_min(track.duration) if track.duration else "00:00"        
    if is_queue_:
        photo = await gen_qthumb(track.vidid, track.thumb)
        caption = _["queue_4"].format(
            len(db.get(chat_id)) - 1,
            title[:30],
            duration,
            user_mention,
        )
        button = close_markup(_)
    else:
        photo = await gen_thumb(track.vidid, track.thumb)
        caption = (
            _["stream_1"].format(
                title[:27],
                link,
                duration,
                user_mention,
            ),
        )
        what, button = play_markup(_, chat_id, track)
    run = await tbot.send_message(
        original_chat_id,
        file=photo,
        message=caption,
        buttons=button,
    )
    if not is_queue_:
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = what
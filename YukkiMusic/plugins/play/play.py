#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import logging
import random
import string

import config
from config import BANNED_USERS, lyrical
from YukkiMusic import tbot
from YukkiMusic.core.enum import SourceType
from YukkiMusic.core.track import Track
from YukkiMusic.platforms import (
    apple,
    resso,
    saavn,
    soundcloud,
    telegram,
    youtube,
)
from YukkiMusic.utils import seconds_to_min, time_to_seconds
from YukkiMusic.utils.database import is_video_allowed
from YukkiMusic.utils.decorators.play import play_wrapper
from YukkiMusic.utils.formatters import formats
from YukkiMusic.utils.inline.play import (
    livestream_markup,
    playlist_markup,
    slider_markup,
    track_markup,
)
from YukkiMusic.utils.inline.playlist import botplaylist_markup
from YukkiMusic.utils.logger import play_logs
from YukkiMusic.utils.stream.stream import stream

logger = logging.getLogger(__name__)


@tbot.on_message(flt.command("PLAY_COMMAND", True) & flt.group & ~BANNED_USERS)
@play_wrapper
async def play_commnd(
    event,
    _,
    chat_id,
    video,
    channel,
    playmode,
    url,
    fplay,
):
    mystic = await event.reply(_["play_2"].format(channel) if channel else _["play_1"])
    plist_id = None
    slider = None
    plist_type = None
    spotify = None
    user_id = event.sender_id
    # user_name = message.from_user.mention
    file = None
    audio_telegram, video_telegram = None, None
    if event.is_reply:
        rmsg = await event.get_reply_message()
        file = rmsg.file
        audio_telegram = rmsg.audio or rmsg.voice
        video_telegram = rmsg.video

    if audio_telegram:
        if file.size > config.TG_AUDIO_FILESIZE_LIMIT:
            return await mystic.edit(_["play_5"])
        duration_min = seconds_to_min(file.duration)
        if (file.duration) > config.DURATION_LIMIT:
            return await mystic.edit(
                _["play_6"].format(config.DURATION_LIMIT_MIN, duration_min)
            )
        if file_path := await telegram.download(_, rmsg, mystic):
            message_link = await telegram.get_link(rmsg)
            file_name = await telegram.get_filename(file)
            details = Track(
                title=file_name,
                link=message_link,
                thumb=config.TELEGRAM_AUDIO_URL,
                duration=file.duration,
                streamtype=SourceType.TELEGRAM,
                video=False,
                file_path=file_path,
            )
    elif video_telegram:
        if not await is_video_allowed(event.chat_id):
            return await mystic.edit(_["play_3"])
        try:
            if file.ext.lower() not in formats:
                return await mystic.edit(_["play_8"].format(f"{' | '.join(formats)}"))
        except Exception:
            return await mystic.edit(_["play_8"].format(f"{' | '.join(formats)}"))
        if file.size > config.TG_VIDEO_FILESIZE_LIMIT:
            return await mystic.edit(_["play_9"])
        if await telegram.download(_, rmsg, mystic, True):
            message_link = await telegram.get_link(rmsg)
            file_name = await telegram.get_filename(file)
            details = Track(
                title=file_name,
                link=message_link,
                thumb=config.TELEGRAM_VIDEO_URL,
                duration=file.duration,
                streamtype=SourceType.TELEGRAM,
                video=True,
                file_path=file_path,
            )

    elif url:
        if await youtube.valid(url):
            if "playlist" in url:
                try:
                    details = await youtube.playlist(
                        url,
                        config.PLAYLIST_FETCH_LIMIT,
                    )
                except Exception:
                    logger.error("", exc_info=True)
                    return await mystic.edit(_["play_3"])

                plist_type = "yt"
                if "&" in url:
                    plist_id = (url.split("=")[1]).split("&")[0]
                else:
                    plist_id = url.split("=")[1]
                img = details[0].thumb
                cap = _["play_10"]
            elif "https://youtu.be" in url:
                videoid = url.split("/")[-1].split("?")[0]
                details, track_id = await youtube.track(
                    f"https://www.youtube.com/watch?v={videoid}"
                )
                streamtype = "youtube"
                img = details["thumb"]
                cap = _["play_11"].format(
                    details["title"],
                    details["duration_min"],
                )
            else:
                try:
                    details, track_id = await youtube.track(url)
                except Exception as e:
                    print(e)
                    return await mystic.edit(_["play_3"])
                streamtype = "youtube"
                img = details["thumb"]
                cap = _["play_11"].format(
                    details["title"],
                    details["duration_min"],
                )
        elif await spotify.valid(url):
            spotify = True
            if not config.SPOTIFY_CLIENT_ID and not config.SPOTIFY_CLIENT_SECRET:
                return await mystic.edit(
                    "This Bot can't play spotify tracks and playlist, please contact my owner and ask him to add Spotify player."
                )
            if "track" in url:
                try:
                    details, track_id = await spotify.track(url)
                except Exception:
                    return await mystic.edit(_["play_3"])
                streamtype = "youtube"
                img = details["thumb"]
                cap = _["play_11"].format(details["title"], details["duration_min"])
            elif "playlist" in url:
                try:
                    details, plist_id = await spotify.playlist(url)
                except Exception:
                    return await mystic.edit(_["play_3"])
                streamtype = "playlist"
                plist_type = "spplay"
                img = config.SPOTIFY_PLAYLIST_IMG_URL
                cap = _["play_12"].format(message.from_user.first_name)
            elif "album" in url:
                try:
                    details, plist_id = await spotify.album(url)
                except Exception:
                    return await mystic.edit(_["play_3"])
                streamtype = "playlist"
                plist_type = "spalbum"
                img = config.SPOTIFY_ALBUM_IMG_URL
                cap = _["play_12"].format(message.from_user.first_name)
            elif "artist" in url:
                try:
                    details, plist_id = await spotify.artist(url)
                except Exception:
                    return await mystic.edit(_["play_3"])
                streamtype = "playlist"
                plist_type = "spartist"
                img = config.SPOTIFY_ARTIST_IMG_URL
                cap = _["play_12"].format(message.from_user.first_name)
            else:
                return await mystic.edit(_["play_17"])
        elif await apple.valid(url):
            if "album" in url:
                try:
                    details, track_id = await apple.track(url)
                except Exception:
                    return await mystic.edit(_["play_3"])
                streamtype = "youtube"
                img = details["thumb"]
                cap = _["play_11"].format(details["title"], details["duration_min"])
            elif "playlist" in url:
                spotify = True
                try:
                    details, plist_id = await apple.playlist(url)
                except Exception:
                    return await mystic.edit(_["play_3"])
                streamtype = "playlist"
                plist_type = "apple"
                cap = _["play_13"].format(message.from_user.first_name)
                img = url
            else:
                return await mystic.edit(_["play_16"])
        elif await resso.valid(url):
            try:
                details, track_id = await resso.track(url)
            except Exception:
                return await mystic.edit(_["play_3"])
            streamtype = "youtube"
            img = details["thumb"]
            cap = _["play_11"].format(details["title"], details["duration_min"])
        elif await saavn.valid(url):
            if "shows" in url:
                return await mystic.edit(_["saavn_1"])

            elif await saavn.is_song(url):
                try:
                    file_path, details = await saavn.download(url)
                except Exception as e:
                    ex_type = type(e).__name__
                    logger.error("An error occurred", exc_info=True)
                    return await mystic.edit(_["play_3"])
                duration_sec = details["duration_sec"]
                streamtype = "saavn_track"

                if duration_sec > config.DURATION_LIMIT:
                    return await mystic.edit(
                        _["play_6"].format(
                            config.DURATION_LIMIT_MIN,
                            details["duration_min"],
                        )
                    )
            elif await saavn.is_playlist(url):
                try:
                    details = await saavn.playlist(
                        url, limit=config.PLAYLIST_FETCH_LIMIT
                    )
                    streamtype = "saavn_playlist"
                except Exception as e:
                    ex_type = type(e).__name__
                    logger.error("An error occurred", exc_info=True)
                    return await mystic.edit(_["play_3"])
        elif await soundcloud.valid(url):
            try:
                details, track_path = await soundcloud.download(url)
            except Exception:
                return await mystic.edit(_["play_3"])
            duration_sec = details["duration_sec"]
            if duration_sec > config.DURATION_LIMIT:
                return await mystic.edit(
                    _["play_6"].format(
                        config.DURATION_LIMIT_MIN,
                        details["duration_min"],
                    )
                )
            try:
                await stream(
                    _,
                    mystic,
                    user_id,
                    details,
                    chat_id,
                    user_name,
                    event.chat_id,
                    streamtype="soundcloud",
                    forceplay=fplay,
                )
            except Exception as e:
                ex_type = type(e).__name__
                if ex_type == "AssistantErr":
                    err = e
                else:
                    logger.error("An error occurred", exc_info=True)
                    err = _["general_3"].format(ex_type)
                return await mystic.edit(err)
            return await mystic.delete()
        else:
            if not await telegram.is_streamable_url(url):
                return await mystic.edit(_["play_19"])

            await mystic.edit(_["str_2"])
            try:
                await stream(
                    _,
                    mystic,
                    event.sender_id,
                    url,
                    chat_id,
                    message.from_user.first_name,
                    event.chat_id,
                    video=video,
                    streamtype="index",
                    forceplay=fplay,
                )
            except Exception as e:
                ex_type = type(e).__name__
                if ex_type == "AssistantErr":
                    err = e
                else:
                    logger.error("An error occurred", exc_info=True)
                    err = _["general_3"].format(ex_type)
                return await mystic.edit(err)
            return await play_logs(message, streamtype="M3u8 or Index Link")
    else:
        if len(message.command) < 2:
            buttons = botplaylist_markup(_)
            return await mystic.edit(
                _["playlist_1"],
                buttons=buttons,
            )
        slider = True
        query = message.text.split(None, 1)[1]
        if "-v" in query:
            query = query.replace("-v", "")
        try:
            details, track_id = await youtube.track(query)
        except Exception:
            return await mystic.edit(_["play_3"])
        streamtype = "youtube"
    if str(playmode) == "DIRECT" and not plist_type:
        if details["duration_min"]:
            duration_sec = time_to_seconds(details["duration_min"])
            if duration_sec > config.DURATION_LIMIT:
                return await mystic.edit(
                    _["play_6"].format(
                        config.DURATION_LIMIT_MIN,
                        details["duration_min"],
                    )
                )
        else:
            buttons = livestream_markup(
                _,
                track_id,
                user_id,
                "v" if video else "a",
                "c" if channel else "g",
                "f" if fplay else "d",
            )
            return await mystic.edit(
                _["play_15"],
                buttons=buttons,
            )
        try:
            await stream(
                _,
                mystic,
                user_id,
                details,
                chat_id,
                user_name,
                event.chat_id,
                video=video,
                streamtype=streamtype,
                spotify=spotify,
                forceplay=fplay,
            )
        except Exception as e:
            ex_type = type(e).__name__
            if ex_type == "AssistantErr":
                err = e
            else:
                logger.error("An error occurred", exc_info=True)

                err = _["general_3"].format(ex_type)
            return await mystic.edit(err)
        await mystic.delete()
        return await play_logs(message, streamtype=streamtype)
    else:
        if plist_type:
            ran_hash = "".join(
                random.choices(string.ascii_uppercase + string.digits, k=10)
            )
            lyrical[ran_hash] = plist_id
            buttons = playlist_markup(
                _,
                ran_hash,
                event.sender_id,
                plist_type,
                "c" if channel else "g",
                "f" if fplay else "d",
            )
            await mystic.delete()
            await message.reply_photo(
                photo=img,
                caption=cap,
                buttons=buttons,
            )
            return await play_logs(message, streamtype=f"Playlist : {plist_type}")
        else:
            if slider:
                buttons = slider_markup(
                    _,
                    track_id,
                    event.sender_id,
                    query,
                    0,
                    "c" if channel else "g",
                    "f" if fplay else "d",
                )
                await mystic.delete()
                await message.reply_photo(
                    photo=details["thumb"],
                    caption=_["play_11"].format(
                        details["title"].title(),
                        details["duration_min"],
                    ),
                    buttons=buttons,
                )
                return await play_logs(message, streamtype=f"Searched on Youtube")
            else:
                buttons = track_markup(
                    _,
                    track_id,
                    event.sender_id,
                    "c" if channel else "g",
                    "f" if fplay else "d",
                )
                await mystic.delete()
                await message.reply_photo(
                    photo=img,
                    caption=cap,
                    buttons=buttons,
                )
                return await play_logs(message, streamtype=f"URL Searched Inline")

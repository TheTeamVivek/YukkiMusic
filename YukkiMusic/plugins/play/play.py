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
import traceback

import config
from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
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
from YukkiMusic.utils import get_message_link, seconds_to_min, time_to_seconds
from YukkiMusic.utils.database import is_video_allowed
from YukkiMusic.utils.decorators.play import play_wrapper
from YukkiMusic.utils.formatters import formats
from YukkiMusic.utils.inline.play import livestream_markup
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
            message_link = await get_message_link(rmsg)
            file_name = file.name or "Telagram audio file"
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
            message_link = await get_message_link(rmsg)
            file_name = file.name or "Telagram video file"
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

            if "https://youtu.be" in url:
                videoid = url.split("/")[-1].split("?")[0]
                details = await youtube.track(
                    f"https://www.youtube.com/watch?v={videoid}"
                )
            else:
                try:
                    details = await youtube.track(url)
                except Exception:
                    traceback.print_exc()
                    return await mystic.edit(_["play_3"])

        elif await spotify.valid(url):
            if not config.SPOTIFY_CLIENT_ID and not config.SPOTIFY_CLIENT_SECRET:
                return await mystic.edit(_["spotify_1"])
            try:
                details = await spotify.track(url)
            except Exception:
                traceback.print_exc()
                return await mystic.edit(_["play_3"])

            if details is None:
                return await mystic.edit(_["play_17"])

        elif await apple.valid(url):
            try:
                details = await apple.track(url)
            except Exception:
                traceback.print_exc()
                return await mystic.edit(_["play_3"])
            if details is None:
                return await mystic.edit(_["play_16"])

        elif await resso.valid(url):
            try:
                details, track_id = await resso.track(url)
            except Exception:
                traceback.print_exc()
                return await mystic.edit(_["play_3"])

        elif await saavn.valid(url):
            if "shows" in url:
                return await mystic.edit(_["saavn_1"])

            try:
                details = await saavn.track(url)
            except Exception:
                traceback.print_exc()
                return await mystic.edit(_["play_3"])

        elif await soundcloud.valid(url):
            try:
                details = await soundcloud.details(url)
            except Exception:
                traceback.print_exc()
                return await mystic.edit(_["play_3"])

        else:
            if not await telegram.is_streamable_url(url):
                return await mystic.edit(_["play_19"])
            details = track = Track(
                title="M3U8 or index Urls",
                link=url,
                thumb=config.STREAM_IMG_URL,
                duration=0,
                streamtype=SourceType.M3U8,
                video=video,
            )
    else:
        if len(message.command) < 2:
            buttons = botplaylist_markup(_)
            return await mystic.edit(
                _["playlist_1"],
                buttons=buttons,
            )
        query = message.text.split(None, 1)[1]
        if "-v" in query:
            query = query.replace("-v", "")
        try:
            details = await youtube.track(query)
        except Exception:
            traceback.print_exc()
            return await mystic.edit(_["play_3"])

    # if str(playmode) == "DIRECT" and not plist_type:
    if True:
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
                chat_id,
                event.chat_id,
                track=details,
                user_id=user_id,
                forceplay=fplay,
            )

        except Exception as e:
            ex_type = type(e).__name__
            if ex_type == "AssistantErr":
                err = e
            else:
                traceback.print_exc()
                err = _["general_3"].format(ex_type)
            return await mystic.edit(err)
        await mystic.delete()
        return await play_logs(message, streamtype=streamtype)
    """else:
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
                return await play_logs(message, streamtype=f"URL Searched Inline")"""

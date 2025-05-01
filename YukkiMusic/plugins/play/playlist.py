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
import os
from random import randint

from pykeyboard.telethon import InlineKeyboard
from telethon import Button, events

import config
from config import SERVER_PLAYLIST_LIMIT
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.track import Track
from YukkiMusic.misc import BANNED_USERS
from YukkiMusic.platforms import carbon, youtube
from YukkiMusic.utils import (
    delete_playlist,
    get_playlist,
    get_playlist_names,
    language,
    parse_flags,
    paste,
    save_playlist,
    stream,
)
from YukkiMusic.utils.inline import (
    botplaylist_markup,
    get_playlist_markup,
    warning_markup,
)

logger = logging.getLogger(__name__)


async def get_keyboard(_, user_id):
    keyboard = InlineKeyboard(row_width=5)
    _playlist = await get_playlist_names(user_id)
    count = len(_playlist)
    for x in _playlist:
        _note = await get_playlist(user_id, x)
        title = _note["title"]
        title = title.title()
        keyboard.row(
            Button.inline(
                text=title,
                data=f"del_playlist {x}",
            )
        )
    keyboard.row(
        Buton.inline(
            text=_["PL_B_5"],
            data=f"delete_warning",
        ),
        Button.inline(text=_["CLOSE_BUTTON"], data=f"close"),
    )
    return keyboard, count


@tbot.on_message(flt.command("PLAYLIST_COMMAND", True) & flt.group & ~BANNED_USERS)
@language
async def check_playlist(event, _):
    _playlist = await get_playlist_names(event.sender_id)
    if _playlist:
        get = await event.reply(_["playlist_2"])
    else:
        return await event.reply(_["playlist_3"])
    msg = _["playlist_4"]
    count = 0
    for ptlist in _playlist:
        _note = await get_playlist(event.sender_id, ptlist)
        title = _note["title"]
        title = title.title()
        duration = _note["duration"]
        count += 1
        msg += f"\n\n{count}- {title[:70]}\n"
        msg += _["playlist_5"].format(duration)
    link = await paste(msg)
    lines = msg.count("\n")
    if lines >= 17:
        car = os.linesep.join(msg.split(os.linesep)[:17])
    else:
        car = msg
    carbon = await carbon.generate(car, randint(100, 10000000000))
    await get.delete()
    await event.reply(file=carbon, message=_["playlist_15"].format(link))


@tbot.on_message(flt.command("DELETE_PLAYLIST_COMMAND", True) & ~BANNED_USERS)
@language
async def del_group_message(event, _):
    if event.is_group:
        upl = Button.inline(
            _["PL_B_6"], f"https://t.me/{tbot.username}?start=delplaylists"
        )
        await event.reply(_["playlist_6"], buttons=upl)

    elif event.is_private:
        _playlist = await get_playlist_names(event.sender_id)
        if _playlist:
            get = await event.reply(_["playlist_2"])
        else:
            return await event.reply(_["playlist_3"])
        keyboard, count = await get_keyboard(_, event.sender_id)
        await get.edit(_["playlist_7"].format(count), buttons=keyboard)


@tbot.on_message(
    flt.command("ADD_PLAYLIST_COMMAND", True) & ~BANNED_USERS
)  # TODO: ADD SUPPORT FOR SPOTIFY, RESSO, APPLE
@language
async def add_playlist(event, _):
    user_id = event.sender_id
    command = event.text.split()
    if len(command) < 2:
        return await event.reply(_["playlist_22"])
    query = event.text.replace(command[0], "")

    if "youtube.com/playlist" in query:
        adding = await event.reply(_["playlist_21"])
        try:
            results = await youtube.playlist(query, config.SERVER_PLAYLIST_LIMIT)
            count = len(await get_playlist_names(user_id))
            for x in results:
                if count == SERVER_PLAYLIST_LIMIT:
                    break

                if isinstance(x, Track):
                    _check = await get_playlist(user_id, x.vidid)
                    if _check:
                        continue
                    t = x
                else:
                    _check = await get_playlist(user_id, x)
                    if _check:
                        continue
                    t = await youtube.track(youtube.base + x)

                video_info = {
                    "videoid": t.vidid,
                    "title": t.title,
                    "duration": t.duration,
                }
                await save_playlist(user_id, t.vidid, video_info)
                count += 1

        except Exception:
            return await event.reply(
                f"Looking like not a valid youtube playlist url or\nPlaylist created by YouTube Not Supported"
            )

        await adding.delete()
        return await event.reply(_["playlist_20"])
    else:
        try:
            track = await youtube.track(query)
            check = await get_playlist(user_id, track.vidid)
            if check:
                return await event.reply(_["playlist_8"])

            count = len(await get_playlist_names(user_id))
            if count == SERVER_PLAYLIST_LIMIT:
                return await event.reply(
                    _["playlist_9"].format(config.SERVER_PLAYLIST_LIMIT)
                )

            m = await event.reply(_["playlist_21"])

            plist = {
                "videoid": track.vidid,
                "title": track.title[:30],
                "duration": track.duration,
            }

            await save_playlist(user_id, track.vidid, plist)
            await m.delete()
            await event.reply(file=thumbnail, message=_["playlist_20"])
        except Exception:
            logger.info("", exc_info=True)
            return await event.reply("**Something wrong happens **\nSee Logs")


@tbot.on(events.CallbackQuery(pattern="add_playlist", func=~BANNED_USERS))
@language
async def add_playlist(event, _):
    callback_data = event.data.decode("utf-8").strip()
    videoid = callback_data.split(None, 1)[1]
    user_id = event.sender_id
    _check = await get_playlist(user_id, videoid)
    if _check:
        try:
            return await event.answer(_["playlist_8"], alert=True)
        except Exception:
            return
    _count = await get_playlist_names(user_id)
    count = len(_count)
    if count == SERVER_PLAYLIST_LIMIT:
        try:
            return await event.answer(
                _["playlist_9"].format(SERVER_PLAYLIST_LIMIT),
                alert=True,
            )
        except Exception:
            return
    track = await youtube.track(videoid)
    title = (track.title[:25]).title()
    plist = {
        "videoid": track.vidid,
        "title": title,
        "duration": track.duration,
    }
    await save_playlist(user_id, videoid, plist)
    try:
        title = (title[:30]).title()
        return await event.answer(_["playlist_10"].format(title), alert=True)
    except Exception:
        return


@tbot.on(events.CallbackQuery(pattern="del_playlist", func=~BANNED_USERS))
@language
async def del_plist(event, _):
    callback_data = event.data.decode("utf-8").strip()
    videoid = callback_data.split(None, 1)[1]
    user_id = event.sender_id
    deleted = await delete_playlist(event.sender_id, videoid)
    if deleted:
        try:
            await event.answer(_["playlist_11"], alert=True)
        except Exception:
            pass
    else:
        try:
            return await event.answer(_["playlist_12"], alert=True)
        except Exception:
            return
    keyboard, count = await get_keyboard(_, user_id)
    return await event.edit(buttons=keyboard)


@tbot.on(events.CallbackQuery(pattern="delete_whole_playlist", func=~BANNED_USERS))
@language
async def del_whole_playlist(event, _):
    _playlist = await get_playlist_names(event.sender_id)
    await event.answer(_["playlist_25"], alert=True)
    for x in _playlist:
        await delete_playlist(event.sender_id, x)
    return await event.edit(_["playlist_13"])


@tbot.on(events.CallbackQuery(pattern="del_back_playlist", func=~BANNED_USERS))
@language
async def del_back_playlist(event, _):
    user_id = event.sender_id
    _playlist = await get_playlist_names(user_id)
    if _playlist:
        try:
            await event.answer(_["playlist_2"], alert=True)
        except Exception:
            pass
    else:
        try:
            return await event.answer(_["playlist_3"], alert=True)
        except Exception:
            return
    keyboard, count = await get_keyboard(_, user_id)
    return await event.edit(_["playlist_7"].format(count), buttons=keyboard)


@tbot.on(events.CallbackQuery(pattern="get_playlist_playmode", func=~BANNED_USERS))
@tbot.on(events.CallbackQuery(pattern="home_play", func=~BANNED_USERS))
@tbot.on(events.CallbackQuery(pattern="delete_warning", func=~BANNED_USERS))
@language
async def playlist_multi_func(event, _):
    name = event.data.decode("utf-8").strip()
    try:
        await event.answer()
    except Exception:
        pass
    if name.startswith("get_playlist_playmode"):
        buttons = get_playlist_markup(_)

    elif name.startswith("home_play"):
        buttons = botplaylist_markup(_)

    elif name.startswith("delete_warning"):
        upl = warning_markup(_)
        return await event.edit(_["playlist_14"], buttons=upl)

    return await event.edit(buttons=buttons)


@tbot.on(events.CallbackQuery(pattern="play_playlist", func=~BANNED_USERS))
@language
async def play_playlist(event, _):
    callback_data = event.data.decode("utf-8").strip()
    mode = callback_data.split(None, 1)[1]
    user_id = event.sender_id
    _playlist = await get_playlist_names(user_id)
    if not _playlist:
        try:
            return await event.answer(
                _["playlist_3"],
                alert=True,
            )
        except Exception:
            return
    chat_id = event.chat_id
    await event.delete()
    result = []
    try:
        await event.answer()
    except Exception:
        pass
    mystic = await event.reply(_["play_1"])
    for vidids in _playlist:
        result.append(vidids)
    result.insert(0, await youtube.track(youtube.base + result.pop(0), mode == "v"))
    try:
        await stream(
            chat_id=chat_id,
            original_chat_id=chat_id,
            track=result,
            user_id=user_id,
        )
    except Exception as e:
        ex_type = type(e).__name__
        if ex_type == "AssistantErr":
            err = e
        else:
            err = _["general_3"].format(ex_type)
            logger.error("", exc_info=True)
        return await mystic.edit(err)
    return await mystic.delete()


@tbot.on_message(flt.command("PLAY_PLAYLIST_COMMAND", True) & ~BANNED_USERS & flt.group)
@language
async def play_playlist_command(event, _):
    video, fplay, cplay = parse_flags(event.text, "playplaylist")
    if cplay:
        chat_id = await get_cmode(event.chat_id)
        if chat_id is None:
            return await event.reply(_["setting_12"])
        try:
            await tbot.get_entity(chat_id)
        except Exception:
            return await event.reply(_["cplay_4"])
    else:
        chat_id = event.chat_id

    user_id = event.sender_id
    _playlist = await get_playlist_names(user_id)
    if not _playlist:
        try:
            return await event.reply(
                _["playlist_3"],
                quote=True,
            )
        except Exception:
            return

    try:
        await event.delete()
    except Exception:
        pass

    result = []
    mystic = await event.reply(_["play_1"])

    for vidids in _playlist:
        result.append(vidids)
    result.insert(0, await youtube.track(youtube.base + result.pop(0), video))

    try:
        await stream(
            chat_id=chat_id,
            original_chat_id=event.chat_id,
            track=result,
            user_id=user_id,
            forceplay=fplay,
        )
    except Exception as e:
        ex_type = type(e).__name__
        if ex_type == "AssistantErr":
            err = e
        else:
            err = _["general_3"].format(ex_type)
            logger.error("", exc_info=True)
        return await mystic.edit(err)
    return await mystic.delete()


@tbot.on(events.CallbackQuery(pattern="remove_playlist", func=~BANNED_USERS))
@language
async def del_plist(event, _):
    callback_data = event.data.decode("utf-8").strip()
    videoid = callback_data.split(None, 1)[1]
    deleted = await delete_playlist(event.sender_id, videoid)
    if deleted:
        try:
            await event.answer(_["playlist_11"], alert=True)
        except Exception:
            pass
    else:
        try:
            return await event.answer(_["playlist_12"], alert=True)
        except Exception:
            return

    return await event.edit(text=_["playlist_23"])

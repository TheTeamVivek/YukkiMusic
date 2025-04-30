#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import random

from telethon import events

import config
from YukkiMusic import tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import BANNED_USERS, SUDOERS, db
from YukkiMusic.platforms import apple, youtube
from YukkiMusic.utils import time_to_seconds
from YukkiMusic.utils.channelplay import get_channeplay_cb
from YukkiMusic.utils.database import (
    is_active_chat,
    is_music_playing,
    is_muted,
    is_nonadmin_chat,
    music_off,
    music_on,
    mute_off,
    mute_on,
    set_loop,
)
from YukkiMusic.utils.decorators import actual_admin_cb
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.formatters import seconds_to_min
from YukkiMusic.utils.inline.play import (
    livestream_markup,
    panel_markup_1,
    panel_markup_2,
    panel_markup_3,
    play_markup,
    slider_markup,
    stream_markup,
    telegram_markup,
)
from YukkiMusic.utils.stream.autoclear import auto_clean
from YukkiMusic.utils.stream.stream import stream

wrong = {}


@tbot.on(events.CallbackQuery(pattern="PanelMarkup", func=~BANNED_USERS))
@language
async def markup_panel(event, _):
    await event.answer()
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    videoid, chat_id = callback_request.split("|")
    chat_id = event.chat_id
    buttons = panel_markup_1(_, videoid, chat_id)
    try:
        await event.edit(buttons=buttons)
    except Exception:
        return
    if chat_id not in wrong:
        wrong[chat_id] = {}
    wrong[chat_id][event.message_id] = False


@tbot.on(events.CallbackQuery(pattern="MainMarkup", func=~BANNED_USERS))
@language
async def main_markup_(event, _):
    await event.answer()
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    videoid, chat_id = callback_request.split("|")
    if videoid == str(None):
        buttons = telegram_markup(_, chat_id)
    else:
        buttons = stream_markup(_, videoid, chat_id)
    chat_id = event.chat_id
    try:
        await event.edit(buttons=buttons)
    except Exception:
        return
    if chat_id not in wrong:
        wrong[chat_id] = {}
    wrong[chat_id][event.message_id] = True


@tbot.on(events.CallbackQuery(pattern="Pages", func=~BANNED_USERS))
@language
async def pages_markup(event, _):
    await event.answer()
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    state, pages, videoid, chat = callback_request.split("|")
    chat_id = int(chat)
    pages = int(pages)
    if state == "Forw":
        if pages == 0:
            buttons = panel_markup_2(_, videoid, chat_id)
        if pages == 2:
            buttons = panel_markup_1(_, videoid, chat_id)
        if pages == 1:
            buttons = panel_markup_3(_, videoid, chat_id)
    if state == "Back":
        if pages == 2:
            buttons = panel_markup_2(_, videoid, chat_id)
        if pages == 1:
            buttons = panel_markup_1(_, videoid, chat_id)
        if pages == 0:
            buttons = panel_markup_3(_, videoid, chat_id)
    try:
        await event.edit(buttons=buttons)
    except Exception:
        return


@tbot.on(events.CallbackQuery(pattern="ADMIN", func=~BANNED_USERS))
@language
async def admin_callback(event, _):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    command, chat = callback_request.split("|")
    chat_id = int(chat)
    command = command.lower().strip()
    if not await is_active_chat(chat_id):
        return await event.answer(_["general_6"], alert=True)
    sender = await event.get_sender()
    mention = await tbot.create_mention(sender)
    is_non_admin = await is_nonadmin_chat(event.chat_id)
    if not is_non_admin:
        if sender.id not in SUDOERS:
            admins = config.adminlist.get(event.chat_id)
            if not admins:
                return await event.answer(_["admin_18"], alert=True)
            else:
                if sender.id not in admins:
                    return await event.answer(_["admin_19"], alert=True)
    if command == "pause":
        if not await is_music_playing(chat_id):
            return await event.answer(_["admin_1"], alert=True)
        await event.answer()
        await music_off(chat_id)
        await Yukki.pause_stream(chat_id)
        await event.reply(_["admin_2"].format(mention), link_preview=False)
    elif command == "resume":
        if await is_music_playing(chat_id):
            return await event.answer(_["admin_3"], alert=True)
        await event.answer()
        await music_on(chat_id)
        await Yukki.resume_stream(chat_id)
        await event.reply(_["admin_4"].format(mention), link_preview=False)
    elif command == "stop" or command == "end":
        try:
            check = db.get(chat_id)
            if check[0].get("mystic"):
                await check[0].get("mystic").delete()
        except Exception:
            pass
        await event.answer()
        await Yukki.stop_stream(chat_id)
        await set_loop(chat_id, 0)
        await event.reply(_["admin_9"].format(mention), link_preview=False)
    elif command == "mute":
        if await is_muted(chat_id):
            return await event.answer(_["admin_5"], alert=True)
        await event.answer()
        await mute_on(chat_id)
        await Yukki.mute_stream(chat_id)
        await event.reply(_["admin_6"].format(mention), link_preview=False)
    elif command == "unmute":
        if not await is_muted(chat_id):
            return await event.answer(_["admin_7"], alert=True)
        await event.answer()
        await mute_off(chat_id)
        await Yukki.unmute_stream(chat_id)
        await event.reply(_["admin_8"].format(mention), link_preview=False)
    elif command == "loop":
        await event.answer()
        await set_loop(chat_id, 3)
        await event.reply(_["admin_26"].format(mention, 3))

    elif command == "shuffle":
        check = db.get(chat_id)
        if not check:
            return await event.answer(_["admin_22"], alert=True)
        try:
            popped = check.pop(0)
        except Exception:
            return await event.answer(_["admin_23"], alert=True)
        check = db.get(chat_id)
        if not check:
            check.insert(0, popped)
            return await event.answer(_["admin_23"], alert=True)
        await event.answer()
        random.shuffle(check)
        check.insert(0, popped)
        await event.reply(_["admin_24"].format(mention), link_preview=False)

    elif command in ["skip", "replay"]:
        check = db.get(chat_id)
        txt = ""
        if command == "skip":
            try:
                txt = _["admin_36"].format(mention)
                popped = check.pop(0)
                if popped:
                    await auto_clean(popped)
                if not check:
                    await event.edit(txt)
                    await event.reply(_["admin_10"].format(mention), link_preview=False)
                    return await Yukki.stop_stream(chat_id)

            except Exception:
                await event.edit(txt)
                await event.reply(_["admin_10"].format(mention), link_preview=False)
                return await Yukki.stop_stream(chat_id)
        else:
            txt = _["admin_36"].format(mention)

        await event.answer()
        track = check[0]["track"]
        user = check[0]["by"]
        url = (
            f"https://t.me/{tbot.username}?start=info_{track.vidid}"
            if track.is_youtube
            else track.link
        )
        db[chat_id][0]["played"] = 0
        mystic = await event.reply(_["call_8"], link_preview=False)

        try:
            file_path = await track.download()
            await Yukki.skip_stream(chat_id, file_path, video=track.video)

        except Exception as e:
            await tbot.handle_error(e, event)
            return await mystic.edit(_["call_7"])

        what, button = play_markup(_, chat_id, track)

        run = await event.respond(
            file=track.thumb,
            message=_["stream_1"].format(
                track.title[:27],
                url,
                seconds_to_min(track.duration),
                user,
            ),
            buttons=button,
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = what
        await event.edit(txt)
        await mystic.delete()

    # elif command == "seek":
    else:
        playing = db.get(chat_id)
        if not playing:
            return await event.answer(_["queue_2"], alert=True)
        duration_seconds = int(playing[0]["seconds"])
        if duration_seconds == 0:
            return await event.answer(_["admin_31"], alert=True)
        track = playing[0]["track"]

        if track.is_live or track.is_m3u8:
            return await event.answer(_["admin_31"], alert=True)

        duration_played = int(playing[0]["played"])
        if int(command) in [1, 2]:
            duration_to_skip = 10
        else:
            duration_to_skip = 30
        duration = seconds_to_min(track.duration)
        if int(command) in [1, 3]:
            if (duration_played - duration_to_skip) <= 10:
                bet = seconds_to_min(duration_played)
                return await event.answer(
                    _["admin_38"].format(bet, duration),
                    alert=True,
                )
            to_seek = duration_played - duration_to_skip + 1
        else:
            if (duration_seconds - (duration_played + duration_to_skip)) <= 10:
                bet = seconds_to_min(duration_played)
                return await event.answer(
                    _["admin_38"].format(bet, duration),
                    alert=True,
                )
            to_seek = duration_played + duration_to_skip + 1
        await event.answer()
        mystic = await event.reply(_["admin_33"])
        try:
            file_path = await track.download()
            await Yukki.seek_stream(
                chat_id,
                file_path,
                seconds_to_min(to_seek),
                duration,
                track.video,
            )
        except Exception as e:
            await tbot.handle_error(e, event)
            return await mystic.edit(_["admin_35"])
        if int(command) in [1, 3]:
            db[chat_id][0]["played"] -= duration_to_skip
        else:
            db[chat_id][0]["played"] += duration_to_skip
        string = _["admin_34"].format(seconds_to_min(to_seek))
        await mystic.edit(string)  # f"{string}\n\nChanges Done by: {mention} !")


@tbot.on(events.CallbackQuery(pattern="MusicStream", func=~BANNED_USERS))
@language
async def play_music(event, _):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    vidid, user_id, mode, cplay, fplay = callback_request.split("|")
    sender = await event.get_sender()
    if sender.id != int(user_id):
        try:
            return await event.answer(_["playcb_1"], alert=True)
        except Exception:
            return
    try:
        chat_id, channel = await get_channeplay_cb(_, cplay, event)
    except Exception:
        return
    user_name = sender.first_name
    try:
        await event.delete()
        await event.answer()
    except Exception:
        pass
    mystic = await event.reply(_["play_2"].format(channel) if channel else _["play_1"])
    try:
        details, track_id = await youtube.track(vidid, True)
    except Exception:
        return await mystic.edit(_["play_3"])
    if details["duration_min"]:
        duration_sec = time_to_seconds(details["duration_min"])
        if duration_sec > config.DURATION_LIMIT:
            return await mystic.edit(
                _["play_6"].format(config.DURATION_LIMIT_MIN, details["duration_min"])
            )
    else:
        buttons = livestream_markup(
            _,
            track_id,
            sender.id,
            mode,
            "c" if cplay == "c" else "g",
            "f" if fplay else "d",
        )
        return await mystic.edit(
            _["play_15"],
            buttons=buttons,
        )
    video = True if mode == "v" else None
    ffplay = True if fplay == "f" else None
    try:
        await stream(
            _,
            mystic,
            sender.id,
            details,
            chat_id,
            user_name,
            event.chat_id,
            video,
            streamtype="youtube",
            forceplay=ffplay,
        )
    except Exception as e:
        ex_type = type(e).__name__
        err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
        return await mystic.edit(err)
    return await mystic.delete()


@tbot.on(events.CallbackQuery(pattern="AnonymousAdmin", func=~BANNED_USERS))
@language
async def anonymous_check(event, _):
    try:
        await event.answer(
            _["anon_admin2"],
            alert=True,
        )
    except Exception:
        return


@tbot.on(events.CallbackQuery(pattern="YukkiPlaylists", func=~BANNED_USERS))
@language
async def play_playlists_cb(event, _):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    (
        videoid,
        user_id,
        ptype,
        mode,
        cplay,
        fplay,
    ) = callback_request.split("|")
    sender = await event.get_sender()
    if sender.id != int(user_id):
        try:
            return await event.answer(_["playcb_1"], alert=True)
        except Exception:
            return
    try:
        chat_id, channel = await get_channeplay_cb(_, cplay, event)
    except Exception:
        return
    user_name = sender.first_name
    await event.delete()
    try:
        await event.answer()
    except Exception:
        pass
    mystic = await event.reply(_["play_2"].format(channel) if channel else _["play_1"])
    videoid = config.lyrical.get(videoid)
    video = True if mode == "v" else None
    ffplay = True if fplay == "f" else None
    spotify = True
    if ptype == "yt":
        spotify = False
        try:
            result = await youtube.playlist(
                videoid,
                config.PLAYLIST_FETCH_LIMIT,
                True,
            )
        except Exception:
            return await mystic.edit(_["play_3"])
    if ptype == "spplay":
        try:
            result, spotify_id = await spotify.playlist(videoid)
        except Exception:
            return await mystic.edit(_["play_3"])
    if ptype == "spalbum":
        try:
            result, spotify_id = await spotify.album(videoid)
        except Exception:
            return await mystic.edit(_["play_3"])
    if ptype == "spartist":
        try:
            result, spotify_id = await spotify.artist(videoid)
        except Exception:
            return await mystic.edit(_["play_3"])
    if ptype == "apple":
        try:
            result, apple_id = await apple.playlist(videoid, True)
        except Exception:
            return await mystic.edit(_["play_3"])
    try:
        await stream(
            _,
            mystic,
            user_id,
            result,
            chat_id,
            user_name,
            event.chat_id,
            video,
            streamtype="playlist",
            spotify=spotify,
            forceplay=ffplay,
        )
    except Exception as e:
        ex_type = type(e).__name__
        err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
        return await mystic.edit(err)
    return await mystic.delete()


@tbot.on(events.CallbackQuery(pattern="slider", func=~BANNED_USERS))
@language
async def slider_queries(event, _):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    (
        what,
        rtype,
        query,
        user_id,
        cplay,
        fplay,
    ) = callback_request.split("|")
    sender = await event.get_sender()
    if sender.id != int(user_id):
        try:
            return await event.answer(_["playcb_1"], alert=True)
        except Exception:
            return
    what = str(what)
    rtype = int(rtype)
    if what == "F":
        if rtype == 9:
            query_type = 0
        else:
            query_type = int(rtype + 1)
        try:
            await event.answer(_["playcb_2"])
        except Exception:
            pass
        title, duration_min, thumbnail, vidid = await youtube.slider(
            query, query_type
        )  # todo use youtube.track
        buttons = slider_markup(_, vidid, user_id, query, query_type, cplay, fplay)
        return await event.edit(
            text=_["play_11"].format(
                title.title(),
                duration_min,
            ),
            file=thumbnail,
            buttons=buttons,
        )
    if what == "B":
        if rtype == 0:
            query_type = 9
        else:
            query_type = int(rtype - 1)
        try:
            await event.answer(_["playcb_2"])
        except Exception:
            pass
        title, duration_min, thumbnail, vidid = await youtube.slider(query, query_type)
        buttons = slider_markup(_, vidid, user_id, query, query_type, cplay, fplay)
        return await event.edit(
            text=_["play_11"].format(
                title.title(),
                duration_min,
            ),
            file=thumbnail,
            buttons=buttons,
        )


@tbot.on(events.CallbackQuery(pattern="close", func=~BANNED_USERS))
async def close_menu(event):
    await event.delete()
    await event.answer()


@tbot.on(events.CallbackQuery(pattern="stop_downloading", func=~BANNED_USERS))
@actual_admin_cb
async def stop_download(event, _):
    message_id = event.message_id
    task = config.lyrical.get(message_id)
    if not task or task.done():
        return await event.answer(_["tg_5"], alert=True)

    elif task.cancelled():
        return await event.answer(
            _["tg_6"],
            alert=True,
        )

    if not task.done():
        try:
            task.cancel()
            try:
                config.lyrical.pop(message_id)
            except Exception:
                pass
            await event.answer("Downloading Cancelled", alert=True)
            return await event.edit(
                f"Downloading cancelled by {await tbot.create_mention(await event.get_sender())}"
            )
        except Exception:
            return await event.answer("Failed to stop downloading", alert=True)

    await event.answer("Failed to Recognise Task", alert=True)

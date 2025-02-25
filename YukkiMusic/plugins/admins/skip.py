#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


import config
from config import BANNED_USERS
from YukkiMusic import Platform, tbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import db
from YukkiMusic.utils.database import get_loop
from YukkiMusic.utils.decorators import admin_rights_check
from YukkiMusic.utils.inline.play import stream_markup, telegram_markup
from YukkiMusic.utils.stream.autoclear import auto_clean
from YukkiMusic.utils.thumbnails import gen_thumb


@tbot.on_message(
    flt.command("SKIP_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS)
)
@admin_rights_check
async def skip(event, _, chat_id):
    mention = await tbot.create_mention(await event.get_sender())
    if not len(event.text.split()) < 2:
        loop = await get_loop(chat_id)
        if loop != 0:
            return await event.reply(_["admin_12"])
        state = event.text.split(None, 1)[1].strip()
        if state.isnumeric():
            state = int(state)
            check = db.get(chat_id)
            if check:
                count = len(check)
                if count > 2:
                    count = int(count - 1)
                    if 1 <= state <= count:
                        for x in range(state):
                            popped = None
                            try:
                                popped = check.pop(0)
                                if popped.get("mystic"):
                                    try:
                                        await popped.get("mystic").delete()
                                    except Exception:
                                        pass
                            except Exception:
                                return await event.reply(_["admin_16"])
                            if popped:
                                await auto_clean(popped)
                            if not check:
                                try:
                                    await event.reply(
                                        _["admin_10"].format(mention),
                                        link_preview=False,
                                    )
                                    await Yukki.stop_stream(chat_id)
                                except Exception:
                                    return
                                break
                    else:
                        return await event.reply(_["admin_15"].format(count))
                else:
                    return await event.reply(_["admin_14"])
            else:
                return await event.reply(_["queue_2"])
        else:
            return await event.reply(_["admin_13"])
    else:
        check = db.get(chat_id)
        popped = None
        try:
            popped = check.pop(0)
            if popped:
                await auto_clean(popped)
                if popped.get("mystic"):
                    try:
                        await popped.get("mystic").delete()
                    except Exception:
                        pass
            if not check:
                await event.reply(
                    _["admin_10"].format(mention),
                    link_preview=False,
                )
                try:
                    return await Yukki.stop_stream(chat_id)
                except Exception:
                    return
        except Exception:
            try:
                await event.reply(
                    _["admin_10"].format(mention),
                    link_preview=False,
                )
                return await Yukki.stop_stream(chat_id)
            except Exception:
                return
    queued = check[0]["file"]
    title = (check[0]["title"]).title()
    user = check[0]["by"]
    streamtype = check[0]["streamtype"]
    videoid = check[0]["vidid"]
    duration_min = check[0]["dur"]
    status = True if str(streamtype) == "video" else None
    if "live_" in queued:
        n, link = await Platform.youtube.video(videoid, True)
        if n == 0:
            return await event.reply(_["admin_11"].format(title))
        try:
            await Yukki.skip_stream(chat_id, link, video=status)
        except Exception:
            return await event.reply(_["STREAM_SWITCH_FAILED"])
        button = telegram_markup(_, chat_id)
        img = await gen_thumb(videoid)
        run = await event.reply(
            file=img,
            text=_["stream_1"].format(
                user,
                f"https://t.me/{tbot.username}?start=info_{videoid}",
            ),
            buttons=button,
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = "tg"
    elif "vid_" in queued:
        mystic = await event.reply(_["DOWNLOADING_NEXT_TRACK"], link_preview=False)
        try:
            file_path, direct = await Platform.youtube.download(
                videoid,
                mystic,
                videoid=True,
                video=status,
            )
        except Exception:
            return await mystic.edit(_["STREAM_SWITCH_FAILED"])
        try:
            await Yukki.skip_stream(chat_id, file_path, video=status)
        except Exception:
            return await mystic.edit(_["STREAM_SWITCH_FAILED"])
        button = stream_markup(_, videoid, chat_id)
        img = await gen_thumb(videoid)
        run = await event.reply(
            file=img,
            text=_["stream_1"].format(
                title[:27],
                f"https://t.me/{tbot.username}?start=info_{videoid}",
                duration_min,
                user,
            ),
            buttons=button,
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = "stream"
        await mystic.delete()
    elif "index_" in queued:
        try:
            await Yukki.skip_stream(chat_id, videoid, video=status)
        except Exception:
            return await event.reply(_["STREAM_SWITCH_FAILED"])
        button = telegram_markup(_, chat_id)
        run = await event.reply(
            file=config.STREAM_IMG_URL,
            text=_["stream_2"].format(user),
            buttons=button,
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = "tg"
    else:
        try:
            await Yukki.skip_stream(chat_id, queued, video=status)
        except Exception:
            return await event.reply(_["STREAM_SWITCH_FAILED"])
        if videoid == "telegram":
            button = telegram_markup(_, chat_id)
            run = await event.reply(
                file=(
                    config.TELEGRAM_AUDIO_URL
                    if str(streamtype) == "audio"
                    else config.TELEGRAM_VIDEO_URL
                ),
                text=_["stream_1"].format(
                    title, config.SUPPORT_GROUP, check[0]["dur"], user
                ),
                buttons=button,
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
        elif videoid == "soundcloud":
            button = telegram_markup(_, chat_id)
            run = await event.reply(
                file=(
                    config.SOUNCLOUD_IMG_URL
                    if str(streamtype) == "audio"
                    else config.TELEGRAM_VIDEO_URL
                ),
                text=_["stream_1"].format(
                    title, config.SUPPORT_GROUP, check[0]["dur"], user
                ),
                buttons=button,
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
        elif "saavn" in videoid:
            button = telegram_markup(_, chat_id)
            url = check[0]["url"]
            details = await Platform.saavn.info(url)
            run = await event.reply(
                file=details["thumb"] or config.TELEGRAM_AUDIO_URL,
                text=_["stream_1"].format(title, url, check[0]["dur"], user),
                buttons=button,
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
        else:
            button = stream_markup(_, videoid, chat_id)
            img = await gen_thumb(videoid)
            run = await event.reply(
                file=img,
                text=_["stream_1"].format(
                    title[:27],
                    f"https://t.me/{tbot.username}?start=info_{videoid}",
                    duration_min,
                    user,
                ),
                buttons=button,
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "stream"

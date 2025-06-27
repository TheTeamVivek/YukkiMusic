#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from pyrogram import filters
from pyrogram.types import InlineKeyboardMarkup, Message

import config
from config import BANNED_USERS
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.core.call import yukki
from yukkimusic.core.help import ModuleHelp
from yukkimusic.misc import db
from yukkimusic.platforms import saavn, youtube
from yukkimusic.utils import fallback
from yukkimusic.utils.database import get_loop
from yukkimusic.utils.decorators import AdminRightsCheck
from yukkimusic.utils.inline.play import stream_markup, telegram_markup
from yukkimusic.utils.stream.autoclear import auto_clean
from yukkimusic.utils.thumbnails import gen_thumb


@app.on_message(command("SKIP_COMMAND") & filters.group & ~BANNED_USERS)
@AdminRightsCheck
async def skip(cli, message: Message, _, chat_id):
    if not len(message.command) < 2:
        loop = await get_loop(chat_id)
        if loop != 0:
            return await message.reply_text(_["skip_3"])
        state = message.text.split(None, 1)[1].strip()
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
                                return await message.reply_text(_["skip_7"])
                            if popped:
                                await auto_clean(popped)
                            if not check:
                                try:
                                    await message.reply_text(
                                        _["skip_1"].format(
                                            message.from_user.first_name
                                        ),
                                    )
                                    await yukki.stop_stream(chat_id)
                                except Exception:
                                    return
                                break
                    else:
                        return await message.reply_text(_["skip_6"].format(count))
                else:
                    return await message.reply_text(_["skip_5"])
            else:
                return await message.reply_text(_["queue_2"])
        else:
            return await message.reply_text(_["skip_4"])
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
                await message.reply_text(
                    _["skip_1"].format(message.from_user.first_name),
                )
                try:
                    return await yukki.stop_stream(chat_id)
                except Exception:
                    return
        except Exception:
            try:
                await message.reply_text(
                    _["skip_1"].format(message.from_user.first_name),
                )
                return await yukki.stop_stream(chat_id)
            except Exception:
                return
    queued = check[0]["file"]
    title = (check[0]["title"]).title()
    user = check[0]["by"]
    message.from_user.id
    streamtype = check[0]["streamtype"]
    videoid = check[0]["vidid"]
    duration_min = check[0]["dur"]
    status = True if str(streamtype) == "video" else None
    if "live_" in queued:
        n, link = await youtube.video(videoid, True)
        if n == 0:
            return await message.reply_text(_["skip_2"].format(title))
        try:
            await yukki.skip_stream(chat_id, link, video=status)
        except Exception:
            return await message.reply_text(_["call_7"])
        button = telegram_markup(_, chat_id)
        img = await gen_thumb(videoid)
        run = await message.reply_photo(
            photo=img,
            caption=_["stream_1"].format(
                user,
                f"https://t.me/{app.username}?start=info_{videoid}",
            ),
            reply_markup=InlineKeyboardMarkup(button),
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = "tg"
    elif "vid_" in queued:
        flink = (f"https://t.me/{app.username}?start=info_{videoid}",)
        thumbnail = None
        mystic = await message.reply_text(_["call_8"])
        try:
            if youtube.use_fallback:
                file_path, _data, status = await fallback.download(
                    title[:12], video=status
                )
                direct = None
                title = _data.get("title", title)
                thumbnail = _data.get("thumb")
                flink = _data.get("url", flink)
                duration_min = _data.get("duration_min", duration_min)
            else:
                try:
                    file_path, direct = await youtube.download(
                        videoid, mystic, videoid=True, video=status
                    )
                except Exception:
                    youtube.use_fallback = True
                    file_path, _data, status = await fallback.download(
                        title[:12], video=status
                    )
                    title = _data.get("title", title)
                    thumbnail = _data.get("thumb")
                    flink = _data.get("url", flink)
                    duration_min = _data.get("duration_min", duration_min)
        except Exception:
            return await mystic.edit_text(_["call_7"])
        try:
            await yukki.skip_stream(chat_id, file_path, video=status)
        except Exception:
            return await mystic.edit_text(_["call_7"])
        check[0]["dur"] = duration_min
        button = stream_markup(_, videoid, chat_id)
        img = await gen_thumb(videoid, thumbnail)
        run = await message.reply_photo(
            photo=img,
            caption=_["stream_1"].format(
                title[:27],
                flink,
                duration_min,
                user,
            ),
            reply_markup=InlineKeyboardMarkup(button),
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = "stream"
        await mystic.delete()
    elif "index_" in queued:
        try:
            await yukki.skip_stream(chat_id, videoid, video=status)
        except Exception:
            return await message.reply_text(_["call_7"])
        button = telegram_markup(_, chat_id)
        run = await message.reply_photo(
            photo=config.STREAM_IMG_URL,
            caption=_["stream_2"].format(user),
            reply_markup=InlineKeyboardMarkup(button),
        )
        db[chat_id][0]["mystic"] = run
        db[chat_id][0]["markup"] = "tg"
    else:
        try:
            await yukki.skip_stream(chat_id, queued, video=status)
        except Exception:
            return await message.reply_text(_["call_7"])
        if videoid == "telegram":
            button = telegram_markup(_, chat_id)
            run = await message.reply_photo(
                photo=(
                    config.TELEGRAM_AUDIO_URL
                    if str(streamtype) == "audio"
                    else config.TELEGRAM_VIDEO_URL
                ),
                caption=_["stream_1"].format(
                    title, config.SUPPORT_GROUP, check[0]["dur"], user
                ),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
        elif videoid == "soundcloud":
            button = telegram_markup(_, chat_id)
            run = await message.reply_photo(
                photo=(
                    config.SOUNCLOUD_IMG_URL
                    if str(streamtype) == "audio"
                    else config.TELEGRAM_VIDEO_URL
                ),
                caption=_["stream_1"].format(
                    title, config.SUPPORT_GROUP, check[0]["dur"], user
                ),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
        elif "saavn" in videoid:
            button = telegram_markup(_, chat_id)
            url = check[0]["url"]
            details = await saavn.info(url)
            run = await message.reply_photo(
                photo=details["thumb"] or config.TELEGRAM_AUDIO_URL,
                caption=_["stream_1"].format(title, url, check[0]["dur"], user),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "tg"
        else:
            button = stream_markup(_, videoid, chat_id)
            img = await gen_thumb(videoid)
            run = await message.reply_photo(
                photo=img,
                caption=_["stream_1"].format(
                    title[:27],
                    f"https://t.me/{app.username}?start=info_{videoid}",
                    duration_min,
                    user,
                ),
                reply_markup=InlineKeyboardMarkup(button),
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = "stream"


(
    ModuleHelp("Admins")
    .name("en", "Admins")
    .add(
        "en",
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'en')}</b> - Skip the current playing music.\n"
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'en')} [Number]</b> - Skip to a specific track in queue. Example: <code>/skip 3</code> will skip first 2 tracks.",
        priority=17,
    )
    .name("ar", "المسؤولين")
    .add(
        "ar",
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'ar')}</b> - تخطي الموسيقى التي يتم تشغيلها حاليًا.\n"
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'ar')} [رقم]</b> - تخطي إلى رقم معين في قائمة الانتظار. مثال: <code>/skip 3</code> سيتخطى أول مقطعين.",
        priority=17,
    )
    .name("as", "প্ৰশাসক")
    .add(
        "as",
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'as')}</b> - বৰ্তমান বজাই থকা সংগীতটো এৰক।\n"
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'as')} [সংখ্যা]</b> - কিউৰ নিৰ্দিষ্ট স্থানলৈ এৰক। উদাহৰণ: <code>/skip 3</code> এ প্ৰথম 2 টা সংগীত এৰি দিব।",
        priority=17,
    )
    .name("hi", "प्रशासक")
    .add(
        "hi",
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'hi')}</b> - वर्तमान में चल रहे संगीत को छोड़ें।\n"
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'hi')} [संख्या]</b> - कतार में किसी विशेष गाने तक छोड़ें। उदाहरण: <code>/skip 3</code> पहले दो गानों को छोड़ देगा।",
        priority=17,
    )
    .name("ku", "بەڕێوەبەرەکان")
    .add(
        "ku",
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'ku')}</b> - گۆرانییە ناوەڕاستەکە بپارێزە.\n"
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'ku')} [ژمارە]</b> - بپارێزە بۆ گۆرانییە دیاری کراوەکە. نموونە: <code>/skip 3</code> دوایین دوو گۆرانی دەجێتەوە.",
        priority=17,
    )
    .name("tr", "Yöneticiler")
    .add(
        "tr",
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'tr')}</b> - Çalan müziği atla.\n"
        f"<b>✧ {pick_commands('SKIP_COMMAND', 'tr')} [Sayı]</b> - Sıradaki belirli şarkıya atla. Örnek: <code>/skip 3</code> ilk iki şarkıyı atlar.",
        priority=17,
    )
)

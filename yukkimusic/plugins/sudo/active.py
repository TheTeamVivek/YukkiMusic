#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
# pylint: disable=missing-module-docstring, missing-function-docstring

from pyrogram.errors import ChannelInvalid
from pyrogram.types import Message

from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.core.help import ModuleHelp
from yukkimusic.misc import SUDOERS, db
from yukkimusic.utils.database.memorydatabase import (
    get_active_chats,
    get_active_video_chats,
    remove_active_chat,
    remove_active_video_chat,
)
from yukkimusic.utils.decorators.language import LanguageStart


# Function for removing the Active voice and video chat also clear the db dictionary for the chat
async def _clear_(chat_id):
    db[chat_id] = []
    await remove_active_video_chat(chat_id)
    await remove_active_chat(chat_id)


@app.on_message(command("ACTIVEVC_COMMAND") & SUDOERS)
@LanguageStart
async def activevc(_, message: Message, lang):
    mystic = await message.reply_text(lang["active_1"])
    served_chats = await get_active_chats()
    text = ""
    j = 0
    for x in served_chats:
        try:
            title = (await app.get_chat(x)).title
            if (await app.get_chat(x)).username:
                user = (await app.get_chat(x)).username
                text += f"<b>{j + 1}.</b>  [{title}](https://t.me/{user})[`{x}`]\n"
            else:
                text += f"<b>{j + 1}. {title}</b> [`{x}`]\n"
            j += 1
        except ChannelInvalid:
            await _clear_(x)
            continue
    if not text:
        await mystic.edit_text(lang["active_2"])
    else:
        await mystic.edit_text(lang["active_3"].format(text))


@app.on_message(command("ACTIVEVIDEO_COMMAND") & SUDOERS)
@LanguageStart
async def activevi_(_, message: Message, lang):
    mystic = await message.reply_text(lang["active_1"])
    served_chats = await get_active_video_chats()
    text = ""
    j = 0
    for x in served_chats:
        try:
            title = (await app.get_chat(x)).title
            if (await app.get_chat(x)).username:
                user = (await app.get_chat(x)).username
                text += f"<b>{j + 1}.</b>  [{title}](https://t.me/{user})[`{x}`]\n"
            else:
                text += f"<b>{j + 1}. {title}</b> [`{x}`]\n"
            j += 1
        except ChannelInvalid:
            await _clear_(x)
            continue
    if not text:
        await mystic.edit_text(lang["active_2"])
    else:
        await mystic.edit_text(
            lang["active_4"].format(text),
        )


@app.on_message(command("AC_COMMAND") & SUDOERS)
@LanguageStart
async def vc(_, message: Message, lang):
    ac_audio = len(await get_active_chats())
    ac_video = len(await get_active_video_chats())
    if ac_audio != 0:
        ac_audio = ac_audio - ac_video
    await message.reply_text(lang["active_5"].format(ac_audio, ac_video))


# pylint: disable=C0301
(
    ModuleHelp("Active")
    .name("en", "ActiveVc")
    .add(
        "en",
        "<b>Commands to view currently active calls.</b>\n"
        f"<b>✧ {pick_commands('ACTIVEVC_COMMAND', 'en')}</b> - Check active voice chats on the bot.\n"
        f"<b>✧ {pick_commands('ACTIVEVIDEO_COMMAND', 'en')}</b> - Check active video calls on the bot.\n"
        f"<b>✧ {pick_commands('AC_COMMAND', 'en')}</b> - Check all active calls (voice & video) on the bot.",
    )
    .name("ar", "الدردشات النشطة")
    .add(
        "ar",
        "<b>أوامر لعرض المكالمات النشطة.</b>\n\n"
        f"<b>✧ {pick_commands('ACTIVEVC_COMMAND', 'ar')}</b> - التحقق من الدردشات الصوتية النشطة على البوت.\n\n"
        f"<b>✧ {pick_commands('ACTIVEVIDEO_COMMAND', 'ar')}</b> - التحقق من المكالمات المرئية النشطة على البوت.\n\n"
        f"<b>✧ {pick_commands('AC_COMMAND', 'ar')}</b> - التحقق من المكالمات الصوتية والمرئية النشطة على البوت.",
    )
    .name("as", "সক্ৰিয় চেট")
    .add(
        "as",
        "<b>বৰ্তমান সক্ৰিয় কলবোৰ চাবলৈ কমান্ডসমূহ।</b>\n\n"
        f"<b>✧ {pick_commands('ACTIVEVC_COMMAND', 'as')}</b> - বটৰ ওপৰত সক্ৰিয় ভইচ চেটবোৰ পৰীক্ষা কৰক।\n\n"
        f"<b>✧ {pick_commands('ACTIVEVIDEO_COMMAND', 'as')}</b> - বটৰ ওপৰত সক্ৰিয় ভিডিঅ’ কলবোৰ পৰীক্ষা কৰক।\n\n"
        f"<b>✧ {pick_commands('AC_COMMAND', 'as')}</b> - বটৰ ওপৰত সক্ৰিয় ভইচ আৰু ভিডিঅ’ কলবোৰ পৰীক্ষা কৰক।",
    )
    .name("hi", "सक्रिय कॉल")
    .add(
        "hi",
        "<b>बॉट पर सक्रिय कॉल्स देखने के लिए कमांड्स।</b>\n\n"
        f"<b>✧ {pick_commands('ACTIVEVC_COMMAND', 'hi')}</b> - बॉट पर सक्रिय वॉइस चैट्स की जांच करें।\n\n"
        f"<b>✧ {pick_commands('ACTIVEVIDEO_COMMAND', 'hi')}</b> - बॉट पर सक्रिय वीडियो कॉल्स की जांच करें।\n\n"
        f"<b>✧ {pick_commands('AC_COMMAND', 'hi')}</b> - बॉट पर सक्रिय वॉइस और वीडियो कॉल्स की जांच करें।",
    )
    .name("ku", "پەیوەندە چالاکەکان")
    .add(
        "ku",
        "<b>فەرمانەکان بۆ بینینی پەیوەندییە چالاکەکان.</b>\n\n"
        f"<b>✧ {pick_commands('ACTIVEVC_COMMAND', 'ku')}</b> - پشکنینی چاتی دەنگییە چالاکەکان لە بۆتەکە.\n\n"
        f"<b>✧ {pick_commands('ACTIVEVIDEO_COMMAND', 'ku')}</b> - پشکنینی چاتی ڤیدیۆییە چالاکەکان لە بۆتەکە.\n\n"
        f"<b>✧ {pick_commands('AC_COMMAND', 'ku')}</b> - پشکنینی هەموو پەیوەندە چالاکەکان (دەنگ و ڤیدیۆ) لە بۆتەکە.",
    )
    .name("tr", "Aktif Çağrılar")
    .add(
        "tr",
        "<b>Şu anda botta aktif olan aramaları görüntüleme komutları.</b>\n\n"
        f"<b>✧ {pick_commands('ACTIVEVC_COMMAND', 'tr')}</b> - Botta aktif sesli sohbetleri kontrol edin.\n\n"
        f"<b>✧ {pick_commands('ACTIVEVIDEO_COMMAND', 'tr')}</b> - Botta aktif görüntülü aramaları kontrol edin.\n\n"
        f"<b>✧ {pick_commands('AC_COMMAND', 'tr')}</b> - Botta aktif sesli ve görüntülü aramaları kontrol edin.",
    )
)

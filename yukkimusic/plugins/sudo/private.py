#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram.types import Message

import config
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import SUDOERS
from yukkimusic.utils.database import (
    add_private_chat,
    get_private_served_chats,
    is_served_private_chat,
    remove_private_chat,
)
from yukkimusic.utils.decorators.language import language

from . import mhelp


@app.on_message(command("AUTHORIZE_COMMAND") & SUDOERS)
@language
async def authorize(client, message: Message, _):
    if not config.PRIVATE_BOT_MODE:
        return await message.reply_text(_["pbot_12"])
    if len(message.command) != 2:
        return await message.reply_text(_["pbot_1"])
    try:
        chat_id = int(message.text.strip().split()[1])
    except Exception:
        return await message.reply_text(_["pbot_7"])
    if not await is_served_private_chat(chat_id):
        await add_private_chat(chat_id)
        await message.reply_text(_["pbot_3"])
    else:
        await message.reply_text(_["pbot_5"])


@app.on_message(command("UNAUTHORIZE_COMMAND") & SUDOERS)
@language
async def unauthorize(client, message: Message, _):
    if not config.PRIVATE_BOT_MODE:
        return await message.reply_text(_["pbot_12"])
    if len(message.command) != 2:
        return await message.reply_text(_["pbot_2"])
    try:
        chat_id = int(message.text.strip().split()[1])
    except Exception:
        return await message.reply_text(_["pbot_7"])
    if not await is_served_private_chat(chat_id):
        return await message.reply_text(_["pbot_6"])
    else:
        await remove_private_chat(chat_id)
        return await message.reply_text(_["pbot_4"])


@app.on_message(command("AUTHORIZED_COMMAND") & SUDOERS)
@language
async def authorized(client, message: Message, _):
    if not config.PRIVATE_BOT_MODE:
        return await message.reply_text(_["pbot_12"])
    m = await message.reply_text(_["pbot_8"])
    served_chats = []
    text = _["pbot_9"]
    chats = await get_private_served_chats()
    for chat in chats:
        served_chats.append(int(chat["chat_id"]))
    count = 0
    co = 0
    msg = _["pbot_13"]
    for served_chat in served_chats:
        try:
            title = (await app.get_chat(served_chat)).title
            count += 1
            text += f"{count}:- {title[:15]} [{served_chat}]\n"
        except Exception:
            title = _["pbot_10"]
            co += 1
            msg += f"{co}:- {title} [{served_chat}]\n"
    if co == 0:
        if count == 0:
            return await m.edit(_["pbot_11"])
        else:
            return await m.edit(text)
    else:
        if count == 0:
            await m.edit(msg)
        else:
            text = f"{text} {msg}"
            return await m.edit(text)


(
    mhelp.add(
        "en",
        (
            "<u><b>⚡️Private Bot:</b></u>\n\n"
            "When your Private Bot Mode is enabled:\n\n"
            f"<b>✧ {pick_commands('AUTHORIZE_COMMAND')}</b> [CHAT_ID] - Allow a chat to use your bot.\n"
            f"<b>✧ {pick_commands('UNAUTHORIZE_COMMAND')}</b> [CHAT_ID] - Disallow a chat from using your bot.\n"
            f"<b>✧ {pick_commands('AUTHORIZED_COMMAND')}</b> - Check all allowed chats of your bot."
        ),
    )
    .add(
        "ar",
        (
            "<u><b>⚡️بوت خاص:</b></u>\n\n"
            "عند تفعيل وضع البوت الخاص:\n\n"
            f"<b>✧ {pick_commands('AUTHORIZE_COMMAND')}</b> [CHAT_ID] - السماح لدردشة باستخدام البوت الخاص بك.\n"
            f"<b>✧ {pick_commands('UNAUTHORIZE_COMMAND')}</b> [CHAT_ID] - منع دردشة من استخدام البوت الخاص بك.\n"
            f"<b>✧ {pick_commands('AUTHORIZED_COMMAND')}</b> - تحقق من جميع الدردشات المسموح بها في البوت الخاص بك."
        ),
    )
    .add(
        "as",
        (
            "<u><b>⚡️প্ৰাইভেট বট:</b></u>\n\n"
            "যেতিয়া আপোনাৰ প্ৰাইভেট বট ম'ড এনেবল কৰা হয়:\n\n"
            f"<b>✧ {pick_commands('AUTHORIZE_COMMAND')}</b> [CHAT_ID] - এখন চেটক আপোনাৰ বট ব্যৱহাৰ কৰাৰ অনুমতি দিয়ক।\n"
            f"<b>✧ {pick_commands('UNAUTHORIZE_COMMAND')}</b> [CHAT_ID] - এখন চেটক আপোনাৰ বট ব্যৱহাৰ নকৰিবলৈ বাধা দিয়ক।\n"
            f"<b>✧ {pick_commands('AUTHORIZED_COMMAND')}</b> - আপোনাৰ বটৰ সকলো অনুমোদিত চেট চাওক।"
        ),
    )
    .add(
        "hi",
        (
            "<u><b>⚡️प्राइवेट बॉट:</b></u>\n\n"
            "जब आपका प्राइवेट बॉट मोड सक्षम होता है:\n\n"
            f"<b>✧ {pick_commands('AUTHORIZE_COMMAND')}</b> [CHAT_ID] - किसी चैट को आपके बॉट का उपयोग करने की अनुमति दें।\n"
            f"<b>✧ {pick_commands('UNAUTHORIZE_COMMAND')}</b> [CHAT_ID] - किसी चैट को आपके बॉट का उपयोग करने से रोकें।\n"
            f"<b>✧ {pick_commands('AUTHORIZED_COMMAND')}</b> - अपने बॉट की सभी अनुमत चैट देखें।"
        ),
    )
    .add(
        "ku",
        (
            "<u><b>⚡️بۆتی تایبەتی:</b></u>\n\n"
            "کاتێک دۆخی بۆتی تایبەتی چالاک بێت:\n\n"
            f"<b>✧ {pick_commands('AUTHORIZE_COMMAND')}</b> [CHAT_ID] - ڕێگە بدە بۆت چاتێک بەکاردەهێنرێت.\n"
            f"<b>✧ {pick_commands('UNAUTHORIZE_COMMAND')}</b> [CHAT_ID] - چاتێک لە بەکارهێنانی بۆت بگرە.\n"
            f"<b>✧ {pick_commands('AUTHORIZED_COMMAND')}</b> - هەموو چاتە ڕێگەپێدراوەکانت لە بۆت بپشکنە."
        ),
    )
    .add(
        "tr",
        (
            "<u><b>⚡️Özel Bot:</b></u>\n\n"
            "Özel Bot Modu etkinleştirildiğinde:\n\n"
            f"<b>✧ {pick_commands('AUTHORIZE_COMMAND')}</b> [CHAT_ID] - Bir sohbetin botunuzu kullanmasına izin verin.\n"
            f"<b>✧ {pick_commands('UNAUTHORIZE_COMMAND')}</b> [CHAT_ID] - Bir sohbetin botunuzu kullanmasını engelleyin.\n"
            f"<b>✧ {pick_commands('AUTHORIZED_COMMAND')}</b> - Botunuzun izin verilen tüm sohbetlerini kontrol edin."
        ),
    )
)

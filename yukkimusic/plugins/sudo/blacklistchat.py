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

from config import BANNED_USERS
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import SUDOERS
from yukkimusic.utils.database import (
    blacklist_chat,
    blacklisted_chats,
    whitelist_chat,
)
from yukkimusic.utils.decorators.language import language

from . import mhelp


@app.on_message(command("BLACKLISTCHAT_COMMAND") & SUDOERS)
@language
async def blacklist_chat_func(client, message: Message, _):
    if len(message.command) != 2:
        return await message.reply_text(_["black_1"])
    chat_id = int(message.text.strip().split()[1])
    if chat_id in await blacklisted_chats():
        return await message.reply_text(_["black_2"])
    blacklisted = await blacklist_chat(chat_id)
    if blacklisted:
        await message.reply_text(_["black_3"])
    else:
        await message.reply_text("sᴏᴍᴇᴛʜɪɴɢ ᴡʀᴏɴɢ ʜᴀᴘᴘᴇɴᴇᴅ.")
    try:
        await app.leave_chat(chat_id)
    except Exception:
        pass


@app.on_message(command("WHITELISTCHAT_COMMAND") & SUDOERS)
@language
async def white_function(client, message: Message, _):
    if len(message.command) != 2:
        return await message.reply_text(_["black_4"])
    chat_id = int(message.text.strip().split()[1])
    if chat_id not in await blacklisted_chats():
        return await message.reply_text(_["black_5"])
    whitelisted = await whitelist_chat(chat_id)
    if whitelisted:
        return await message.reply_text(_["black_6"])
    await message.reply_text("Something wrong happened")


@app.on_message(command("BLACKLISTEDCHAT_COMMAND") & ~BANNED_USERS)
@language
async def all_chats(client, message: Message, _):
    text = _["black_7"]
    j = 0
    for count, chat_id in enumerate(await blacklisted_chats(), 1):
        try:
            title = (await app.get_chat(chat_id)).title
        except Exception:
            title = "Private"
        j = 1
        text += f"**{count}. {title}** [`{chat_id}`]\n"
    if j == 0:
        await message.reply_text(_["black_8"])
    else:
        await message.reply_text(text)


(
    mhelp.add(
        "en",
        (
            f"<b>✧ {pick_commands('BLACKLISTCHAT_COMMAND')}</b> [chat ID] - Blacklist any chat from using the Music Bot.\n"
            f"<b>✧ {pick_commands('WHITELISTCHAT_COMMAND')}</b> [chat ID] - Whitelist any blacklisted chat from using the Music Bot.\n"
            f"<b>✧ {pick_commands('BLACKLISTEDCHAT_COMMAND')}</b> - Check all blocked chats."
        ),
    )
    .add(
        "ar",
        (
            f"<b>✧ {pick_commands('BLACKLISTCHAT_COMMAND')}</b> [معرف الدردشة] - حظر أي دردشة من استخدام البوت.\n"
            f"<b>✧ {pick_commands('WHITELISTCHAT_COMMAND')}</b> [معرف الدردشة] - إلغاء حظر أي دردشة محظورة من استخدام البوت.\n"
            f"<b>✧ {pick_commands('BLACKLISTEDCHAT_COMMAND')}</b> - عرض جميع الدردشات المحظورة."
        ),
    )
    .add(
        "as",
        (
            f"<b>✧ {pick_commands('BLACKLISTCHAT_COMMAND')}</b> [চেট আইডি] - সংগীত বট ব্যৱহাৰ কৰিবলৈ যিকোনো চেট ব্লেকলিষ্ট কৰক।\n"
            f"<b>✧ {pick_commands('WHITELISTCHAT_COMMAND')}</b> [চেট আইডি] - সংগীত বট ব্যৱহাৰ কৰিবলৈ ব্লেকলিষ্ট কৰা চেট হোৱাইটলিষ্ট কৰক।\n"
            f"<b>✧ {pick_commands('BLACKLISTEDCHAT_COMMAND')}</b> - সকলো ব্লক কৰা চেট চাওক।"
        ),
    )
    .add(
        "hi",
        (
            f"<b>✧ {pick_commands('BLACKLISTCHAT_COMMAND')}</b> [चैट आईडी] - किसी भी चैट को म्यूजिक बॉट के उपयोग से ब्लैकलिस्ट करें।\n"
            f"<b>✧ {pick_commands('WHITELISTCHAT_COMMAND')}</b> [चैट आईडी] - किसी भी ब्लैकलिस्ट की गई चैट को ब्लैकलिस्ट से हटाएँ।\n"
            f"<b>✧ {pick_commands('BLACKLISTEDCHAT_COMMAND')}</b> - सभी ब्लॉक की गई चैट्स देखें।"
        ),
    )
    .add(
        "ku",
        (
            f"<b>✧ {pick_commands('BLACKLISTCHAT_COMMAND')}</b> [ناسنامەی چات] - هەر چاتێک لە بەکارهێنانی بۆتی میوزیک قەدەغە بکە.\n"
            f"<b>✧ {pick_commands('WHITELISTCHAT_COMMAND')}</b> [ناسنامەی چات] - هەر چاتێکی قەدەغەکراو لە قەدەغەکردن دەربهێنە.\n"
            f"<b>✧ {pick_commands('BLACKLISTEDCHAT_COMMAND')}</b> - هەموو چاتە قەدەغەکراوەکان بپشکنە."
        ),
    )
    .add(
        "tr",
        (
            f"<b>✧ {pick_commands('BLACKLISTCHAT_COMMAND')}</b> [sohbet ID] - Müzik botunu kullanmasını engellemek için herhangi bir sohbeti kara listeye al.\n"
            f"<b>✧ {pick_commands('WHITELISTCHAT_COMMAND')}</b> [sohbet ID] - Kara listeye alınmış herhangi bir sohbeti kara listeden kaldır.\n"
            f"<b>✧ {pick_commands('BLACKLISTEDCHAT_COMMAND')}</b> - Tüm engellenen sohbetleri kontrol et."
        ),
    )
)

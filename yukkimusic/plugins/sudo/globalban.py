#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import asyncio

from pyrogram import filters
from pyrogram.errors import FloodWait
from pyrogram.types import Message

from config import BANNED_USERS, OWNER_ID
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import SUDOERS
from yukkimusic.utils import get_readable_time
from yukkimusic.utils.database import (
    add_banned_user,
    get_banned_count,
    get_banned_users,
    get_served_chats,
    is_banned_user,
    remove_banned_user,
)
from yukkimusic.utils.decorators.language import language

from . import ohelp


@app.on_message(command("GBAN_COMMAND") & filters.user(OWNER_ID))
@language
async def gbanuser(client, message: Message, _):
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        user = await app.get_users(user)
        user_id = user.id
        mention = user.mention
    else:
        user_id = message.reply_to_message.from_user.id
        mention = message.reply_to_message.from_user.mention
    if user_id == message.from_user.id:
        return await message.reply_text(_["gban_1"])
    elif user_id == app.id:
        return await message.reply_text(_["gban_2"])
    elif user_id in SUDOERS:
        return await message.reply_text(_["gban_3"])
    is_gbanned = await is_banned_user(user_id)
    if is_gbanned:
        return await message.reply_text(_["gban_4"].format(mention))
    if user_id not in BANNED_USERS:
        BANNED_USERS.add(user_id)
    served_chats = []
    chats = await get_served_chats()
    for chat in chats:
        served_chats.append(int(chat["chat_id"]))
    time_expected = len(served_chats)
    time_expected = get_readable_time(time_expected)
    mystic = await message.reply_text(_["gban_5"].format(mention, time_expected))
    number_of_chats = 0
    for chat_id in served_chats:
        try:
            await app.ban_chat_member(chat_id, user_id)
            number_of_chats += 1
        except FloodWait as e:
            await asyncio.sleep(int(e.value))
        except Exception:
            pass
    await add_banned_user(user_id)
    await message.reply_text(_["gban_6"].format(mention, number_of_chats))
    await mystic.delete()


@app.on_message(command("UNGBAN_COMMAND") & filters.user(OWNER_ID))
@language
async def gungabn(client, message: Message, _):
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        user = await app.get_users(user)
        user_id = user.id
        mention = user.mention
    else:
        user_id = message.reply_to_message.from_user.id
        mention = message.reply_to_message.from_user.mention
    is_gbanned = await is_banned_user(user_id)
    if not is_gbanned:
        return await message.reply_text(_["gban_7"].format(mention))
    if user_id in BANNED_USERS:
        BANNED_USERS.remove(user_id)
    served_chats = []
    chats = await get_served_chats()
    for chat in chats:
        served_chats.append(int(chat["chat_id"]))
    time_expected = len(served_chats)
    time_expected = get_readable_time(time_expected)
    mystic = await message.reply_text(_["gban_8"].format(mention, time_expected))
    number_of_chats = 0
    for chat_id in served_chats:
        try:
            await app.unban_chat_member(chat_id, user_id)
            number_of_chats += 1
        except FloodWait as e:
            await asyncio.sleep(int(e.value))
        except Exception:
            pass
    await remove_banned_user(user_id)
    await message.reply_text(_["gban_9"].format(mention, number_of_chats))
    await mystic.delete()


@app.on_message(command("GBANNED_COMMAND") & filters.user(OWNER_ID))
@language
async def gbanned_list(client, message: Message, _):
    counts = await get_banned_count()
    if counts == 0:
        return await message.reply_text(_["gban_10"])
    mystic = await message.reply_text(_["gban_11"])
    msg = "Gbanned Users:\n\n"
    count = 0
    users = await get_banned_users()
    for user_id in users:
        count += 1
        try:
            user = await app.get_users(user_id)
            user = user.first_name if not user.mention else user.mention
            msg += f"{count}➤ {user}\n"
        except Exception:
            msg += f"{count}➤ [Unfetched User]{user_id}\n"
            continue
    if count == 0:
        return await mystic.edit_text(_["gban_10"])
    else:
        return await mystic.edit_text(msg)


(
    ohelp.add(
        "en",
        (
            f"<b>✧ {pick_commands('GBAN_COMMAND')}</b> [Username or reply to a user] - Gban a user from all served chats and stop them from using your bot.\n"
            f"<b>✧ {pick_commands('UNGBAN_COMMAND')}</b> [Username or reply to a user] - Remove a user from the bot's gban list and allow them to use your bot.\n"
            f"<b>✧ {pick_commands('GBANNED_COMMAND')}</b> - Check the list of gban users."
        ),
    )
    .add(
        "ar",
        (
            f"<b>✧ {pick_commands('GBAN_COMMAND')}</b> [اسم المستخدم أو الرد على المستخدم] - حظر مستخدم من جميع الدردشات ومنعه من استخدام البوت.\n"
            f"<b>✧ {pick_commands('UNGBAN_COMMAND')}</b> [اسم المستخدم أو الرد على المستخدم] - إزالة مستخدم من قائمة الحظر العام والسماح له باستخدام البوت.\n"
            f"<b>✧ {pick_commands('GBANNED_COMMAND')}</b> - عرض قائمة المستخدمين المحظورين."
        ),
    )
    .add(
        "as",
        (
            f"<b>✧ {pick_commands('GBAN_COMMAND')}</b> [ব্যৱহাৰকাৰী নাম বা ব্যৱহাৰকাৰীক প্ৰত্যুত্তৰ কৰক] - সকলো চেটৰ পৰা এজন ব্যৱহাৰকাৰীক গ্লোবেলি ব্যান কৰক আৰু তেওঁক বট ব্যৱহাৰ কৰাৰ পৰা আটকাওক।\n"
            f"<b>✧ {pick_commands('UNGBAN_COMMAND')}</b> [ব্যৱহাৰকাৰী নাম বা ব্যৱহাৰকাৰীক প্ৰত্যুত্তৰ কৰক] - এজন ব্যৱহাৰকাৰীক গ্লোবেলি ব্যান তালিকাৰ পৰা আতৰাওক আৰু তেওঁক বট ব্যৱহাৰ কৰিবলৈ দিয়ক।\n"
            f"<b>✧ {pick_commands('GBANNED_COMMAND')}</b> - গ্লোবেলি ব্যান কৰা ব্যৱহাৰকাৰীৰ তালিকা চাওক।"
        ),
    )
    .add(
        "hi",
        (
            f"<b>✧ {pick_commands('GBAN_COMMAND')}</b> [यूज़रनेम या यूज़र को रिप्लाई करें] - किसी यूज़र को सभी चैट से ग्लोबल बैन करें और उसे बॉट इस्तेमाल करने से रोकें।\n"
            f"<b>✧ {pick_commands('UNGBAN_COMMAND')}</b> [यूज़रनेम या यूज़र को रिप्लाई करें] - किसी यूज़र को ग्लोबल बैन सूची से हटाएँ और उसे बॉट इस्तेमाल करने दें।\n"
            f"<b>✧ {pick_commands('GBANNED_COMMAND')}</b> - ग्लोबल बैन किए गए यूज़र्स की सूची देखें।"
        ),
    )
    .add(
        "ku",
        (
            f"<b>✧ {pick_commands('GBAN_COMMAND')}</b> [ناوی بەکارهێنەر یان وەڵامدانەوە بە بەکارهێنەر] - بەکارهێنەرێک لە هەموو چاتەکان بەنێ و مانهێڵە لە بەکارهێنانی بۆت.\n"
            f"<b>✧ {pick_commands('UNGBAN_COMMAND')}</b> [ناوی بەکارهێنەر یان وەڵامدانەوە بە بەکارهێنەر] - بەکارهێنەرێک لە لیستی بەنە گشتی دەربهێنە و ڕێگە بدە لە بەکارهێنانی بۆت.\n"
            f"<b>✧ {pick_commands('GBANNED_COMMAND')}</b> - لیستی بەکارهێنەرانی بەنە گشتی بپشکنە."
        ),
    )
    .add(
        "tr",
        (
            f"<b>✧ {pick_commands('GBAN_COMMAND')}</b> [Kullanıcı adı veya kullanıcıya yanıt ver] - Bir kullanıcıyı tüm sohbetlerden global olarak yasakla ve botu kullanmasını engelle.\n"
            f"<b>✧ {pick_commands('UNGBAN_COMMAND')}</b> [Kullanıcı adı veya kullanıcıya yanıt ver] - Bir kullanıcıyı global yasak listesinden kaldır ve botu kullanmasına izin ver.\n"
            f"<b>✧ {pick_commands('GBANNED_COMMAND')}</b> - Global olarak yasaklanan kullanıcıların listesini kontrol et."
        ),
    )
)

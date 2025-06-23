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
from pyrogram.types import Message

from config import BANNED_USERS, adminlist
from strings import command, pick_commands
from YukkiMusic import app
from YukkiMusic.utils.database import (
    delete_authuser,
    get_authuser,
    get_authuser_names,
    save_authuser,
)
from YukkiMusic.utils.decorators import AdminActual, language
from YukkiMusic.utils.formatters import int_to_alpha

from . import mhelp


@app.on_message(command("AUTH_COMMAND") & filters.group & ~BANNED_USERS)
@AdminActual
async def auth(client, message: Message, _):
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await app.get_users(user)
        user_id = message.from_user.id
        token = await int_to_alpha(user.id)
        from_user_name = message.from_user.first_name
        from_user_id = message.from_user.id
        _check = await get_authuser_names(message.chat.id)
        count = len(_check)
        if int(count) == 20:
            return await message.reply_text(_["auth_1"])
        if token not in _check:
            assis = {
                "auth_user_id": user.id,
                "auth_name": user.first_name,
                "admin_id": from_user_id,
                "admin_name": from_user_name,
            }
            get = adminlist.get(message.chat.id)
            if get:
                if user.id not in get:
                    get.append(user.id)
            await save_authuser(message.chat.id, token, assis)
            return await message.reply_text(_["auth_2"])
        await message.reply_text(_["auth_3"])
        return
    from_user_id = message.from_user.id
    user_id = message.reply_to_message.from_user.id
    user_name = message.reply_to_message.from_user.first_name
    token = await int_to_alpha(user_id)
    from_user_name = message.from_user.first_name
    _check = await get_authuser_names(message.chat.id)
    count = 0
    for smex in _check:
        count += 1
    if int(count) == 20:
        return await message.reply_text(_["auth_1"])
    if token not in _check:
        assis = {
            "auth_user_id": user_id,
            "auth_name": user_name,
            "admin_id": from_user_id,
            "admin_name": from_user_name,
        }
        get = adminlist.get(message.chat.id)
        if get:
            if user_id not in get:
                get.append(user_id)
        await save_authuser(message.chat.id, token, assis)
        return await message.reply_text(_["auth_2"])
    await message.reply_text(_["auth_3"])


@app.on_message(command("UNAUTH_COMMAND") & filters.group & ~BANNED_USERS)
@AdminActual
async def unauthusers(client, message: Message, _):
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await app.get_users(user)
        token = await int_to_alpha(user.id)
        deleted = await delete_authuser(message.chat.id, token)
        get = adminlist.get(message.chat.id)
        if get:
            if user.id in get:
                get.remove(user.id)
        if deleted:
            return await message.reply_text(_["auth_4"])
        return await message.reply_text(_["auth_5"])
    user_id = message.reply_to_message.from_user.id
    token = await int_to_alpha(user_id)
    deleted = await delete_authuser(message.chat.id, token)
    get = adminlist.get(message.chat.id)
    if get:
        if user_id in get:
            get.remove(user_id)
    if deleted:
        return await message.reply_text(_["auth_4"])
    return await message.reply_text(_["auth_5"])


@app.on_message(command("AUTHUSERS_COMMAND") & filters.group & ~BANNED_USERS)
@language
async def authusers(client, message: Message, _):
    _playlist = await get_authuser_names(message.chat.id)
    if not _playlist:
        return await message.reply_text(_["setting_5"])
    else:
        j = 0
        mystic = await message.reply_text(_["auth_6"])
        text = _["auth_7"]
        for note in _playlist:
            _note = await get_authuser(message.chat.id, note)
            user_id = _note["auth_user_id"]
            admin_id = _note["admin_id"]
            admin_name = _note["admin_name"]
            try:
                user = await app.get_users(user_id)
                user = user.first_name
                j += 1
            except Exception:
                continue
            text += f"{j}➤ {user}[`{user_id}`]\n"
            text += f"   {_['auth_8']} {admin_name}[`{admin_id}`]\n\n"
        await mystic.delete()
        await message.reply_text(text)


(
    mhelp.add(
        "en",
        f"<b>Auth Users can use Admin commands without admin rights in your chat.</b>\n\n"
        f"<b>✧ {pick_commands('AUTH_COMMAND', 'en')}</b> [Username] - Add a user to AUTH LIST of the group.\n"
        f"<b>✧ {pick_commands('UNAUTH_COMMAND', 'en')}</b> [Username] - Remove a user from AUTH LIST of the group.\n"
        f"<b>✧ {pick_commands('AUTHUSERS_COMMAND', 'en')}</b> - Check AUTH LIST of the group.",
        priority=13,
    )
    .add(
        "ar",
        f"<b>يمكن للمستخدمين المفوضين استخدام أوامر المسؤول دون صلاحيات المسؤول في الدردشة.</b>\n\n"
        f"<b>✧ {pick_commands('AUTH_COMMAND', 'ar')}</b> [اسم المستخدم] - أضف مستخدمًا إلى قائمة AUTH في المجموعة.\n"
        f"<b>✧ {pick_commands('UNAUTH_COMMAND', 'ar')}</b> [اسم المستخدم] - إزالة مستخدم من قائمة AUTH في المجموعة.\n"
        f"<b>✧ {pick_commands('AUTHUSERS_COMMAND', 'ar')}</b> - تحقق من قائمة AUTH في المجموعة.",
        priority=13,
    )
    .add(
        "as",
        f"<b>Auth ব্যৱহাৰকাৰীসকলে আপোনাৰ চেটত প্ৰশাসকসকলৰ অধিকার নোহোৱাকৈ প্ৰশাসক নিৰ্দেশনা ব্যৱহাৰ কৰিব পাৰে।</b>\n\n"
        f"<b>✧ {pick_commands('AUTH_COMMAND', 'as')}</b> [ব্যৱহাৰকাৰী নাম] - গোটৰ AUTH তালিকাত ব্যৱহাৰকাৰীক যোগ কৰক।\n"
        f"<b>✧ {pick_commands('UNAUTH_COMMAND', 'as')}</b> [ব্যৱহাৰকাৰী নাম] - গোটৰ AUTH তালিকাৰ পৰা ব্যৱহাৰকাৰীক আঁতৰাওক।\n"
        f"<b>✧ {pick_commands('AUTHUSERS_COMMAND', 'as')}</b> - গোটৰ AUTH তালিকা পৰীক্ষা কৰক।",
        priority=13,
    )
    .add(
        "hi",
        f"<b>Auth उपयोगकर्ता बिना एडमिन राइट्स के एडमिन कमांड्स का उपयोग कर सकते हैं।</b>\n\n"
        f"<b>✧ {pick_commands('AUTH_COMMAND', 'hi')}</b> [Username] - उपयोगकर्ता को AUTH सूची में जोड़ें।\n"
        f"<b>✧ {pick_commands('UNAUTH_COMMAND', 'hi')}</b> [Username] - उपयोगकर्ता को AUTH सूची से हटाएं।\n"
        f"<b>✧ {pick_commands('AUTHUSERS_COMMAND', 'hi')}</b> - AUTH सूची देखें।",
        priority=13,
    )
    .add(
        "ku",
        f"<b>بەکارهێنەرەکانى AUTH دەتوانن فەرمانە ئەدمینەکان بەبێ مافەکانی ئەدمین بەکارببەن.</b>\n\n"
        f"<b>✧ {pick_commands('AUTH_COMMAND', 'ku')}</b> [ناوی بەکارهێنەر] - زیادکردنی بەکارهێنەر بۆ لیستی AUTH.\n"
        f"<b>✧ {pick_commands('UNAUTH_COMMAND', 'ku')}</b> [ناوی بەکارهێنەر] - سڕینەوەی بەکارهێنەر لە لیستی AUTH.\n"
        f"<b>✧ {pick_commands('AUTHUSERS_COMMAND', 'ku')}</b> - بینینی لیستی AUTH.",
        priority=13,
    )
    .add(
        "tr",
        f"<b>Yetkilendirilmiş kullanıcılar, sohbette yönetici olmadan yönetici komutlarını kullanabilir.</b>\n\n"
        f"<b>✧ {pick_commands('AUTH_COMMAND', 'tr')}</b> [Kullanıcı Adı] - Grubun AUTH listesine kullanıcı ekleyin.\n"
        f"<b>✧ {pick_commands('UNAUTH_COMMAND', 'tr')}</b> [Kullanıcı Adı] - Grubun AUTH listesinden kullanıcıyı kaldırın.\n"
        f"<b>✧ {pick_commands('AUTHUSERS_COMMAND', 'tr')}</b> - Grubun AUTH listesini kontrol edin.",
        priority=13,
    )
)

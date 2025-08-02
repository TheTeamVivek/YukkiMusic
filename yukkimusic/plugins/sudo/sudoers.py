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

from config import BANNED_USERS, MONGO_DB_URI, OWNER_ID
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import SUDOERS
from yukkimusic.utils.database import add_sudo, remove_sudo
from yukkimusic.utils.decorators.language import language

from . import ohelp


@app.on_message(command("ADDSUDO_COMMAND") & filters.user(OWNER_ID))
@language
async def useradd(client, message: Message, _):
    if MONGO_DB_URI is None:
        return await message.reply_text(
            "**Due to privacy issues, You can't manage sudoers when you are on yukki Database.\n\n Please fill Your MONGO_DB_URI in your vars to use this features**"
        )
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await app.get_users(user)
        if user.id in SUDOERS:
            return await message.reply_text(_["sudo_1"].format(user.mention))
        added = await add_sudo(user.id)
        if added:
            SUDOERS.add(user.id)
            await message.reply_text(_["sudo_2"].format(user.mention))
        else:
            await message.reply_text("Something wrong happened")
        return
    if message.reply_to_message.from_user.id in SUDOERS:
        return await message.reply_text(
            _["sudo_1"].format(message.reply_to_message.from_user.mention)
        )
    added = await add_sudo(message.reply_to_message.from_user.id)
    if added:
        SUDOERS.add(message.reply_to_message.from_user.id)
        await message.reply_text(
            _["sudo_2"].format(message.reply_to_message.from_user.mention)
        )
    else:
        await message.reply_text("Something wrong happened")
    return


@app.on_message(command("DELSUDO_COMMAND") & filters.user(OWNER_ID))
@language
async def userdel(client, message: Message, _):
    if MONGO_DB_URI is None:
        return await message.reply_text(
            "**Due to privacy issues, You can't manage sudoers when you are on yukki Database.\n\n Please fill Your MONGO_DB_URI in your vars to use this features**"
        )
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await app.get_users(user)
        if user.id not in SUDOERS:
            return await message.reply_text(_["sudo_3"])
        removed = await remove_sudo(user.id)
        if removed:
            SUDOERS.remove(user.id)
            await message.reply_text(_["sudo_4"])
            return
        await message.reply_text("Something wrong happened")
        return
    user_id = message.reply_to_message.from_user.id
    if user_id not in SUDOERS:
        return await message.reply_text(_["sudo_3"])
    removed = await remove_sudo(user_id)
    if removed:
        SUDOERS.remove(user_id)
        await message.reply_text(_["sudo_4"])
        return
    await message.reply_text("Something wrong happened")


@app.on_message(command("SUDOUSERS_COMMAND") & ~BANNED_USERS)
@language
async def sudoers_list(client, message: Message, _):
    text = _["sudo_5"]
    count = 0
    for x in OWNER_ID:
        try:
            user = await app.get_users(x)
            user = user.first_name if not user.mention else user.mention
            count += 1
        except Exception:
            continue
        text += f"{count}➤ {user} (`{x}`)\n"
    smex = 0
    for user_id in SUDOERS:
        if user_id not in OWNER_ID:
            try:
                user = await app.get_users(user_id)
                user = user.first_name if not user.mention else user.mention
                if smex == 0:
                    smex += 1
                    text += _["sudo_6"]
                count += 1
                text += f"{count}➤ {user} (`{user_id}`)\n"
            except Exception:
                continue
    if not text:
        await message.reply_text(_["sudo_7"])
    else:
        await message.reply_text(text)


(
    ohelp.add(
        "en",
        (
            "<b><u>Add and remove sudoers:</u></b>\n\n"
            f"<b>{pick_commands('ADDSUDO_COMMAND')} [Username or reply to a user]</b> - Add sudo in your bot\n"
            f"<b>{pick_commands('DELSUDO_COMMAND')} [Username or user ID or reply to a user]</b> - Remove from bot sudoers\n"
            f"<b>{pick_commands('SUDOUSERS_COMMAND')}</b> - Get a list of all sudoers"
        ),
        priority=100,
    )
    .add(
        "ar",
        (
            "<b><u>إضافة وإزالة المدراء:</u></b>\n\n"
            f"<b>{pick_commands('ADDSUDO_COMMAND')} [اسم المستخدم أو الرد على مستخدم]</b> - أضف مديرًا إلى البوت\n"
            f"<b>{pick_commands('DELSUDO_COMMAND')} [اسم المستخدم أو معرف المستخدم أو الرد على مستخدم]</b> - إزالة من مدراء البوت\n"
            f"<b>{pick_commands('SUDOUSERS_COMMAND')}</b> - احصل على قائمة بجميع المدراء"
        ),
        priority=100,
    )
    .add(
        "as",
        (
            "<b><u>সুডোৰ্ছ যোগ আৰু আঁতৰাওক:</u></b>\n\n"
            f"<b>{pick_commands('ADDSUDO_COMMAND')} [ব্যৱহাৰকাৰী নাম বা এজন ব্যৱহাৰকাৰীক উত্তৰ দিয়ক]</b> - আপোনাৰ বটত সুডো যোগ কৰক\n"
            f"<b>{pick_commands('DELSUDO_COMMAND')} [ব্যৱহাৰকাৰী নাম বা ব্যৱহাৰকাৰী আইডি বা এজন ব্যৱহাৰকাৰীক উত্তৰ দিয়ক]</b> - বটৰ সুডোৰ্ছৰ পৰা আঁতৰাওক\n"
            f"<b>{pick_commands('SUDOUSERS_COMMAND')}</b> - সকলো সুডোৰ্ছৰ তালিকা লাভ কৰক"
        ),
        priority=100,
    )
    .add(
        "hi",
        (
            "<b><u>सूडोअर्स जोड़ें और हटाएँ:</u></b>\n\n"
            f"<b>{pick_commands('ADDSUDO_COMMAND')} [उपयोगकर्ता नाम या किसी उपयोगकर्ता को उत्तर दें]</b> - अपने बॉट में सूडो जोड़ें\n"
            f"<b>{pick_commands('DELSUDO_COMMAND')} [उपयोगकर्ता नाम या उपयोगकर्ता आईडी या किसी उपयोगकर्ता को उत्तर दें]</b> - बॉट के सूडोअर्स से हटाएँ\n"
            f"<b>{pick_commands('SUDOUSERS_COMMAND')}</b> - सभी सूडोअर्स की सूची प्राप्त करें"
        ),
        priority=100,
    )
    .add(
        "ku",
        (
            "<b><u>زیادکردن و سڕینەوەی سودوەکان:</u></b>\n\n"
            f"<b>{pick_commands('ADDSUDO_COMMAND')} [ناوی بەکارهێنەر یان وەڵامدانەوە بە بەکارهێنەرێک]</b> - سودو زیاد بکە بۆ بۆت\n"
            f"<b>{pick_commands('DELSUDO_COMMAND')} [ناوی بەکارهێنەر یان ناسنامەی بەکارهێنەر یان وەڵامدانەوە]</b> - سڕینەوە لە سودوەکانی بۆت\n"
            f"<b>{pick_commands('SUDOUSERS_COMMAND')}</b> - وەشانێکی لیستی هەموو سودوەکان بگرە"
        ),
        priority=100,
    )
    .add(
        "tr",
        (
            "<b><u>Sudo kullanıcılarını ekle ve kaldır:</u></b>\n\n"
            f"<b>{pick_commands('ADDSUDO_COMMAND')} [Kullanıcı adı veya bir kullanıcıya yanıt ver]</b> - Bota sudo ekle\n"
            f"<b>{pick_commands('DELSUDO_COMMAND')} [Kullanıcı adı veya kullanıcı ID’si veya bir kullanıcıya yanıt ver]</b> - Bot sudo kullanıcılarından kaldır\n"
            f"<b>{pick_commands('SUDOUSERS_COMMAND')}</b> - Tüm sudo kullanıcılarının listesini al"
        ),
        priority=100,
    )
)

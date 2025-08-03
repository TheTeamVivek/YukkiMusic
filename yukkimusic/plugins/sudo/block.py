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
from yukkimusic.utils.database import add_gban_user, remove_gban_user
from yukkimusic.utils.decorators.language import language

from . import mhelp


@app.on_message(command("BLOCK_COMMAND") & SUDOERS)
@language
async def useradd(client, message: Message, _):
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await app.get_users(user)
        if user.id in BANNED_USERS:
            return await message.reply_text(_["block_1"].format(user.mention))
        await add_gban_user(user.id)
        BANNED_USERS.add(user.id)
        await message.reply_text(_["block_2"].format(user.mention))
        return
    if message.reply_to_message.from_user.id in BANNED_USERS:
        return await message.reply_text(
            _["block_1"].format(message.reply_to_message.from_user.mention)
        )
    await add_gban_user(message.reply_to_message.from_user.id)
    BANNED_USERS.add(message.reply_to_message.from_user.id)
    await message.reply_text(
        _["block_2"].format(message.reply_to_message.from_user.mention)
    )


@app.on_message(command("UNBLOCK_COMMAND") & SUDOERS)
@language
async def userdel(client, message: Message, _):
    if not message.reply_to_message:
        if len(message.command) != 2:
            return await message.reply_text(_["general_1"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await app.get_users(user)
        if user.id not in BANNED_USERS:
            return await message.reply_text(_["block_3"])
        await remove_gban_user(user.id)
        BANNED_USERS.remove(user.id)
        await message.reply_text(_["block_4"])
        return
    user_id = message.reply_to_message.from_user.id
    if user_id not in BANNED_USERS:
        return await message.reply_text(_["block_3"])
    await remove_gban_user(user_id)
    BANNED_USERS.remove(user_id)
    await message.reply_text(_["block_4"])


@app.on_message(command("BLOCKED_COMMAND") & SUDOERS)
@language
async def sudoers_list(client, message: Message, _):
    if not BANNED_USERS:
        return await message.reply_text(_["block_5"])
    mystic = await message.reply_text(_["block_6"])
    msg = _["block_7"]
    count = 0
    for users in BANNED_USERS:
        try:
            user = await app.get_users(users)
            user = user.first_name if not user.mention else user.mention
            count += 1
        except Exception:
            continue
        msg += f"{count}➤ {user}\n"
    if count == 0:
        return await mystic.edit_text(_["block_5"])
    else:
        return await mystic.edit_text(msg)


(
    mhelp.add(
        "en",
        (
            f"<b>✧ {pick_commands('BLOCK_COMMAND')}</b> [Username or reply to a user] - Prevents a user from using bot commands.\n"
            f"<b>✧ {pick_commands('UNBLOCK_COMMAND')}</b> [Username or reply to a user] - Remove a user from the bot's blocked list.\n"
            f"<b>✧ {pick_commands('BLOCKED_COMMAND')}</b> - Check the list of blocked users."
        ),
    )
    .add(
        "ar",
        (
            f"<b>✧ {pick_commands('BLOCK_COMMAND')}</b> [اسم المستخدم أو الرد على المستخدم] - يمنع المستخدم من استخدام أوامر البوت.\n"
            f"<b>✧ {pick_commands('UNBLOCK_COMMAND')}</b> [اسم المستخدم أو الرد على المستخدم] - إزالة المستخدم من قائمة الحظر.\n"
            f"<b>✧ {pick_commands('BLOCKED_COMMAND')}</b> - عرض قائمة المستخدمين المحظورين."
        ),
    )
    .add(
        "as",
        (
            f"<b>✧ {pick_commands('BLOCK_COMMAND')}</b> [ব্যৱহাৰকাৰী নাম বা ব্যৱহাৰকাৰীক প্ৰত্যুত্তৰ কৰক] - এজন ব্যৱহাৰকাৰীক বটৰ কমান্ড ব্যৱহাৰ কৰা পৰা ৰোধ কৰে।\n"
            f"<b>✧ {pick_commands('UNBLOCK_COMMAND')}</b> [ব্যৱহাৰকাৰী নাম বা ব্যৱহাৰকাৰীক প্ৰত্যুত্তৰ কৰক] - এজন ব্যৱহাৰকাৰীক ব্লক তালিকাৰ পৰা আতৰাওক।\n"
            f"<b>✧ {pick_commands('BLOCKED_COMMAND')}</b> - ব্লক কৰা ব্যৱহাৰকাৰীৰ তালিকা চাওক।"
        ),
    )
    .add(
        "hi",
        (
            f"<b>✧ {pick_commands('BLOCK_COMMAND')}</b> [यूज़रनेम या यूज़र को रिप्लाई करें] - किसी यूज़र को बॉट कमांड्स इस्तेमाल करने से रोकें।\n"
            f"<b>✧ {pick_commands('UNBLOCK_COMMAND')}</b> [यूज़रनेम या यूज़र को रिप्लाई करें] - किसी यूज़र को ब्लॉक सूची से हटाएँ।\n"
            f"<b>✧ {pick_commands('BLOCKED_COMMAND')}</b> - ब्लॉक किए गए यूज़र्स की सूची देखें।"
        ),
    )
    .add(
        "ku",
        (
            f"<b>✧ {pick_commands('BLOCK_COMMAND')}</b> [ناوی بەکارهێنەر یان وەڵامدانەوە بە بەکارهێنەر] - بەکارهێنەرێک لە بەکارهێنانی فەرمانەکانی بۆت بەرز دەگرێت.\n"
            f"<b>✧ {pick_commands('UNBLOCK_COMMAND')}</b> [ناوی بەکارهێنەر یان وەڵامدانەوە بە بەکارهێنەر] - بەکارهێنەرێک لە لیستی بلۆک دەربهێنە.\n"
            f"<b>✧ {pick_commands('BLOCKED_COMMAND')}</b> - لیستی بەکارهێنەرانی بلۆک کراو بپشکنە."
        ),
    )
    .add(
        "tr",
        (
            f"<b>✧ {pick_commands('BLOCK_COMMAND')}</b> [Kullanıcı adı veya kullanıcıya yanıt ver] - Bir kullanıcının bot komutlarını kullanmasını engeller.\n"
            f"<b>✧ {pick_commands('UNBLOCK_COMMAND')}</b> [Kullanıcı adı veya kullanıcıya yanıt ver] - Bir kullanıcıyı botun engelli listesinden kaldır.\n"
            f"<b>✧ {pick_commands('BLOCKED_COMMAND')}</b> - Engellenmiş kullanıcıların listesini kontrol et."
        ),
    )
)

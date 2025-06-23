#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os
import shutil

from pyrogram import filters
from pyrogram.enums import ChatType

from config import BANNED_USERS
from strings import command, pick_commands
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_active_chats,
    get_cmode,
    remove_active_chat,
    remove_active_video_chat,
)
from YukkiMusic.utils.decorators import AdminActual

from . import mhelp


@app.on_message(command("REBOOT_COMMAND") & filters.group & ~BANNED_USERS)
@AdminActual
async def reboot(client, message, _):
    mystic = await message.reply_text(
        f"Please Wait... \nRebooting{app.mention} For Your Chat."
    )
    await Yukki.stop_stream(message.chat.id)
    chat_id = await get_cmode(message.chat.id)
    if chat_id:
        await Yukki.stop_stream(chat_id)
    return await mystic.edit_text("Sucessfully Restarted \nTry playing Now..")


@app.on_message(command("RESTART_COMMAND") & ~BANNED_USERS)
async def restart_(client, message):
    if message.from_user and message.from_user.id not in SUDOERS:
        if message.chat.type not in [ChatType.GROUP, ChatType.SUPERGROUP]:
            return
        return await reboot(client, message)
    response = await message.reply_text("Restarting...")
    ac_chats = await get_active_chats()
    for x in ac_chats:
        try:
            await app.send_message(
                chat_id=int(x),
                text=f"{app.mention} Is restarting...\n\nYou can start playing after 15-20 seconds",
            )
            await remove_active_chat(x)
            await remove_active_video_chat(x)
        except Exception:
            pass

    try:
        shutil.rmtree("downloads")
        shutil.rmtree("raw_files")
        shutil.rmtree("cache")
    except Exception:
        pass
    await response.edit_text(
        "Restart process started, please wait for few seconds until the bot starts..."
    )
    os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")


(
    mhelp.add(
        "en",
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'en')}</b> - Reboot the bot in your current chat.",
        priority=13,
    )
    .add(
        "ar",
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'ar')}</b> - إعادة تشغيل البوت في الدردشة الحالية.",
        priority=13,
    )
    .add(
        "as",
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'as')}</b> - বৰ্তমান চেটত বটটো ৰিবুট কৰক।",
        priority=13,
    )
    .add(
        "hi",
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'hi')}</b> - इस चैट में बॉट को रीबूट करें।",
        priority=13,
    )
    .add(
        "ku",
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'ku')}</b> - بۆتەکە لە چاتی ئێستا دا دووبارە دەستپێبکەوە.",
        priority=13,
    )
    .add(
        "tr",
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'tr')}</b> - Botu bu sohbette yeniden başlat.",
        priority=13,
    )
)

(
    mhelp("Sudoers")
    .add(
        "en",
        f"<b>✧ {pick_commands('RESTART_COMMAND', 'en')}</b> - Fully restart the bot and all processes (OS-level).\n"
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'en')}</b> - Reboot current chat. Also triggered if non-sudo uses restart.",
        priority=15,
    )
    .add(
        "ar",
        f"<b>✧ {pick_commands('RESTART_COMMAND', 'ar')}</b> - إعادة تشغيل كاملة للبوت وكل العمليات.\n"
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'ar')}</b> - إعادة تشغيل الدردشة الحالية. يتم تنفيذه أيضًا إذا استخدم غير سودو الأمر restart.",
        priority=15,
    )
    .add(
        "as",
        f"<b>✧ {pick_commands('RESTART_COMMAND', 'as')}</b> - বট আৰু সকলো প্ৰক্ৰিয়া সম্পূৰ্ণৰূপে ৰিষ্টাৰ্ট কৰে।\n"
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'as')}</b> - বৰ্তমান চেট ৰিবুট কৰক। (নন-সুডোৱে restart দিলে এইটোেই হ’ব।)",
        priority=15,
    )
    .add(
        "hi",
        f"<b>✧ {pick_commands('RESTART_COMMAND', 'hi')}</b> - बॉट और सभी प्रक्रियाओं को पूरी तरह से रीस्टार्ट करें।\n"
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'hi')}</b> - वर्तमान चैट को रीबूट करें। (नॉन-सुडो द्वारा restart पर यही चलता है)",
        priority=15,
    )
    .add(
        "ku",
        f"<b>✧ {pick_commands('RESTART_COMMAND', 'ku')}</b> - بۆتەکە و هەموو پڕۆسەکان تەواوەتی دووبارە دەستپێ بکەوە.\n"
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'ku')}</b> - چاتی ئێستا دووبارە بکەرەوە. (ئەگەر نەسودۆیەک restart بەکاربێنێت، ئەمە دەچێتە جێ)",
        priority=15,
    )
    .add(
        "tr",
        f"<b>✧ {pick_commands('RESTART_COMMAND', 'tr')}</b> - Botu ve tüm süreçleri tamamen yeniden başlat.\n"
        f"<b>✧ {pick_commands('REBOOT_COMMAND', 'tr')}</b> - Yalnızca bu sohbeti yeniden başlat. (Sudo olmayan biri restart kullanırsa da tetiklenir)",
        priority=15,
    )
)

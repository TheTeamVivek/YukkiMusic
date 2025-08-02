#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import SUDOERS
from yukkimusic.utils.database import autoend_off, autoend_on

from . import mhelp


@app.on_message(command("AUTOEND_COMMAND") & SUDOERS)
async def auto_end_stream(client, message):
    usage = "**ᴜsᴀɢᴇ:**\n\n/autoend [enable|disable]"
    if len(message.command) != 2:
        return await message.reply_text(usage)
    state = message.text.split(None, 1)[1].strip()
    state = state.lower()
    if state == "enable":
        await autoend_on()
        await message.reply_text(
            "Auto End enabled.\n\nBot will leave voicechat automatically after 30 secinds if one is listening song with a warning message.."
        )
    elif state == "disable":
        await autoend_off()
        await message.reply_text("Autoend disabled")
    else:
        await message.reply_text(usage)


(
    mhelp.add(
        "en",
        f"<b>{pick_commands('AUTOEND_COMMAND')}</b> [enable / disable] - Automatically end the stream after 30s if no one is listening to songs",
    )
    .add(
        "ar",
        f"<b>{pick_commands('AUTOEND_COMMAND')}</b> [enable / disable] - إنهاء البث تلقائيًا بعد 30 ثانية إذا لم يكن هناك أي مستمع",
    )
    .add(
        "as",
        f"<b>{pick_commands('AUTOEND_COMMAND')}</b> [enable / disable] - ৩০ ছেকেণ্ড পিছত কোনো শ্ৰোতা নাথাকিলে স্বয়ংক্ৰিয়ভাৱে ষ্ট্ৰীম বন্ধ কৰিব",
    )
    .add(
        "hi",
        f"<b>{pick_commands('AUTOEND_COMMAND')}</b> [enable / disable] - यदि कोई श्रोता नहीं है तो 30 सेकंड बाद स्वतः स्ट्रीम बंद हो जाएगी",
    )
    .add(
        "ku",
        f"<b>{pick_commands('AUTOEND_COMMAND')}</b> [enable / disable] - بەشێوەی خۆکار دەرچوونی بلەیەک لە دوای 30 چرکە ئەگەر هیچ گوێگرێک نەبێت",
    )
    .add(
        "tr",
        f"<b>{pick_commands('AUTOEND_COMMAND')}</b> [enable / disable] - 30 saniye içinde dinleyici yoksa yayını otomatik olarak sonlandır",
    )
)

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#

import asyncio
import random
from datetime import datetime, timedelta

from pyrogram import filters
from pyrogram.enums import ChatMembersFilter
from pyrogram.errors import FloodWait
from pyrogram.raw import types

import config
from config import adminlist, chatstats, clean, userstats
from strings import command, get_command, pick_command
from yukkimusic import app
from yukkimusic.core.help import ModuleHelp
from yukkimusic.utils.database import (
    get_active_chats,
    get_authuser_names,
    get_client,
    get_particular_top,
    get_served_chats,
    get_served_users,
    get_user_top,
    is_cleanmode_on,
    set_queries,
    update_particular_top,
    update_user_top,
)
from yukkimusic.utils.decorators.language import language
from yukkimusic.utils.formatters import alpha_to_int

AUTO_DELETE = config.CLEANMODE_DELETE_MINS
AUTO_SLEEP = 5
IS_BROADCASTING = False
cleanmode_group = 15


@app.on_raw_update(group=cleanmode_group)
async def clean_mode(client, update, users, chats):
    global IS_BROADCASTING
    if IS_BROADCASTING:
        return
    try:
        if not isinstance(update, types.UpdateReadChannelOutbox):
            return
    except Exception:
        return
    if users:
        return
    if chats:
        return
    message_id = update.max_id
    chat_id = int(f"-100{update.channel_id}")
    if not await is_cleanmode_on(chat_id):
        return
    if chat_id not in clean:
        clean[chat_id] = []
    time_now = datetime.now()
    put = {
        "msg_id": message_id,
        "timer_after": time_now + timedelta(minutes=AUTO_DELETE),
    }
    clean[chat_id].append(put)
    await set_queries(1)


@app.on_message(command("BROADCAST_COMMAND") & filters.user(config.OWNER_ID))
@language
async def braodcast_message(client, message, _):
    global IS_BROADCASTING
    if message.reply_to_message:
        x = message.reply_to_message.id
        y = message.chat.id
    else:
        if len(message.command) < 2:
            return await message.reply_text(_["broad_5"])
        query = message.text.split(None, 1)[1]

        if "-nobot" in query:
            query = query.replace("-nobot", "")
        if "-pinloud" in query:
            query = query.replace("-pinloud", "")
        if "-pin" in query:
            query = query.replace("-pin", "")
        if "-assistant" in query:
            query = query.replace("-assistant", "")
        if "-user" in query:
            query = query.replace("-user", "")
        if query == "":
            return await message.reply_text(_["broad_6"])

    IS_BROADCASTING = True

    # Bot broadcast inside chats
    if "-nobot" not in message.text:
        sent = 0
        pin = 0
        chats = []
        schats = await get_served_chats()
        for chat in schats:
            chats.append(int(chat["chat_id"]))
        for i in chats:
            if i == config.LOG_GROUP_ID:
                continue
            try:
                m = (
                    await app.forward_messages(i, y, x)
                    if message.reply_to_message
                    else await app.send_message(i, text=query)
                )
                if "-pin" in message.text:
                    try:
                        await m.pin(disable_notification=True)
                        pin += 1
                    except Exception:
                        continue
                elif "-pinloud" in message.text:
                    try:
                        await m.pin(disable_notification=False)
                        pin += 1
                    except Exception:
                        continue
                sent += 1
            except FloodWait as e:
                flood_time = int(e.value)
                if flood_time > 200:
                    continue
                await asyncio.sleep(flood_time)
            except Exception:
                continue
        try:
            await message.reply_text(_["broad_1"].format(sent, pin))
        except Exception:
            pass

    # Bot broadcasting to users
    if "-user" in message.text:
        susr = 0
        pin = 0
        served_users = []
        susers = await get_served_users()
        for user in susers:
            served_users.append(int(user["user_id"]))
        for i in served_users:
            try:
                m = (
                    await app.forward_messages(i, y, x)
                    if message.reply_to_message
                    else await app.send_message(i, text=query)
                )
                if "-pin" in message.text:
                    try:
                        await m.pin(both_sides=True, disable_notification=True)
                        pin += 1
                    except Exception:
                        continue
                elif "-pinloud" in message.text:
                    try:
                        await m.pin(both_sides=True, disable_notification=False)
                        pin += 1
                    except Exception:
                        continue
                susr += 1
            except FloodWait as e:
                flood_time = int(e.value)
                if flood_time > 200:
                    continue
                await asyncio.sleep(flood_time)
            except Exception:
                pass
        try:
            await message.reply_text(_["broad_7"].format(susr, pin))
        except Exception:
            pass

    # Bot broadcasting by assistant
    if "-assistant" in message.text:
        aw = await message.reply_text(_["broad_2"])
        text = _["broad_3"]
        from yukkimusic.core.userbot import assistants

        for num in assistants:
            sent = 0
            client = await get_client(num)
            contacts = [user.id for user in await client.get_contacts()]
            async for dialog in client.get_dialogs():
                if dialog.chat.id == config.LOG_GROUP_ID:
                    continue
                if dialog.chat.id in contacts:
                    continue
                try:
                    (
                        await client.forward_messages(dialog.chat.id, y, x)
                        if message.reply_to_message
                        else await client.send_message(dialog.chat.id, text=query)
                    )
                    sent += 1
                except FloodWait as e:
                    flood_time = int(e.value)
                    if flood_time > 200:
                        continue
                    await asyncio.sleep(flood_time)
                except Exception as e:
                    print(e)
                    continue
            text += _["broad_4"].format(num, sent)
        try:
            await aw.edit_text(text)
        except Exception:
            pass
    IS_BROADCASTING = False


async def auto_clean():
    while not await asyncio.sleep(AUTO_SLEEP):
        try:
            for chat_id in chatstats:
                for dic in chatstats[chat_id]:
                    vidid = dic["vidid"]
                    title = dic["title"]
                    chatstats[chat_id].pop(0)
                    spot = await get_particular_top(chat_id, vidid)
                    if spot:
                        spot = spot["spot"]
                        next_spot = spot + 1
                        new_spot = {"spot": next_spot, "title": title}
                        await update_particular_top(chat_id, vidid, new_spot)
                    else:
                        next_spot = 1
                        new_spot = {"spot": next_spot, "title": title}
                        await update_particular_top(chat_id, vidid, new_spot)
            for user_id in userstats:
                for dic in userstats[user_id]:
                    vidid = dic["vidid"]
                    title = dic["title"]
                    userstats[user_id].pop(0)
                    spot = await get_user_top(user_id, vidid)
                    if spot:
                        spot = spot["spot"]
                        next_spot = spot + 1
                        new_spot = {"spot": next_spot, "title": title}
                        await update_user_top(user_id, vidid, new_spot)
                    else:
                        next_spot = 1
                        new_spot = {"spot": next_spot, "title": title}
                        await update_user_top(user_id, vidid, new_spot)
        except Exception:
            continue
        try:
            for chat_id in clean:
                if chat_id == config.LOG_GROUP_ID:
                    continue
                for x in clean[chat_id]:
                    if datetime.now() > x["timer_after"]:
                        try:
                            await app.delete_messages(chat_id, x["msg_id"])
                        except FloodWait as e:
                            await asyncio.sleep(e.value)
                        except Exception:
                            continue
                    else:
                        continue
        except Exception:
            continue
        try:
            served_chats = await get_active_chats()
            for chat_id in served_chats:
                if chat_id not in adminlist:
                    adminlist[chat_id] = []
                    admins = app.get_chat_members(
                        chat_id, filter=ChatMembersFilter.ADMINISTRATORS
                    )
                    async for user in admins:
                        if user.privileges.can_manage_video_chats:
                            adminlist[chat_id].append(user.user.id)
                    authusers = await get_authuser_names(chat_id)
                    for user in authusers:
                        user_id = await alpha_to_int(user)
                        adminlist[chat_id].append(user_id)
        except Exception:
            continue


asyncio.create_task(auto_clean())

# pylint: disable=C0301
(
    ModuleHelp("GCast")
    .name("en", "Broadcast")
    .add(
        "en",
        f"<b>{pick_command('BROADCAST_COMMAND', 'en')} [Message or Reply to any message]</b> » Broadcast a message to served chats of bot\n"
        "<u>Broadcasting Modes:</u>\n\n"
        "<b><code>-pin</code></b> » Pins your broadcasted message in served chats\n\n"
        "<b><code>-pinloud</code></b> » Pins your broadcasted message in served chats and sends notification to the members\n\n"
        "<b><code>-user</code></b> » Broadcast the message to users who have started your bot [You can also pin the message by adding `-pin` or `-pinloud`]\n\n"
        "<b><code>-assistant</code></b> » Broadcast your message through all assistants of the bot\n\n"
        "<b><code>-nobot</code></b> » Ensures that the <b>bot</b> doesn't broadcast the message [Useful when you don't want to broadcast the message to groups]\n\n"
        f"> <b>Example:</b> <code>/{random.choice(get_command('BROADCAST_COMMAND', 'en'))} -user -assistant -pin Testing broadcast</code>",
    )
    .name("ar", "البث")
    .add(
        "ar",
        f"<b>{pick_command('BROADCAST_COMMAND', 'ar')} [رسالة أو الرد على أي رسالة]</b> » بث رسالة إلى الدردشات المقدمة من البوت\n"
        "<u>أوضاع البث:</u>\n\n"
        "<b><code>-pin</code></b> » تثبيت الرسالة المذاعة في الدردشات المقدمة\n\n"
        "<b><code>-pinloud</code></b> » تثبيت الرسالة المذاعة وإرسال إشعار للأعضاء\n\n"
        "<b><code>-user</code></b> » بث الرسالة للمستخدمين الذين بدأوا البوت [يمكنك أيضًا تثبيت الرسالة بإضافة `-pin` أو `-pinloud`]\n\n"
        "<b><code>-assistant</code></b> » بث رسالتك من خلال جميع مساعدي البوت\n\n"
        "<b><code>-nobot</code></b> » التأكد من أن <b>البوت</b> لا يبث الرسالة [مفيد عندما لا ترغب في بث الرسالة إلى المجموعات]\n\n"
        f"> <b>مثال:</b> <code>/{random.choice(get_command('BROADCAST_COMMAND', 'ar'))} -user -assistant -pin اختبار البث</code>",
    )
    .name("hi", "प्रसारण")
    .add(
        "hi",
        f"<b>{pick_command('BROADCAST_COMMAND', 'hi')} [संदेश या किसी भी संदेश का उत्तर]</b> » बॉट द्वारा सेवा किए गए चैट्स में एक संदेश प्रसारित करें\n"
        "<u>प्रसारण मोड:</u>\n\n"
        "<b><code>-pin</code></b> » प्रसारित किए गए संदेश को सेवा किए गए चैट्स में पिन करें\n\n"
        "<b><code>-pinloud</code></b> » प्रसारित किए गए संदेश को पिन करें और सदस्यों को अधिसूचना भेजें\n\n"
        "<b><code>-user</code></b> » संदेश को उन उपयोगकर्ताओं तक प्रसारित करें जिन्होंने आपके बॉट को शुरू किया है [आप संदेश को पिन करने के लिए `-pin` या `-pinloud` भी जोड़ सकते हैं]\n\n"
        "<b><code>-assistant</code></b> » अपने संदेश को बॉट के सभी सहायकों के माध्यम से प्रसारित करें\n\n"
        "<b><code>-nobot</code></b> » सुनिश्चित करता है कि <b>बॉट</b> संदेश को प्रसारित नहीं करता [उपयोगी जब आप संदेश को समूहों में प्रसारित नहीं करना चाहते हैं]\n\n"
        f"> <b>उदाहरण:</b> <code>/{random.choice(get_command('BROADCAST_COMMAND', 'hi'))} -user -assistant -pin परीक्षण प्रसारण</code>",
    )
    .name("as", "প্ৰচাৰ")
    .add(
        "as",
        f"<b>{pick_command('BROADCAST_COMMAND', 'as')} [মেছেজ বা কোনো মেছেজত প্ৰতিক্ৰিয়া]</b> » বটৰ দ্বাৰা পূৰ্বতে সেৱা কৰা চেটসমূহত মেছেজ প্ৰচাৰ কৰক\n"
        "<u>প্ৰচাৰ মড:</u>\n\n"
        "<b><code>-pin</code></b> » প্ৰচাৰ কৰা মেছেজ সেৱা কৰা চেটসমূহত পিন কৰক\n\n"
        "<b><code>-pinloud</code></b> » প্ৰচাৰ কৰা মেছেজ সেৱা কৰা চেটসমূহত পিন কৰক আৰু সদস্যসকলক নিৰ্দেশনা প্ৰেৰণ কৰক\n\n"
        "<b><code>-user</code></b> » বটে আৰম্ভ কৰা ব্যৱহাৰকাৰীসকললৈ মেছেজ প্ৰচাৰ কৰক [আপুনি `-pin` বা `-pinloud` যোগ কৰি মেছেজ পিন কৰিব পাৰে]\n\n"
        "<b><code>-assistant</code></b> » আপোনাৰ মেছেজ বটৰ সকলো সহায়কৰ দ্বাৰা প্ৰচাৰ কৰক\n\n"
        "<b><code>-nobot</code></b> » নিশ্চিত কৰে যে <b>বট</b> মেছেজটো প্ৰচাৰ নকৰে [আপুনি মেছেজটো গোটসমূহলৈ প্ৰচাৰ কৰিবলৈ নাছাহিলে সুবিধাজনক]\n\n"
        f"> <b>উদাহৰণ:</b> <code>/{random.choice(get_command('BROADCAST_COMMAND', 'as'))} -user -assistant -pin পৰীক্ষা প্ৰচাৰ</code>",
    )
    .name("ckb", "بڵاوکردنەوە")
    .add(
        "ckb",
        f"<b>{pick_command('BROADCAST_COMMAND', 'ckb')} [نامە یان وەڵامدانەوەی نامە]</b> » نامەیە بڵاوبکەوە بۆ هەموو گرووپی بۆتەکە\n"
        "<u>جۆرەکانی ناردن:</u>\n\n"
        "<b><code>-pin</code></b> » نامەکەت لە گرووپەکان پین بکە\n\n"
        "<b><code>-pinloud</code></b> » نامەکەت لە گرووپەکان پین بکە و ئاگاداری بنێرە بۆ ئەندامەکان\n\n"
        "<b><code>-user</code></b> » نامە بڵاوبکەوە بۆ ئەوانەی کە بۆتەکەیان لەلایە [دەتوانیت پەیامەکە پین بکەی، تەنها `pin` یان `-pinloud` بەکارببە]\n\n"
        "<b><code>-assistant</code></b> » نامەکەت بڵاوبکەوە لە ڕێگەی یاریدەدەری بۆتەکەوە\n\n"
        "<b><code>-nobot</code></b> » بۆتەکەت وادەکات کە نامەکە بڵاونەکاتەوە [بە سوودە بۆ کاتێک ناتەوێت نامەکە بڵاوبکرێتەوە لە گرووپەکان]\n\n"
        f"> <b>نموونە:</b> <code>/{random.choice(get_command('BROADCAST_COMMAND', 'ckb'))} -user -assistant -pin تاقیکردنەوەی بڵاوکردنەوە</code>",
    )
    .name("tr", "Yayınla")
    .add(
        "tr",
        f"<b>{pick_command('BROADCAST_COMMAND', 'tr')} [Mesaj veya herhangi bir mesaja yanıt]</b> » Botun hizmet verdiği sohbetlere bir mesaj yayınlayın\n"
        "<u>Yayın Modları:</u>\n\n"
        "<b><code>-pin</code></b> » Yayınlanan mesajı hizmet verilen sohbetlerde sabitleyin\n\n"
        "<b><code>-pinloud</code></b> » Yayınlanan mesajı sabitleyin ve üyelere bildirim gönderin\n\n"
        "<b><code>-user</code></b> » Mesajı botu başlatan kullanıcılara yayınlayın [Mesajı sabitlemek için `-pin` veya `-pinloud` ekleyebilirsiniz]\n\n"
        "<b><code>-assistant</code></b> » Mesajınızı botun tüm asistanları aracılığıyla yayınlayın\n\n"
        "<b><code>-nobot</code></b> » <b>Bot</b>'un mesajı yayınlamamasını sağlar [Mesajı gruplara yayınlamak istemediğinizde kullanışlıdır]\n\n"
        f"> <b>Örnek:</b> <code>/{random.choice(get_command('BROADCAST_COMMAND', 'tr'))} -user -assistant -pin Test yayını</code>",
    )
)

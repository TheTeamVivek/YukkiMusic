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
from datetime import datetime, timedelta

from pyrogram.errors import FloodWait
from telethon import events
from telethon.tl.types import ChannelParticipantsAdmins, UpdateReadChannelOutbox

import config
from config import adminlist, chatstats, clean, userstats
from YukkiMusic import tbot
from YukkiMusic.utils.database import (
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
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.formatters import alpha_to_int

AUTO_DELETE = config.CLEANMODE_DELETE_TIME
AUTO_SLEEP = 5
IS_BROADCASTING = False
cleanmode_group = 15


@tbot.on(events.Raw(types=UpdateReadChannelOutbox))
async def clean_mode(event):
    global IS_BROADCASTING
    if IS_BROADCASTING:
        return
    logger.info(update.stringify())
    update = event
    if hasattr(event, "users") and event.users:
        return
    if hasattr(event, "chats") and event.chats:
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


@tbot.on_message(flt.command("BROADCAST_COMMAND", True) & flt.user(config.OWNER_ID))
@language
async def braodcast_message(event, _):
    global IS_BROADCASTING
    if event.is_reply:
        r_msg = await event.get_reply_message()
    else:
        if len(event.text.split()) < 2:
            return await event.reply(_["broad_5"])
        query = event.text.split(None, 1)[1]

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
            return await event.reply(_["broad_6"])

    IS_BROADCASTING = True

    # Bot broadcast inside chats
    if "-nobot" not in event.text:
        sent = 0
        pin = 0
        schats = await get_served_chats()
        chats = [
            int(chat["chat_id"])
            for chat in schats
            if int(chat["chat_id"]) != config.LOG_GROUP_ID
        ]

        for chat_id in chats:
            try:
                m = (
                    await r_msg.forward_to(chat_id)
                    if event.is_reply
                    else await tbot.send_message(chat_id, query)
                )
                if "-pin" in event.text:
                    try:
                        await m.pin(notify="-pinloud" in event.text)
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
            await event.reply(_["broad_1"].format(sent, pin))
        except Exception:
            pass

    # Bot broadcasting to users
    if "-user" in event.text:
        susr = 0
        pin = 0
        served_users = []
        susers = await get_served_users()
        served_users = [int(user["user_id"]) for user in susers]

        for user_id in served_users:

            try:
                m = (
                    await r_msg.forward_to(user_id)
                    if event.is_reply
                    else await tbot.send_message(user_id, query)
                )
                if "-pin" in event.text:
                    try:
                        await m.pin(notify="-pinloud" in event.text)
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
            await event.reply(_["broad_7"].format(susr, pin))
        except Exception:
            pass

    # Bot broadcasting by assistant
    if "-assistant" in event.text:
        aw = await event.reply(_["broad_2"])
        text = _["broad_3"]
        from YukkiMusic.core.userbot import assistants

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
                        await client.forward_messages(
                            dialog.chat.id, event.chat_id, r_msg.id
                        )
                        if message.reply_to_message
                        else await client.send_message(dialog.chat.id, text=query)
                    )
                    sent += 1
                except FloodWait as e:
                    flood_time = int(e.value)
                    if flood_time > 200:
                        continue
                    await asyncio.sleep(flood_time)
                except Exception:
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
                            await tbot.delete_messages(chat_id, x["msg_id"])
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
                    async for user in tbot.iter_participants(
                        chat_id, filter=ChannelParticipantsAdmins
                    ):
                        if (await tbot.get_permissions(chat_id, user)).manage_call:
                            adminlist[chat_id].append(user.id)
                    authusers = await get_authuser_names(chat_id)
                    for user in authusers:
                        user_id = await alpha_to_int(user)
                        adminlist[chat_id].append(user_id)
        except Exception:
            continue


asyncio.create_task(auto_clean())

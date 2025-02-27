#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import random

from telethon import Button, events

from config import LOG, LOG_GROUP_ID
from YukkiMusic import tbot
from YukkiMusic.utils.database import (
    delete_served_chat,
    get_assistant,
    is_on_off,
)

join_msgs = [
    "✨ **New Home Unlocked!** ✨\n\n"
    "🎶 **Bot has joined a new group!**\n"
    "**🏠 Group:** {chat_title}\n"
    "**🆔 ID:** `{chat_id}`\n"
    "**🔗 Username:** {username}\n"
    "**👥 Members:** {member_count}\n"
    "**👤 Added By:** {added_by}",
    "🎉 **Guess What?** 🎉\n\n"
    "🤖 **I've been invited to a new group!**\n"
    "**📍 Location:** {chat_title}\n"
    "**📌 Chat ID:** `{chat_id}`\n"
    "**🔗 Link:** {username}\n"
    "**👥 Population:** {member_count}\n"
    "**🚀 Summoner:** {added_by}",
    "💫 **New Mission Accepted!** 💫\n\n"
    "🎧 **The music has arrived in:** {chat_title}\n"
    "**🆔 Chat ID:** `{chat_id}`\n"
    "**🔗 Username:** {username}\n"
    "**👥 People Here:** {member_count}\n"
    "**✨ Invited By:** {added_by}",
]

leave_msgs = [
    "😢 **The Show is Over!** 😢\n\n"
    "🚪 **Bot has been removed from a group.**\n"
    "**🏠 Group:** {chat_title}\n"
    "**🆔 ID:** `{chat_id}`\n"
    "**🔗 Username:** {username}\n"
    "**👤 Removed By:** {removed_by}",
    "🔕 **Silence Falls...** 🔕\n\n"
    "📍 **I have left the following group:**\n"
    "**🏠 Name:** {chat_title}\n"
    "**🆔 Chat ID:** `{chat_id}`\n"
    "**🔗 Link:** {username}\n"
    "**🚶 Kicked By:** {removed_by}",
    "⚠️ **Mission Terminated!** ⚠️\n\n"
    "🚀 **I've been removed from:** {chat_title}\n"
    "**📌 Chat ID:** `{chat_id}`\n"
    "**🔗 Username:** {username}\n"
    "**👤 Removed By:** {removed_by}",
]


@tbot.on(events.ChatAction)
async def on_chat_action(event):
    chat = await event.get_chat()
    username = f"@{chat.username}" if chat.username else "Private Chat"
    chat_title = chat.title
    chat_id = chat.id
    member_count = chat.participants_count
    userbot = await get_assistant(chat.id)

    if event.user_added:
        for user in event.users:
            if user.id == tbot.id:
                added_by = await event.get_added_by()
                added_by = f"**{await tbot.create_mention(added_by)}**"
                msg = random.choice(join_msgs).format(
                    chat_title=chat_title,
                    chat_id=chat_id,
                    username=username,
                    member_count=member_count,
                    added_by=added_by,
                )

                if await is_on_off(LOG):
                    await tbot.send_message(
                        LOG_GROUP_ID,
                        msg,
                        buttons=(
                            [
                                [
                                    Button.url(
                                        f"🔍 View {chat_title}",
                                        f"https://t.me/{chat.username}",
                                    )
                                ]
                            ]
                            if chat.username
                            else None
                        ),
                    )

                if chat.username:
                    await userbot.join_chat(chat.username)

    elif event.user_left:
        for user in event.users:
            if user.id == tbot.id:
                removed_by = await event.get_kicked_by()
                removed_by = (
                    f"**{await tbot.create_mention(removed_by)}**"
                    if removed_by
                    else "Unknown User"
                )
                msg = random.choice(leave_msgs).format(
                    chat_title=chat_title,
                    chat_id=chat_id,
                    username=username,
                    removed_by=removed_by,
                )

                if await is_on_off(LOG):
                    await tbot.send_message(
                        LOG_GROUP_ID,
                        msg,
                        buttons=(
                            [
                                [
                                    Button.url(
                                        f"🔍 View {chat_title}",
                                        f"https://t.me/{chat.username}",
                                    )
                                ]
                            ]
                            if chat.username
                            else None
                        ),
                    )

                if userbot:
                    await userbot.leave_chat(chat.id)

                await delete_served_chat(chat_id)


# Kindly Not usinf try except because i want to see errors that is it working correctly or not

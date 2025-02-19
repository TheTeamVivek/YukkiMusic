#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from YukkiMusic import tbot
from YukkiMusic.core import filters
from YukkiMusic.misc import db
from YukkiMusic.utils.database.memorydatabase import (
    get_active_chats,
    get_active_video_chats,
    remove_active_chat,
    remove_active_video_chat,
)


# Function for removing the Active voice and video chat also clear the db dictionary for the chat
async def _clear_(chat_id):
    db[chat_id] = []
    await remove_active_video_chat(chat_id)
    await remove_active_chat(chat_id)


@tbot.on_message(
    filters.command("ACTIVEVC_COMMAND", use_strings=True)
    & filters.user(list(BANNED_USERS))
)
async def active_audio(event):
    mystic = await event.reply("Getting Active Voicechats....\nPlease hold on")
    served_chats = await get_active_chats()
    text = ""
    j = 0
    for x in served_chats:
        try:
            chat = await tbot.get_entity(x)
            title = chat.title
            if getattr(chat, "username"):
                text += (
                    f"**{j + 1}.**  [{title}](https://t.me/{chat.username})[`{x}`]\n"
                )
            else:
                text += f"**{j + 1}. {title}** [`{x}`]\n"
            j += 1
        except ValueError as e:
            await tbot.handle_error(e)  # Just for verifying the error
            # await _clear_(x)
            continue
    if not text:
        await mystic.edit("No active Chats Found")
    else:
        await mystic.edit(
            f"**Active Voice Chat's:-**\n\n{text}",
            link_preview=False,
        )


@tbot.on_message(
    filters.command("ACTIVEVIDEO_COMMAND", use_strings=True)
    & filters.user(list(BANNED_USERS))
)
async def active_video(event):
    mystic = await event.reply("Getting Active Voicechats....\nPlease hold on")
    served_chats = await get_active_video_chats()
    text = ""
    j = 0
    for x in served_chats:
        try:
            chat = await tbot.get_entity(x)
            title = chat.title
            if getattr(chat, "username"):
                text += f"**{j + 1}.** [{title}](https://t.me/{chat.username})[`{x}`]\n"
            else:
                text += f"**{j + 1}. {title}** [`{x}`]\n"
            j += 1
        except ValueError:
            #  await _clear_(x)
            continue
    if not text:
        await mystic.edit("No active Chats Found")
    else:
        await mystic.edit(
            f"**Active Video Chat's:-**\n\n{text}",
            link_preview=False,
        )


@tbot.on_message(
    filters.command("AC_COMMAND", use_strings=True) & filters.user(list(BANNED_USERS))
)
async def ac_counts(event):
    ac_audio = len(await get_active_chats())
    ac_video = len(await get_active_video_chats())
    total_audio = int(ac_audio - ac_video)
    await event.reply(f"Active Chats info:\nAudio: {total_audio}\nVideo: {ac_video}")

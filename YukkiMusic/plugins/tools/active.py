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
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils import get_active_chats, get_active_video_chats,  language


@tbot.on_message(filters.command("ACTIVEVC_COMMAND", True) & SUDOERS)
@language(no_check=True)
async def active_audio(event, _):
    mystic = await event.reply(_["ac_1"])
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
            # await clear(x)
            continue
    if not text:
        await mystic.edit(_["ac_2"])
    else:
        await mystic.edit(
            _["ac_3"] + text,
            link_preview=False,
        )


@tbot.on_message(filters.command("ACTIVEVIDEO_COMMAND", True) & SUDOERS)
@language(no_check=True)
async def active_video(event, _):
    mystic = await event.reply(_["ac_1"])
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
            #  await clear(x)
            continue
    if not text:
        await mystic.edit(_["ac_2"])
    else:
        await mystic.edit(
            _["ac_3"] + text,
            link_preview=False,
        )


@tbot.on_message(filters.command("AC_COMMAND", True) & SUDOERS)
@language(no_check=True)
async def ac_counts(event, _):
    ac_audio = len(await get_active_chats())
    ac_video = len(await get_active_video_chats())
    total_audio = ac_audio - ac_video
    await event.reply(_["ac_4"].format(total_audio, ac_video))

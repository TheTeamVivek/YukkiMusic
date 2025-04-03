#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from telethon import Button
from telethon.tl.types import PeerUser

from config import BANNED_USERS, EXTRA_PLUGINS, adminlist
from strings import get_string
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import SUDOERS, db
from YukkiMusic.utils.database import (
    delete_filter,
    get_cmode,
    get_lang,
    is_active_chat,
    is_commanddelete_on,
    is_maintenance,
    is_nonadmin_chat,
    set_loop,
)


@tbot.on_message(
    flt.command("STOP_COMMAND", True) & flt.group & ~flt.user(BANNED_USERS)
)
async def stop_music(event):
    if await is_maintenance() is False:
        if event.sender_id not in SUDOERS:
            return
    comm = event.text.split()
    if not len(comm) < 2:
        if EXTRA_PLUGINS:
            if not comm[0][1] == "c" and not comm[0][1] == "e":
                filter = " ".join(comm[1:])
                deleted = await delete_filter(event.chat_id, filter)
                if deleted:
                    return await event.reply(f"**ᴅᴇʟᴇᴛᴇᴅ ғɪʟᴛᴇʀ {filter}.**")
                else:
                    return await event.reply("**ɴᴏ sᴜᴄʜ ғɪʟᴛᴇʀ.**")

    if await is_commanddelete_on(event.chat_id):
        try:
            await event.delete()
        except Exception:
            pass
    try:
        language = await get_lang(event.chat_id)
        _ = get_string(language)
    except Exception:
        _ = get_string("en")

    if not isinstance(event.message.from_id, PeerUser):
        upl = [
            [
                Button.inline(
                    text="How to Fix this? ",
                    data="AnonymousAdmin",
                ),
            ]
        ]
        return await event.reply(_["general_4"], buttons=upl)

    if comm[0][1] == "c":
        chat_id = await get_cmode(event.chat_id)
        if chat_id is None:
            return await event.reply(_["setting_12"])
        try:
            await tbot.get_entity(chat_id)
        except Exception:
            return await event.reply(_["cplay_4"])
    else:
        chat_id = event.chat_id
    if not await is_active_chat(chat_id):
        return await event.reply(_["general_6"])
    is_non_admin = await is_nonadmin_chat(event.chat_id)
    if not is_non_admin:
        if event.sender_id not in SUDOERS:
            admins = adminlist.get(event.chat_id)
            if not admins:
                return await event.reply(_["admin_18"])
            else:
                if event.sender_id not in admins:
                    return await event.reply(_["admin_19"])
    try:
        check = db.get(chat_id)
        if check[0].get("mystic"):
            await check[0].get("mystic").delete()
    except Exception:
        pass
    await Yukki.stop_stream(chat_id)
    await set_loop(chat_id, 0)
    mention = await tbot.create_mention(await event.get_sender())
    await event.reply(_["admin_9"].format(mention))

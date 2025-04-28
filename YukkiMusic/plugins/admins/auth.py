#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from config import adminlist
from YukkiMusic import app, tbot
from YukkiMusic.misc import BANNED_USERS
from YukkiMusic.core import filters as flt
from YukkiMusic.utils.database import (
    delete_authuser,
    get_authuser,
    get_authuser_names,
    save_authuser,
)
from YukkiMusic.utils.decorators import admin_actual, language
from YukkiMusic.utils.formatters import int_to_alpha


@tbot.on_message(flt.command("AUTH_COMMAND", True) & flt.group & ~BANNED_USERS)
@admin_actual
async def auth(event, _):
    await event.get_sender()
    if not event.reply_to:
        if len(event.text.split()) != 2:
            return await event.reply(_["general_1"])
        user = event.text.split(None, 1)[1]
        user = await app.get_entity(user)
        token = await int_to_alpha(user.id)
        from_user_name = event.sender.first_name
        from_user_id = event.sender.id
        _check = await get_authuser_names(event.chat_id)
        count = len(_check)
        if int(count) == 20:
            return await event.reply(_["auth_1"])
        if token not in _check:
            assis = {
                "auth_user_id": user.id,
                "auth_name": user.first_name,
                "admin_id": from_user_id,
                "admin_name": from_user_name,
            }
            get = adminlist.get(event.chat_id)
            if get:
                if user.id not in get:
                    get.append(user.id)
            await save_authuser(event.chat_id, token, assis)
            return await event.reply(_["auth_2"])
        else:
            await event.reply(_["auth_3"])
        return
    replied_msg = await event.get_reply_message()
    from_user_id = event.sender_id
    user = await replied_msg.get_sender()
    user_id = user.id
    user_name = user.first_name
    token = await int_to_alpha(user_id)
    from_user_name = event.sender.first_name
    _check = await get_authuser_names(event.chat_id)
    count = 0
    for smex in _check:
        count += 1
    if int(count) == 20:
        return await event.reply(_["auth_1"])
    if token not in _check:
        assis = {
            "auth_user_id": user_id,
            "auth_name": user_name,
            "admin_id": from_user_id,
            "admin_name": from_user_name,
        }
        get = adminlist.get(event.chat_id)
        if get:
            if user_id not in get:
                get.append(user_id)
        await save_authuser(event.chat_id, token, assis)
        return await event.reply(_["auth_2"])
    else:
        await event.reply(_["auth_3"])


@tbot.on_message(flt.command("UNAUTH_COMMAND", True) & flt.group & ~BANNED_USERS)
@admin_actual
async def unauthusers(event, _):
    if not event.reply_to:
        if len(event.text.split()) != 2:
            return await event.reply(_["general_1"])
        user = event.text.split(None, 1)[1]
        user = await app.get_entity(user)
        token = await int_to_alpha(user.id)
        deleted = await delete_authuser(event.chat_id, token)
        get = adminlist.get(event.chat_id)
        if get:
            if user.id in get:
                get.remove(user.id)
        if deleted:
            return await event.reply(_["auth_4"])
        else:
            return await event.reply(_["auth_5"])
    r_msg = await event.get_reply_message()
    r_user = await r_msg.get_sender()
    user_id = r_user.id
    token = await int_to_alpha(user_id)
    deleted = await delete_authuser(event.chat_id, token)
    get = adminlist.get(event.chat_id)
    if get:
        if user_id in get:
            get.remove(user_id)
    if deleted:
        return await event.reply(_["auth_4"])
    else:
        return await event.reply(_["auth_5"])


@tbot.on_message(flt.command("AUTHUSERS_COMMAND", True) & flt.group & ~BANNED_USERS)
@language
async def authusers(event, _):
    _playlist = await get_authuser_names(event.chat_id)
    if not _playlist:
        return await event.reply(_["setting_5"])
    else:
        j = 0
        mystic = await event.reply(_["auth_6"])
        text = _["auth_7"]
        for note in _playlist:
            _note = await get_authuser(event.chat_id, note)
            user_id = _note["auth_user_id"]
            admin_id = _note["admin_id"]
            admin_name = _note["admin_name"]
            try:
                user = await app.get_entity(user_id)
                user = user.first_name
                j += 1
            except Exception:
                continue
            text += f"{j}➤ {user}[`{user_id}`]\n"
            text += f"   {_['auth_8']} {admin_name}[`{admin_id}`]\n\n"
        await mystic.delete()
        await event.reply(text)

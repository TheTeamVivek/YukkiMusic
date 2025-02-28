#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from config import BANNED_USERS
from YukkiMusic import tbot
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import add_gban_user, remove_gban_user
from YukkiMusic.utils.decorators.language import language


@tbot.on_message(flt.command("BLOCK_COMMAND", True) & flt.user(SUDOERS))
@language
async def useradd(event, _):
    if not event.is_reply:
        if len(event.text.split()) != 2:
            return await event.reply(_["USER_IDENTIFIER_REQUIRED"])
        user = event.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await tbot.get_entity(user)
        mention = await tbot.create_mention(user)
        if user.id in BANNED_USERS:
            return await event.reply(_["block_1"].format(mention))
        await add_gban_user(user.id)
        BANNED_USERS.add(user.id)
        await event.reply(_["block_2"].format(mention))
        return
    reply = await tbot.get_reply_message()
    mention = await tbot.create_mention(await reply.get_sender())
    if reply.sender_id in BANNED_USERS:
        return await event.reply(_["block_1"].format(mention))
    await add_gban_user(reply.sender_id)
    BANNED_USERS.add(reply.sender_id)
    await event.reply(_["block_2"].format(mention))


@tbot.on_message(flt.command("UNBLOCK_COMMAND", True) & flt.user(SUDOERS))
@language
async def userdel(event, _):
    if not event.is_reply:
        if len(event.text.split()) != 2:
            return await event.reply(_["USER_IDENTIFIER_REQUIRED"])
        user = message.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await tbot.get_entity(user)
        if user.id not in BANNED_USERS:
            return await event.reply(_["block_3"])
        await remove_gban_user(user.id)
        BANNED_USERS.remove(user.id)
        await event.reply(_["block_4"])
        return
    reply = await tbot.get_reply_message()
    user_id = reply.sender_id
    if user_id not in BANNED_USERS:
        return await event.reply(_["block_3"])
    await remove_gban_user(user_id)
    BANNED_USERS.remove(user_id)
    await event.reply(_["block_4"])


@tbot.on_message(flt.command("BLOCKED_COMMAND", True) & ~flt.user(BANNED_USERS))
@language
async def sudoers_list(event, _):
    if not BANNED_USERS:
        return await event.reply(_["block_5"])

    mystic = await event.reply(_["block_6"])
    msg = _["block_7"]

    for count, user_id in enumerate(BANNED_USERS, start=1):
        try:
            user = await tbot.get_entity(user_id)
            mention = await tbot.create_mention(user)
        except Exception:
            continue
        msg += f"{count}➤ {mention} ({user.id})\n"

    return await mystic.edit(msg if count else _["block_5"])

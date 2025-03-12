#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from config import BANNED_USERS, MONGO_DB_URI, OWNER_ID
from YukkiMusic import tbot
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import add_sudo, remove_sudo
from YukkiMusic.utils.decorators.language import language


@tbot.on_message(flt.command("ADDSUDO_COMMAND", True) & flt.user(OWNER_ID))
@language
async def useradd(event, _):
    if MONGO_DB_URI is None:
        return await event.reply(
            "**Due to privacy issues, You can't manage sudoers when you are on Yukki Database.\n\n Please fill Your MONGO_DB_URI in your vars to use this features**"
        )
    if not event.is_reply:
        if len(event.text.split()) != 2:
            return await event.reply(_["general_1"])
        user = event.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await tbot.get_entity(user)
        mention = await tbot.create_mention(user)
        if user.id in SUDOERS:
            return await event.reply(_["sudo_1"].format(mention))
        added = await add_sudo(user.id)
        if added:
            SUDOERS.add(user.id)
            await event.reply(_["sudo_2"].format(mention))
        else:
            await event.reply("Something wrong happened")
        return
    rmsg = await event.get_reply_message()
    user_id = rmsg.sender_id
    mention = await tbot.create_mention(user_id)
    if rmsg.sender_id in SUDOERS:
        return await event.reply(_["sudo_1"].format(mention))
    added = await add_sudo(user_id)
    if added:
        SUDOERS.add(user_id)
        await event.reply(_["sudo_2"].format(mention))
    else:
        await event.reply("Something wrong happened")
    return


@tbot.on_message(flt.command("DELSUDO_COMMAND", True) & flt.user(OWNER_ID))
@language
async def userdel(event, _):
    if MONGO_DB_URI is None:
        return await event.reply(
            "**Due to privacy issues, You can't manage sudoers when you are on Yukki Database.\n\n Please fill Your MONGO_DB_URI in your vars to use this features**"
        )
    if not event.is_reply:
        if len(event.text.split()) != 2:
            return await event.reply(_["general_1"])
        user = event.text.split(None, 1)[1]
        if "@" in user:
            user = user.replace("@", "")
        user = await tbot.get_entity(user)
        if user.id not in SUDOERS:
            return await event.reply(_["sudo_3"])
        removed = await remove_sudo(user.id)
        if removed:
            SUDOERS.remove(user.id)
            await event.reply(_["sudo_4"])
            return
        await event.reply(f"Something wrong happened")
        return
    rmsg = await event.get_reply_message()
    user_id = rmsg.sender_id
    if user_id not in SUDOERS:
        return await event.reply(_["sudo_3"])
    removed = await remove_sudo(user_id)
    if removed:
        SUDOERS.remove(user_id)
        await event.reply(_["sudo_4"])
        return
    await event.reply(f"Something wrong happened")


@tbot.on_message(flt.command("SUDOUSERS_COMMAND", True) & ~flt.user(BANNED_USERS))
@language
async def sudoers_list(event, _):
    text = _["sudo_5"]
    count = 0
    for x in OWNER_ID:
        try:
            user = await tbot.get_entity(x)
            user = await tbot.create_mention(user)
            count += 1
        except Exception:
            continue
        text += f"{count}➤ {user} (`{x}`)\n"
    smex = 0
    for user_id in SUDOERS:
        if user_id not in OWNER_ID:
            try:
                user = await tbot.get_entity(user_id)
                user = await tbot.create_mention(user)
                if smex == 0:
                    smex += 1
                    text += _["sudo_6"]
                count += 1
                text += f"{count}➤ {user} (`{user_id}`)\n"
            except Exception:
                continue
    if not text:
        await event.reply(_["sudo_7"])
    else:
        await event.reply(text)

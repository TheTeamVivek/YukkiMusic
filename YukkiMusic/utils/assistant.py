#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from YukkiMusic import app
from YukkiMusic.utils.database import get_client


async def get_assistant_details():
    ms = ""
    msg = "ᴜsᴀsɢᴇ : /setassistant [ᴀssɪsᴛᴀɴᴛ ɴᴏ ] ᴛᴏ ᴄʜᴀɴɢᴇ ᴀɴᴅ sᴇᴛ ᴍᴀɴᴜᴀʟʟʏ ɢʀᴏᴜᴘ ᴀssɪsᴛᴀɴᴛ \n ʙᴇʟᴏᴡ sᴏᴍᴇ ᴀᴠᴀɪʟᴀʙʟᴇ ᴀssɪsᴛᴀɴᴛ ᴅᴇᴛᴀɪʟ's ᴏɴ ʙᴏᴛ sᴇʀᴠᴇʀ\n"
    try:
        a = await get_client(1)
        msg += f"ᴀssɪsᴛᴀɴᴛ ɴᴜᴍʙᴇʀ:- `1` ɴᴀᴍᴇ :- [{a.name}](https://t.me/{a.username})  ᴜsᴇʀɴᴀᴍᴇ :-  @{a.username} ɪᴅ :- {a.id}\n\n"
    except:
        pass

    try:
        b = await get_client(2)
        msg += f"ᴀssɪsᴛᴀɴᴛ ɴᴜᴍʙᴇʀ:- `2` ɴᴀᴍᴇ :- [{b.name}](https://t.me/{b.username})  ᴜsᴇʀɴᴀᴍᴇ :-  @{b.username} ɪᴅ :- {b.id}\n"
    except:
        pass

    try:
        c = await get_client(3)
        msg += f"ᴀssɪsᴛᴀɴᴛ ɴᴜᴍʙᴇʀ:- `3` ɴᴀᴍᴇ :- [{c.name}](https://t.me/{c.username})  ᴜsᴇʀɴᴀᴍᴇ :-  @{c.username} ɪᴅ :- {c.id}\n"
    except:
        pass

    try:
        d = await get_client(4)
        msg += f"ᴀssɪsᴛᴀɴᴛ ɴᴜᴍʙᴇʀ:- `4` ɴᴀᴍᴇ :- [{d.name}](https://t.me/{d.username})  ᴜsᴇʀɴᴀᴍᴇ :-  @{d.username} ɪᴅ :- {d.id}\n"
    except:
        pass

    try:
        e = await get_client(5)
        msg += f"ᴀssɪsᴛᴀɴᴛ ɴᴜᴍʙᴇʀ:- `5` ɴᴀᴍᴇ :- [{e.name}](https://t.me/{e.username})  ᴜsᴇʀɴᴀᴍᴇ :-  @{e.username} ɪᴅ :- {e.id}\n"
    except:
        pass

    return msg


async def assistant():
    from config import STRING1, STRING2, STRING3, STRING4, STRING5

    filled_count = sum(
        1
        for var in [STRING1, STRING2, STRING3, STRING4, STRING5]
        if var and var.strip()
    )
    if filled_count == 1:
        return True
    else:
        return False



from pyrogram import filters
from functools import wraps
from YukkiMusic import userbot

clients = [userbot.one, userbot.two, userbot.three, userbot.four, userbot.five]

def userbot_on_cmd(commands, other_filters=None):
    def decorator(func):
        @wraps(func)
        async def wrapper(client, message, *args, **kwargs):
            return await func(client, message, *args, **kwargs)

        for client in clients:
            combined_filters = filters.command(commands, ".")
            
            if other_filters:
                combined_filters &= other_filters

            client.on_message(combined_filters)(wrapper)

        return wrapper

    return decorator


# @userbot_on_cmd(["ckv"], SUDOERS)
# async def clean(client, message):
   # await message.reply_text("Cleaning in progress...")

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
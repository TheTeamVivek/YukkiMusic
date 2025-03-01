#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import set_video_limit
from YukkiMusic.utils.decorators.language import language


@app.on_message(flt.command("VIDEOLIMIT_COMMAND", True) & flt.user(SUDOERS))
@language
async def set_video_limit_kid(event, _):
    if len(event.text.split()) != 2:
        usage = _["vid_1"]
        return await event.reply(usage)
    state = event.text.split(None, 1)[1].strip()
    if state.lower() == "disable":
        limit = 0
        await set_video_limit(limit)
        return await event.reply(_["vid_4"])
    if state.isnumeric():
        limit = int(state)
        await set_video_limit(limit)
        if limit == 0:
            return await event.reply(_["vid_4"])
        await event.reply(_["vid_3"].format(limit))
    else:
        return await event.reply(_["vid_2"])

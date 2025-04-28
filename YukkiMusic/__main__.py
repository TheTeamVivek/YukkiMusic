#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
import asyncio
import logging
import sys

from pytgcalls.exceptions import NoActiveGroupCall

import config
from config import BANNED_USERS, fetch_cookies
from YukkiMusic import tbot, userbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import sudo
from YukkiMusic.utils.database import get_banned_users, get_gbanned

logger = logging.getLogger("YukkiMusic")
loop = asyncio.get_event_loop()


async def init():
    if not config.STRING_SESSIONS:
        logger.error("No Assistant Clients Vars Defined!.. Exiting Process.")
        return
    if not config.SPOTIFY_CLIENcT_ID and not config.SPOTIFY_CLIENT_SECRET:
        logger.warning(
            "No Spotify Vars defined.Your bot won't be able to play spotify queries."
        )

    try:
        users = await get_gbanned()
        for user_id in users:
            BANNED_USERS.add(user_id)
        users = await get_banned_users()
        for user_id in users:
            BANNED_USERS.add(user_id)
    except Exception:
        pass
    await sudo()
    await tbot.start()

    await tbot.load_plugins_from("YukkiMusic/plugins")
    logger.info("Successfully Imported All Modules ")
    await fetch_cookies()
    await userbot.start()
    await Yukki.start()
    logger.info("Assistant Started Sucessfully")
    try:
        await Yukki.stream_call(
            "http://docs.evostream.com/sample_content/assets/sintel1m720p.mp4"
        )
    except NoActiveGroupCall:
        logger.error("Please ensure there are a voice call,In your log group active.")
        sys.exit()
    logger.info("YukkiMusic Started Successfully")
    tbot.run_until_disconnected()
    await app.stop()
    await userbot.stop()
    await Yukki.stop()


if __name__ == "__main__":
    loop.run_until_complete(init())
    logger.info("Stopping YukkiMusic! GoodBye")

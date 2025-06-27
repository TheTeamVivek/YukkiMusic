#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
# pylint: disable=missing-module-docstring, missing-function-docstring
import asyncio
import logging
import sys

from pyrogram import idle
from pytgcalls.exceptions import NoActiveGroupCall

import config
from config import BANNED_USERS
from yukkimusic import app, userbot
from yukkimusic.core.call import yukki
from yukkimusic.core.help import ModuleHelp
from yukkimusic.core.modules import LoaderContext, load_mod
from yukkimusic.core.mongo import mongodb
from yukkimusic.misc import sudo
from yukkimusic.utils.database import get_banned_users, get_gbanned

logger = logging.getLogger("yukkimusic")


async def init():
    if not config.STRING_SESSIONS:
        logger.error("No Assistant Clients Vars Defined!.. Exiting Process.")
        return
    if not config.SPOTIFY_CLIENT_ID and not config.SPOTIFY_CLIENT_SECRET:
        logger.warning(
            "No Spotify Vars defined. Your bot won't be able to play spotify queries."
        )
    try:
        logger.debug("Loading Banned users...")
        users = await get_gbanned()
        for user_id in users:
            BANNED_USERS.add(user_id)
        users = await get_banned_users()
        for user_id in users:
            BANNED_USERS.add(user_id)
        logger.debug("Sucessfully loaded banned users")
    except Exception as e:  # pylint: disable=broad-exception-caught
        logger.debug("Failed loaded banned users %s", e)
    await sudo()
    await app.start()
    if config.EXTRA_PLUGINS:
        await load_mod(
            config.EXTRA_PLUGINS,
            LoaderContext(
                app=app,
                userbot=userbot,
                mongodb=mongodb,
                help=ModuleHelp,
            ),
        )
    logger.info("Successfully Imported All Modules ")
    await userbot.start()
    await yukki.start()
    logger.info("Assistant Started Sucessfully")
    try:
        await yukki.stream_call(
            "http://docs.evostream.com/sample_content/assets/sintel1m720p.mp4"
        )
    except NoActiveGroupCall:
        logger.error("Please ensure the voice call in your log group is active.")
        sys.exit()

    await yukki.decorators()
    logger.info("yukkimusic Started Successfully")
    await idle()
    await app.stop()
    await userbot.stop()
    await yukki.stop()


def main():
    asyncio.get_event_loop().run_until_complete(init())
    logger.info("Stopping yukkimusic! GoodBye")


if __name__ == "__main__":
    main()

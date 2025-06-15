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
import os

from pyrogram import idle
from pytgcalls.exceptions import NoActiveGroupCall

import config
import YukkiMusic.plugins
from config import BANNED_USERS
from YukkiMusic import HELPABLE, app, userbot
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import sudo
from YukkiMusic.utils.database import get_banned_users, get_gbanned

logger = logging.getLogger("YukkiMusic")
loop = asyncio.get_event_loop()

_ = YukkiMusic.plugins


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
    except Exception as e:
        logger.debug("Failed loaded banned users %s", e)
    await sudo()
    await app.start()
    if config.EXTRA_PLUGINS:
        if os.path.exists("xtraplugins"):
            result = await app.run_shell_command(["git", "-C", "xtraplugins", "pull"])
            if result["returncode"] != 0:
                logger.error(
                    f"Error pulling updates for extra plugins: {result['stderr']}"
                )
                exit()
        else:
            result = await app.run_shell_command(
                ["git", "clone", config.EXTRA_PLUGINS_REPO, "xtraplugins"]
            )
            if result["returncode"] != 0:
                logger.error(f"Error cloning extra plugins: {result['stderr']}")
                exit()

        req = os.path.join("xtraplugins", "requirements.txt")
        if os.path.exists(req):
            result = await app.run_shell_command(
                ["uv", "pip", "install", "--system", "-r", req]
            )
            if result["returncode"] != 0:
                logger.error(f"Error installing requirements: {result['stderr']}")

        for mod in app.load_plugins_from("xtraplugins"):
            if mod and hasattr(mod, "__MODULE__") and mod.__MODULE__:
                if hasattr(mod, "__HELP__") and mod.__HELP__:
                    HELPABLE[mod.__MODULE__.lower()] = mod

    logger.info("Successfully Imported All Modules ")
    await userbot.start()
    await Yukki.start()
    logger.info("Assistant Started Sucessfully")
    try:
        await Yukki.stream_call(
            "http://docs.evostream.com/sample_content/assets/sintel1m720p.mp4"
        )
    except NoActiveGroupCall:
        logger.error("Please ensure the voice call in your log group is active.")
        exit()

    await Yukki.decorators()
    logger.info("YukkiMusic Started Successfully")
    await idle()
    await app.stop()
    await userbot.stop()
    await Yukki.stop()


def main():
    loop.run_until_complete(init())
    logger.info("Stopping YukkiMusic! GoodBye")


if __name__ == "__main__":
    main()

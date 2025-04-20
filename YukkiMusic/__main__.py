#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
import os
import sys

from pytgcalls.exceptions import NoActiveGroupCall

import config
from config import BANNED_USERS, fetch_cookies
from YukkiMusic import app, logger, tbot 
from YukkiMusic.core.call import Yukki
from YukkiMusic.utils.database import get_banned_users, get_gbanned

logger = logger("YukkiMusic")


async def init():
    if not config.STRING_SESSIONS:
        logger.error("No Assistant Clients Vars Defined!.. Exiting Process.")
        return
    if not config.SPOTIFY_CLIENcT_ID and not config.SPOTIFY_CLIENT_SECRET:
        logger.warning(
            "No Spotify Vars defined. Your bot won't be able to play spotify queries." 
        )
    from YukkiMusic import userbot
    try:
        users = await get_gbanned()
        for user_id in users:
            BANNED_USERS.add(user_id)
        users = await get_banned_users()
        for user_id in users:
            BANNED_USERS.add(user_id)
    except Exception:
        pass
    await app.start()
    await tbot.start()

    attrs = {"userbot": userbot, "tbot": tbot}
    await app.load_plugins_from("YukkiMusic/plugins", attrs)

    if config.EXTRA_PLUGINS:
        if os.path.exists("xtraplugins"):
            result = await app.run_shell_command(["git", "-C", "xtraplugins", "pull"])
            if result.returncode != 0:
                logger.error(
                    "Error pulling updates for extra plugins:\n %s", result["stderr"]
                )
                sys.exit()
        else:
            result = await app.run_shell_command(
                ["git", "clone", config.EXTRA_PLUGINS_REPO, "xtraplugins"]
            )
            if result.returncode != 0:
                logger.error("Error cloning extra plugins:\n%s", result["stderr"])
                sys.exit()

        req = os.path.join("xtraplugins", "requirements.txt")
        if os.path.exists(req):
            result = await app.run_shell_command(["pip", "install", "-r", req])
            if result.returncode != 0:
                logger.error("Error installing requirements:\n %s", result["stderr"])

        await app.load_plugins_from("xtraplugins", attrs)

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
        logger.error("Please ensure the voice call in your log group is active.")
        sys.exit()
    logger.info("YukkiMusic Started Successfully")
    tbot.run_until_disconnected()
    await app.stop()
    await userbot.stop()
    await Yukki.stop()


if __name__ == "__main__":
    app.run(init())
    logger.info("Stopping YukkiMusic! GoodBye")

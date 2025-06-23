#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.


import logging
import socket
import time

import heroku3
from pyrogram import filters

import config
from YukkiMusic.core.mongo import mongodb

SUDOERS = filters.user()


HAPP = None
_boot_ = time.time()
logger = logging.getLogger(__name__)
db = {}


def is_heroku():
    return "heroku" in socket.getfqdn()


async def sudo():
    if config.MONGO_DB_URI is None:
        for user_id in config.OWNER_ID:
            SUDOERS.add(user_id)
    else:
        sudoersdb = mongodb.sudoers
        db_sudoers = await sudoersdb.find_one({"sudo": "sudo"})
        db_sudoers = [] if not db_sudoers else db_sudoers["sudoers"]
        for user_id in config.OWNER_ID:
            SUDOERS.add(user_id)
            if user_id not in db_sudoers:
                db_sudoers.append(user_id)
                await sudoersdb.update_one(
                    {"sudo": "sudo"},
                    {"$set": {"sudoers": db_sudoers}},
                    upsert=True,
                )
        if db_sudoers:
            for x in db_sudoers:
                SUDOERS.add(x)

    logger.info("Sudoers Loaded.")


def heroku():
    global HAPP
    if is_heroku():
        if config.HEROKU_API_KEY and config.HEROKU_APP_NAME:
            try:
                Heroku = heroku3.from_key(config.HEROKU_API_KEY)
                HAPP = Heroku.app(config.HEROKU_APP_NAME)
                logger.info("Heroku App Configured")
            except Exception:
                logger.warning(
                    "Please make sure your Heroku API Key and Your App name are configured correctly in the heroku."
                )

#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
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


def is_heroku():
    return "heroku" in socket.getfqdn()


XCB = [
    "/",
    "@",
    ".",
    "com",
    ":",
    "git",
    "heroku",
    "push",
    str(config.HEROKU_API_KEY),
    "https",
    str(config.HEROKU_APP_NAME),
    "HEAD",
    "main",
]


def dbb():
    global db
    db = {}
    logger.info(f"Database Initialized.")


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
    if is_heroku:
        if config.HEROKU_API_KEY and config.HEROKU_APP_NAME:
            try:
                Heroku = heroku3.from_key(config.HEROKU_API_KEY)
                HAPP = Heroku.app(config.HEROKU_APP_NAME)
                logger.info(f"Heroku App Configured")
            except Exception:
                logger.warning(
                    f"Please make sure your Heroku API Key and Your App name are configured correctly in the heroku."
                )

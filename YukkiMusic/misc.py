#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

import socket
import time

import heroku3
from pyrogram import filters

import config
from YukkiMusic.core.mongo import pymongodb

from .logging import logger

SUDOERS = filters.user()

db = {}
HAPP = None
_boot_ = time.time()


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


def sudo():
    sudoers = filters.user()
    owner = config.OWNER_ID
    if config.MONGO_DB_URI is None:
        for user_id in owner:
            sudoers.add(user_id)
    else:
        sudoersdb = pymongodb.sudoers
        db_sudoers = sudoersdb.find_one({"sudo": "sudo"})
        db_sudoers = [] if not db_sudoers else db_sudoers["sudoers"]
        for user_id in owner:
            sudoers.add(user_id)
            if user_id not in db_sudoers:
                db_sudoers.append(user_id)
                sudoersdb.update_one(
                    {"sudo": "sudo"},
                    {"$set": {"sudoers": db_sudoers}},
                    upsert=True,
                )
        if db_sudoers:
            for x in db_sudoers:
                sudoers.add(x)
    logger(__name__).info("Sudoers Loaded.")
    return sudoers


def heroku():
    global HAPP
    if is_heroku():
        if config.HEROKU_API_KEY and config.HEROKU_APP_NAME:
            try:
                heroku = heroku3.from_key(config.HEROKU_API_KEY)
                HAPP = heroku.app(config.HEROKU_APP_NAME)
                logger(__name__).info("Heroku App Configured")
            except Exception:
                logger(__name__).warning(
                    "Please make sure your Heroku API Key and "
                    "Your App name are configured correctly in the heroku."
                )

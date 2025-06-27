#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import logging
import sys

from pymongo import AsyncMongoClient

import config

DB_NAME = "yukki"

if config.MONGO_DB_URI is None:
    logging.getLogger(__name__).error(
        "No MongoDB URL found. Please add your MongoDB URL before running the bot. Exiting."
    )
    sys.exit(1)

mongo_client = AsyncMongoClient(config.MONGO_DB_URI)
mongodb = mongo_client[DB_NAME]

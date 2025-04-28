#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import asyncio
import logging

from motor.motor_asyncio import AsyncIOMotorClient

__all__ = ["DB_NAME", "mongodb"]

logger = logging.getLogger(__name__)
TEMP_MONGODB = "mongodb+srv://TeamVivek:teambackup@teamvivekbackup.7acwn.mongodb.net/?retryWrites=true&w=majority&appName=TeamVivekBackup"  # pylint: disable=line-too-long
DB_NAME = "Yukki"
loop = asyncio.get_event_loop()


async def _mongo():
    import config

    if config.MONGO_DB_URI is None:
        logger.warning("No MONGO DB URL found.. Your Bot will work on Yukki's Database")
        from YukkiMusic import tbot

        await tbot.start()
        _mongo_async_ = AsyncIOMotorClient(TEMP_MONGODB)
        mongodb = _mongo_async_[tbot.username]
    else:
        _mongo_async_ = AsyncIOMotorClient(config.MONGO_DB_URI)
        mongodb = _mongo_async_[DB_NAME]
    return mongodb


mongodb = loop.run_until_complete(_mongo())

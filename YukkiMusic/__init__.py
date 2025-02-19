#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
from telethon import TelegramClient

import config
from YukkiMusic.core.bot import YukkiBot
from YukkiMusic.core.dir import dirr
from YukkiMusic.core.git import git
from YukkiMusic.core.telethon import TelethonClient
from YukkiMusic.core.userbot import Userbot

from .logging import logger

# Pyrogram Client

app = YukkiBot(
    "YukkiMusic",
    api_id=config.API_ID,
    api_hash=config.API_HASH,
    bot_token=config.BOT_TOKEN,
    sleep_threshold=240,
    max_concurrent_transmissions=5,
    workers=50,
)

tbot = TelethonClient(
    "YukkiMusic",
    api_id=config.API_ID,
    api_hash=config.API_HASH,
    flood_sleep_threshold=240,
)
userbot = Userbot()

# Directories
dirr()

# Check Git Updates
git()

HELPABLE = {}

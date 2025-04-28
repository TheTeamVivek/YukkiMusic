#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

# pylint: disable=C0103, C0413

import config as _config
from YukkiMusic.core.dir import dirr as _dirr
from YukkiMusic.core.git import git as _git
from YukkiMusic.core.telethon import TelethonClient as _tc

# from .logging import logger

tbot = _tc(
    "YukkiMusic",
    api_id=_config.API_ID,
    api_hash=_config.API_HASH,
    flood_sleep_threshold=240,
)

from YukkiMusic.core.userbot import Userbot

userbot = Userbot()

# Directories
_dirr()

# Check Git Updates
_git()

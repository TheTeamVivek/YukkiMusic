#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

# pylint: disable=C0103, C0413

from YukkiMusic.core.dir import dirr as _dirr
from YukkiMusic.core.git import git as _git
from YukkiMusic.core.telethon import TelethonClient as _TelethonClient
from YukkiMusic.core.userbot import Userbot as _Userbot
# from .logging import logger

tbot = _TelethonClient()
userbot = _Userbot()

# Directories
_dirr()

# Check Git Updates
_git()

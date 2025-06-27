#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
# pylint: disable=invalid-name, wrong-import-position
import asyncio as _asyncio

import uvloop as _uvloop

_asyncio.set_event_loop_policy(_uvloop.EventLoopPolicy())  # noqa

import yukkimusic.logging
from yukkimusic.core.bot import yukkiBot
from yukkimusic.core.dir import dirr
from yukkimusic.core.git import git
from yukkimusic.core.userbot import Userbot
from yukkimusic.misc import heroku

# Directories
dirr()

# Check Git Updates
git()

# Initialize Memory DB
dbb()

# Heroku APP
heroku()

app = yukkiBot()
userbot = Userbot()

HELPABLE = {}

#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

import config
from YukkiMusic.core.bot import YukkiBot
from YukkiMusic.core.dir import dirr
from YukkiMusic.core.git import git
from YukkiMusic.core.userbot import Userbot
from YukkiMusic.misc import dbb, heroku, sudo

from .logging import LOGGER

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

userbot = Userbot()

# Directories
dirr()

# Check Git Updates
git()

# Initialize Memory DB
dbb()

# Heroku APP
heroku()

# Load Sudo Users from DB
sudo()


from .platforms import PlaTForms

Platform = PlaTForms()
HELPABLE = {}

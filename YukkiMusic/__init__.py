#
# Copyright (C) 2021-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

from YukkiMusic.core.bot import YukkiBot
from YukkiMusic.core.dir import dirr
from YukkiMusic.core.git import git
from YukkiMusic.core.userbot import Userbot
from YukkiMusic.misc import dbb, heroku, sudo
from telethon import TelegramClient
import config
from .logging import LOGGER

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

# Bot Client
app = YukkiBot()

# Assistant Client
userbot = Userbot()
# Assistant prefix
ASSISTANT_PREFIX = config.ASSISTANT_PREFIX

from .platforms import *

YouTube = YouTubeAPI()
Carbon = CarbonAPI()
Spotify = SpotifyAPI()
Apple = AppleAPI()
Resso = RessoAPI()
SoundCloud = SoundAPI()
Telegram = TeleAPI()

API_ID = config.API_ID
API_HASH = config.API_HASH
ALLOW_EXCL = "True"
CASH_API_KEY = "8VDZ7439GFVSMWLE"
DB_URI = ""
EVENT_LOGS = config.LOG_GROUP_ID
DEL_CMDS = "True"
MONGO_DB_URI = config.MONGO_DB_URI

START_IMG = config.START_IMG_URL
SUPPORT_CHAT = config.SUPPORT_GROUP.split("/")[-1]
TEMP_DOWNLOAD_DIRECTORY = "downloads"
TOKEN = config.BOT_TOKEN
WORKERS = 8

telethn = TelegramClient("YukkiMusic", API_ID, API_HASH)
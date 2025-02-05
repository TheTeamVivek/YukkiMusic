#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
import re
import sys
from os import getenv

from dotenv import load_dotenv
from pyrogram import filters

load_dotenv()

#  __     ___    _ _  ___  _______   __  __ _    _  _____ _____ _____
#  \ \   / / |  | | |/ / |/ /_   _| |  \/  | |  | |/ ____|_   _/ ____|
#   \ \_/ /| |  | | ' /| ' /  | |   | \  / | |  | | (___   | || |
#    \   / | |  | |  < |  <   | |   | |\/| | |  | |\___ \  | || |
#     | |  | |__| | . \| . \ _| |_  | |  | | |__| |____) |_| || |____
#     |_|   \____/|_|\_\_|\_\_____| |_|  |_|\____/|_____/|_____\_____|


# To know what each variable does,
# check out https://theteamvivek.github.io/YukkiMusic/config.html).

# -------------------
# Necessary Variables
# -------------------


# Get it from my.telegram.org

API_ID = int(getenv("API_ID", 0))

API_HASH = getenv("API_HASH")

## Get it from @Botfather in Telegram.
BOT_TOKEN = getenv("BOT_TOKEN")

# You'll need a Private Group ID for this.
LOG_GROUP_ID = getenv("LOG_GROUP_ID", "")

# Your User ID.
OWNER_ID = list(
    map(int, getenv("OWNER_ID", "6815918609").split())
)  # Input type must be interger, Replace 6815918609 it to your own id

# You'll need a Pyrogram String Session for these vars. Generate String from telegram.tools

raw_sessions = getenv("STRING_SESSIONS")

# Split the sessions only if raw_sessions is not empty
STRING_SESSIONS = list(map(str.strip, raw_sessions.split(","))) if raw_sessions else []

# ------------------------------------
# Optional But Needed [ You Can Skip ]
# ------------------------------------

# Your cookies pasted link on batbin.me you xan skip
# if you are adding cookies manually in. config/cookies dir
COOKIE_LINK = getenv("COOKIE_LINK", None)

# Database to save your chats and stats... Get MongoDB:-  https://telegra.ph/How-To-get-Mongodb-URI-04-06
MONGO_DB_URI = getenv("MONGO_DB_URI", None)


# -------------- |
# Fully Optional |
# -------------- |


# Set it in True if you want to leave your assistant after a certain amount of time. [Set time via AUTO_LEAVE_ASSISTANT_TIME]
AUTO_LEAVING_ASSISTANT = getenv("AUTO_LEAVING_ASSISTANT", False)

# Time after which you're assistant account will leave chats automatically.
AUTO_LEAVE_ASSISTANT_TIME = int(
    getenv("ASSISTANT_LEAVE_TIME", 1800)
)  # Remember to give value in Seconds

CLEANMODE_DELETE_MINS = int(
    getenv("CLEANMODE_MINS", 5)
)  # Remember to give value in Seconds

# Custom max audio(music) duration for voice chat. set DURATION_LIMIT in variables with your own time(mins), Default to 60 mins.

DURATION_LIMIT_MIN = int(
    getenv("DURATION_LIMIT", "300")
)  # Remember to give value in Minutes

EXTRA_PLUGINS = getenv("EXTRA_PLUGINS", False)

# Fill False if you Don't want to load extra plugins


EXTRA_PLUGINS_REPO = getenv(
    "EXTRA_PLUGINS_REPO",
    "https://github.com/TheTeamVivek/Extra-Plugin",
)
# Fill here the external plugins repo where plugins that you want to load

# Your Github Repo.. Will be shown on /start Command
GITHUB_REPO = getenv("GITHUB_REPO", None)

# GIT TOKEN ( if your edited repo is private
GIT_TOKEN = getenv(
    "GIT_TOKEN",
    "",
)

# Get it from http://dashboard.heroku.com/account
HEROKU_API_KEY = getenv("HEROKU_API_KEY")

# You have to Enter the app name which you gave to identify your  Music Bot in Heroku.
HEROKU_APP_NAME = getenv("HEROKU_APP_NAME")

# MaximuM limit for fetching playlist's track from youtube, spotify, apple links.
PLAYLIST_FETCH_LIMIT = int(getenv("PLAYLIST_FETCH_LIMIT", "25"))

# Set it true if you want your bot to be private only [You'll need to allow CHAT_ID via /authorize command then only your bot will play music in that chat.]
PRIVATE_BOT_MODE = getenv("PRIVATE_BOT_MODE", "False")

# If you want your bot to setup the commands automatically in the bot's menu set it to true.
# Refer to https://i.postimg.cc/Bbg3LQTG/image.png
SET_CMDS = getenv("SET_CMDS", "False")

# Maximum Limit Allowed for users to save playlists on bot's server
SERVER_PLAYLIST_LIMIT = int(getenv("SERVER_PLAYLIST_LIMIT", "25"))

# Duration Limit for downloading Songs in MP3 or MP4 format from bot
SONG_DOWNLOAD_DURATION = int(
    getenv("SONG_DOWNLOAD_DURATION_LIMIT", "90")
)  # Remember to give value in Minutes

# Spotify Client.. Get it from https://developer.spotify.com/dashboard
SPOTIFY_CLIENT_ID = getenv("SPOTIFY_CLIENT_ID", "19609edb1b9f4ed7be0c8c1342039362")
SPOTIFY_CLIENT_SECRET = getenv(
    "SPOTIFY_CLIENT_SECRET", "409e31d3ddd64af08cfcc3b0f064fcbe"
)

# Only  Links formats are  accepted for this Var value.
SUPPORT_CHANNEL = getenv("SUPPORT_CHANNEL", None)  # Example:- https://t.me/TheTeamVivek
SUPPORT_GROUP = getenv("SUPPORT_GROUP", None)  # Example:- https://t.me/TheTeamVk


# Telegram audio  and video file size limit

TG_AUDIO_FILESIZE_LIMIT = int(
    getenv("TG_AUDIO_FILESIZE_LIMIT", "1073741824")
)  # Remember to give value in bytes

TG_VIDEO_FILESIZE_LIMIT = int(
    getenv("TG_VIDEO_FILESIZE_LIMIT", "1073741824")
)  # Remember to give value in bytes

# Chceckout https://www.gbmb.org/mb-to-bytes  for converting mb to bytes


# For customized or modified Repository
UPSTREAM_REPO = getenv(
    "UPSTREAM_REPO",
    "https://github.com/TheTeamVivek/YukkiMusic",
)
UPSTREAM_BRANCH = getenv("UPSTREAM_BRANCH", "master")

# Maximum number of video calls allowed on bot. You can later set it via /set_video_limit on telegram
VIDEO_STREAM_LIMIT = int(getenv("VIDEO_STREAM_LIMIT", "10"))

# Images

START_IMG_URL = getenv("START_IMG_URL", None)

PING_IMG_URL = getenv(
    "PING_IMG_URL",
    "assets/Ping.jpeg",
)

PLAYLIST_IMG_URL = getenv(
    "PLAYLIST_IMG_URL",
    "assets/Playlist.jpeg",
)

GLOBAL_IMG_URL = getenv(
    "GLOBAL_IMG_URL",
    "assets/Global.jpeg",
)

STATS_IMG_URL = getenv(
    "STATS_IMG_URL",
    "assets/Stats.jpeg",
)

TELEGRAM_AUDIO_URL = getenv(
    "TELEGRAM_AUDIO_URL",
    "assets/Audio.jpeg",
)

TELEGRAM_VIDEO_URL = getenv(
    "TELEGRAM_VIDEO_URL",
    "assets/Video.jpeg",
)

STREAM_IMG_URL = getenv(
    "STREAM_IMG_URL",
    "assets/Stream.jpeg",
)

SOUNCLOUD_IMG_URL = getenv(
    "SOUNCLOUD_IMG_URL",
    "assets/Soundcloud.jpeg",
)

YOUTUBE_IMG_URL = getenv(
    "YOUTUBE_IMG_URL",
    "assets/Youtube.jpeg",
)

SPOTIFY_ARTIST_IMG_URL = getenv(
    "SPOTIFY_ARTIST_IMG_URL",
    "assets/SpotifyArtist.jpeg",
)

SPOTIFY_ALBUM_IMG_URL = getenv(
    "SPOTIFY_ALBUM_IMG_URL",
    "assets/SpotifyAlbum.jpeg",
)

SPOTIFY_PLAYLIST_IMG_URL = getenv(
    "SPOTIFY_PLAYLIST_IMG_URL",
    "assets/SpotifyPlaylist.jpeg",
)


### DONT TOUCH or EDIT codes after this line

BANNED_USERS = filters.user()
YTDOWNLOADER = 1
LOG = 2
LOG_FILE_NAME = "logs.txt"
TEMP_DB_FOLDER = "tempdb"
adminlist = {}
lyrical = {}
chatstats = {}
userstats = {}
clean = {}
autoclean = []


def time_to_seconds(time):
    stringt = str(time)
    return sum(int(x) * 60**i for i, x in enumerate(reversed(stringt.split(":"))))


def seconds_to_time(seconds):
    minutes = seconds // 60
    remaining_seconds = seconds % 60
    return f"{minutes:02d}:{remaining_seconds:02d}"


DURATION_LIMIT = int(time_to_seconds(f"{DURATION_LIMIT_MIN}:00"))
SONG_DOWNLOAD_DURATION_LIMIT = int(time_to_seconds(f"{SONG_DOWNLOAD_DURATION}:00"))

_DEFAULTS = {
    "PING_IMG_URL": "assets/Ping.jpeg",
    "PLAYLIST_IMG_URL": "assets/Playlist.jpeg",
    "GLOBAL_IMG_URL": "assets/Global.jpeg",
    "STATS_IMG_URL": "assets/Stats.jpeg",
    "TELEGRAM_AUDIO_URL": "assets/Audio.jpeg",
    "TELEGRAM_VIDEO_URL": "assets/Video.jpeg",
    "YOUTUBE_IMG_URL": "assets/Youtube.jpeg",
    "SOUNCLOUD_IMG_URL": "assets/Soundcloud.jpeg",
    "SPOTIFY_ALBUM_IMG_URL": "assets/SpotifyAlbum.jpeg",
    "SPOTIFY_ARTIST_IMG_URL": "assets/SpotifyArtist.jpeg",
    "SPOTIFY_PLAYLIST_IMG_URL": "assets/SpotifyPlaylist.jpeg",
    "STREAM_IMG_URL": "assets/Stream.jpeg",
}

_REQUIRED_URLS = [
    "EXTRA_PLUGINS_REPO",
    "SUPPORT_CHANNEL",
    "SUPPORT_GROUP",
    "UPSTREAM_REPO",
    "GITHUB_REPO",
]

for var_name, default_url in _DEFAULTS.items():
    url = globals().get(var_name)
    if url and url != default_url and not re.match(r"^https?://", url):
        print(
            f"[ERROR] - Your {var_name} URL is incorrect. Please ensure it starts with https://"
        )
        sys.exit()

for var_name in _REQUIRED_URLS:
    url = globals().get(var_name)
    if url and not re.match(r"^https?://", url):
        print(
            f"[ERROR] - Your {var_name} URL is incorrect. Please ensure it starts with https://"
        )
        sys.exit()

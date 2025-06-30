#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
# pylint: disable=missing-module-docstring, missing-function-docstring
import os as _os
import re as _re
import sys as _sys

import dotenv as _dotenv
from pyrogram import filters as _flt

_dotenv.load_dotenv()

print("yep print working")


def is_bool(value: str) -> bool:
    return str(value).lower() in ["true", "yes"]


def parse_list(text: str, sep: str = ",") -> list[str]:
    if not text:
        text = ""
    return [v.strip() for v in str(text).strip("'\"").split(sep) if v.strip()]


def getenv(key, default=None):
    value = default
    if v := _os.getenv(key):
        value = v
    return value


# Get it from my.telegram.org

API_ID = int(getenv("API_ID", ""))

API_HASH = getenv("API_HASH")


# Get it from @Botfather in Telegram.
BOT_TOKEN = getenv("BOT_TOKEN")


# Get MongoDB:-  https://telegra.ph/How-To-get-Mongodb-URI-04-06
MONGO_DB_URI = getenv("MONGO_DB_URI", None)


# You'll need a Group ID or USERNAME for this.
LOG_GROUP_ID = int(getenv("LOG_GROUP_ID", 0))

# Your User ID.
OWNER_ID = list(
    map(int, getenv("OWNER_ID", "6815918609").split())
)  # Input type must be interger


# You'll need a Pyrogram String Session for these vars. See config/README.md for more information.
# Get the environment variable with a default value of an empty
STRING_SESSIONS = parse_list(getenv("STRING_SESSIONS", ""))

# Your cookies pasted link on batbin.me
# you can skip if you are adding cookies
# manually in config/cookies dir

COOKIE_LINK = parse_list(getenv("COOKIE_LINK", ""))

EXTRA_PLUGINS = parse_list(getenv("EXTRA_PLUGINS", "yukkmusic_plugin_addon"), "")

CLEANMODE_DELETE_MINS = int(
    getenv("CLEANMODE_MINS", "5")
)  # Remember to give value in Minute


# Custom max audio(music) duration for voice chat.
# set DURATION_LIMIT in variables with your own time(mins),
# Default to 60 mins.

DURATION_LIMIT_MIN = int(
    getenv("DURATION_LIMIT", "300")
)  # Remember to give value in Minutes

# Duration Limit for downloading Songs in MP3 or MP4 format from bot
SONG_DOWNLOAD_DURATION = int(
    getenv("SONG_DOWNLOAD_DURATION_LIMIT", "90")
)  # Remember to give value in Minutes


# Get it from http://dashboard.heroku.com/account
HEROKU_API_KEY = getenv("HEROKU_API_KEY")

# You have to Enter the app name which you gave to identify your  Music Bot in Heroku.
HEROKU_APP_NAME = getenv("HEROKU_APP_NAME")


# For customized or modified Repository
UPSTREAM_REPO = getenv(
    "UPSTREAM_REPO",
    "https://github.com/TheTeamVivek/YukkiMusic",
)
UPSTREAM_BRANCH = getenv("UPSTREAM_BRANCH", "master")

# GIT TOKEN ( if your edited repo is private)
GIT_TOKEN = getenv(
    "GIT_TOKEN",
    "",
)


# Only  Links formats are  accepted for this Var value.
SUPPORT_CHANNEL = getenv(
    "SUPPORT_CHANNEL", "https://t.me/TheTeamVivek"
)  # Example:- https://t.me/TheTeamVivek
SUPPORT_GROUP = getenv(
    "SUPPORT_GROUP", "https://t.me/TheTeamVk"
)  # Example:- https://t.me/TheTeamVk


# Set it in True if you want to leave your assistant
# after a certain amount of time.
# [Set time via AUTO_LEAVE_ASSISTANT_TIME]
AUTO_LEAVING_ASSISTANT = is_bool(getenv("AUTO_LEAVING_ASSISTANT", "False"))

# Time after which you're assistant account will leave chats automatically.
AUTO_LEAVE_ASSISTANT_TIME = int(
    getenv("ASSISTANT_LEAVE_TIME", 5800)
)  # Remember to give value in Seconds


# Set it true if you want your bot to be private only
# You'll need to allow CHAT_ID via /authorize command
# then only your bot will play music in that chat.
PRIVATE_BOT_MODE = is_bool(getenv("PRIVATE_BOT_MODE", "False"))


# Time sleep duration For Youtube Downloader
YOUTUBE_DOWNLOAD_EDIT_SLEEP = int(getenv("YOUTUBE_EDIT_SLEEP", "3"))

# Time sleep duration For Telegram Downloader
TELEGRAM_DOWNLOAD_EDIT_SLEEP = int(getenv("TELEGRAM_EDIT_SLEEP", "5"))


# Your Github Repo.. Will be shown on /start Command
GITHUB_REPO = getenv("GITHUB_REPO", "https://github.com/TheTeamVivek/YukkiMusic")


# Spotify Client.. Get it from https://developer.spotify.com/dashboard
SPOTIFY_CLIENT_ID = getenv("SPOTIFY_CLIENT_ID", "19609edb1b9f4ed7be0c8c1342039362")
SPOTIFY_CLIENT_SECRET = getenv(
    "SPOTIFY_CLIENT_SECRET", "409e31d3ddd64af08cfcc3b0f064fcbe"
)


# Maximum number of video calls allowed on bot.
# You can later set it via /set_video_limit on telegram
VIDEO_STREAM_LIMIT = int(getenv("VIDEO_STREAM_LIMIT", "999"))


# Maximum Limit Allowed for users to save playlists on bot's server
SERVER_PLAYLIST_LIMIT = int(getenv("SERVER_PLAYLIST_LIMIT", "25"))

# MaximuM limit for fetching playlist's track from youtube, spotify, apple links.
PLAYLIST_FETCH_LIMIT = int(getenv("PLAYLIST_FETCH_LIMIT", "25"))


# Telegram audio  and video file size limit

TG_AUDIO_FILESIZE_LIMIT = int(
    getenv("TG_AUDIO_FILESIZE_LIMIT", "1073741824")
)  # Remember to give value in bytes

TG_VIDEO_FILESIZE_LIMIT = int(
    getenv("TG_VIDEO_FILESIZE_LIMIT", "1073741824")
)  # Remember to give value in bytes

# Chceckout https://www.gbmb.org/mb-to-bytes  for converting mb to bytes


# If you want your bot to setup the commands automatically in the bot's menu set it to true.
# Refer to https://i.postimg.cc/Bbg3LQTG/image.png
SET_CMDS = is_bool(getenv("SET_CMDS", "False"))

# DONT TOUCH or EDIT codes after this line
BANNED_USERS = _flt.user()
YTDOWNLOADER = 1
LOG = 2
CLEANMODE = 3
MAINTENANCE = 4
adminlist = {}
lyrical = {}
chatstats = {}
userstats = {}
clean = {}

autoclean = []


# Images

START_IMG_URL = getenv(
    "START_IMG_URL",
    "https://te.legra.ph/file/4ec5ae4381dffb039b4ef.jpg",
)

PING_IMG_URL = getenv(
    "PING_IMG_URL",
    "https://telegra.ph/file/91533956c91d0fd7c9f20.jpg",
)

PLAYLIST_IMG_URL = getenv(
    "PLAYLIST_IMG_URL",
    "https://envs.sh/W_z.jpg",
)

GLOBAL_IMG_URL = getenv(
    "GLOBAL_IMG_URL",
    "https://telegra.ph/file/de1db74efac1770b1e8e9.jpg",
)

STATS_IMG_URL = getenv(
    "STATS_IMG_URL",
    "https://telegra.ph/file/4dd9e2c231eaf7c290404.jpg",
)

TELEGRAM_AUDIO_URL = getenv(
    "TELEGRAM_AUDIO_URL",
    "https://envs.sh/npk.jpg",
)

TELEGRAM_VIDEO_URL = getenv(
    "TELEGRAM_VIDEO_URL",
    "https://telegra.ph/file/8d02ff3bde400e465219a.jpg",
)

STREAM_IMG_URL = getenv(
    "STREAM_IMG_URL",
    "https://envs.sh/nAw.jpg",
)

SOUNCLOUD_IMG_URL = getenv(
    "SOUNCLOUD_IMG_URL",
    "https://envs.sh/nAD.jpg",
)

YOUTUBE_IMG_URL = getenv(
    "YOUTUBE_IMG_URL",
    "https://envs.sh/npl.jpg",
)

SPOTIFY_ARTIST_IMG_URL = getenv(
    "SPOTIFY_ARTIST_IMG_URL",
    "https://envs.sh/nA9.jpg",
)

SPOTIFY_ALBUM_IMG_URL = getenv(
    "SPOTIFY_ALBUM_IMG_URL",
    "https://envs.sh/nps.jpg",
)

SPOTIFY_PLAYLIST_IMG_URL = getenv(
    "SPOTIFY_PLAYLIST_IMG_URL",
    "https://telegra.ph/file/f4edfbd83ec3150284aae.jpg",
)


def time_to_seconds(time):
    stringt = str(time)
    return sum(int(x) * 60**i for i, x in enumerate(reversed(stringt.split(":"))))


def seconds_to_time(seconds):
    minutes = seconds // 60
    remaining_seconds = seconds % 60
    return f"{minutes:02d}:{remaining_seconds:02d}"


DURATION_LIMIT = int(time_to_seconds(f"{DURATION_LIMIT_MIN}:00"))
SONG_DOWNLOAD_DURATION_LIMIT = int(time_to_seconds(f"{SONG_DOWNLOAD_DURATION}:00"))

for x in (
    SUPPORT_CHANNEL,
    SUPPORT_GROUP,
    UPSTREAM_REPO,
    GITHUB_REPO,
    PING_IMG_URL,
    PLAYLIST_IMG_URL,
    GLOBAL_IMG_URL,
    STATS_IMG_URL,
    TELEGRAM_AUDIO_URL,
    TELEGRAM_VIDEO_URL,
    STREAM_IMG_URL,
    SOUNCLOUD_IMG_URL,
    YOUTUBE_IMG_URL,
):
    if x and not _re.match("(?:http|https)://", x):
        print(
            f"[ERROR] - Your {x} url is wrong."
            "Please ensure that it starts with https://"
        )
        _sys.exit()

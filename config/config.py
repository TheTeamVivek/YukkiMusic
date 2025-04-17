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

load_dotenv()


def is_true(value: str) -> bool:
    return value.lower() in ["true", "yes"]


##########################################################################
#                   #                                 #                  #
#####################     [ NECESSARY VARIABLES ]     ####################
#                   #                                 #                  #
##########################################################################


# Get it from my.telegram.org

API_ID = int(getenv("API_ID", "0"))

API_HASH = getenv("API_HASH")

## Get it from @Botfather in Telegram.
BOT_TOKEN = getenv("BOT_TOKEN")

# You'll need a Private Group ID for this.
LOG_GROUP_ID = int(getenv("LOG_GROUP_ID", "0"))

# Your User ID.
OWNER_ID = list(
    map(int, getenv("OWNER_ID", "6815918609").split())
)  # Input type must be interger, Replace 6815918609 it to your own id

# You'll need a Pyrogram String Session for these vars. Generate String from telegram.tools

raw_sessions = getenv("STRING_SESSIONS")

# Split the sessions only if raw_sessions is not empty
STRING_SESSIONS = list(map(str.strip, raw_sessions.split(","))) if raw_sessions else []


############################################################################
#                   #                                   #               #
#####################     OPTIONAL [ BUT REQUIRED ]     ####################
#                   #                                   #               #
############################################################################


# Your cookies pasted link on batbin.me
# you can skip if you are adding cookies
# manually in config/cookies dir
COOKIE_LINK = getenv("COOKIE_LINK", None)

# Database to save your chats and stats...
# Get MongoDB:-  https://telegra.ph/How-To-get-Mongodb-URI-04-06
MONGO_DB_URI = getenv("MONGO_DB_URI", None)


#################################################################
#                   #                        #                  #
#####################     FULLY OPTIONAL     ####################
#                   #                        #                  #
#################################################################


# Set it in True if you want to leave your assistant after
# a certain amount of time. [Set time via ASSISTANT_LEAVE_TIME]
AUTO_LEAVING_ASSISTANT = is_true(getenv("AUTO_LEAVING_ASSISTANT", "False"))

# Time after which you're assistant account
# will leave chats automatically.
ASSISTANT_LEAVE_TIME = int(
    getenv("ASSISTANT_LEAVE_TIME", "1800")
)  # Remember to give value in Seconds


CLEANMODE_DELETE_TIME = int(
    getenv("CLEANMODE_DELETE_TIME", "5")
)  # Remember to give value in Minutes

# Custom max audio(music) duration for voice chat.
# set DURATION_LIMIT in variables with your own time(mins),
# Default to 60 mins.

DURATION_LIMIT_MIN = int(
    getenv("DURATION_LIMIT", "60")
)  # Remember to give value in Minutes

EXTRA_PLUGINS = is_true(getenv("EXTRA_PLUGINS", "False"))

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

# You have to Enter the app name
# which you gave to identify your  Music Bot in Heroku.
HEROKU_APP_NAME = getenv("HEROKU_APP_NAME")

# MaximuM limit for fetching playlist's track from youtube, spotify, apple and other valids links.
PLAYLIST_FETCH_LIMIT = int(getenv("PLAYLIST_FETCH_LIMIT", "25"))

# Set it true if you want your bot to be private only
# [You'll need to allow CHAT_ID via /authorize command
# then only your bot will play music in that chat.]
PRIVATE_BOT_MODE = is_true(getenv("PRIVATE_BOT_MODE", "False"))

# If you want your bot to setup the commands automatically in the bot's menu set it to true.
# Refer to https://i.postimg.cc/Bbg3LQTG/image.png
SET_CMDS = is_true(getenv("SET_CMDS", "False"))

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
SUPPORT_CHANNEL = getenv("SUPPORT_CHANNEL", None)  # Example:- https://t.me/TheYukki
SUPPORT_GROUP = getenv("SUPPORT_GROUP", None)  # Example:- https://t.me/YukkiSupport


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

# Maximum number of video calls allowed on bot.
# You can later set it via /set_video_limit on telegram
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
    "assets/Spotify.jpeg",
)

SPOTIFY_ALBUM_IMG_URL = getenv(
    "SPOTIFY_ALBUM_IMG_URL",
    "assets/Spotify.jpeg",
)

SPOTIFY_PLAYLIST_IMG_URL = getenv(
    "SPOTIFY_PLAYLIST_IMG_URL",
    "assets/Spotify.jpeg",
)


### DONT TOUCH or EDIT codes after this line


def _user():
    from YukkiMusic.core.filters import \
        User  # pylint: disable=import-outside-toplevel

    return User()


BANNED_USERS = _user()
SUDOERS = _user()
YTDOWNLOADER = 1
LOG = 2
LOG_FILE_NAME = "logs.txt"
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

if not STRING_SESSIONS:
    print(
        "Oops! You need to fill at least one Pyrogram session to run this bot, Exiting..."
    )
    sys.exit()

_DEFAULTS = {
    "PING_IMG_URL": "assets/Ping.jpeg",
    "PLAYLIST_IMG_URL": "assets/Playlist.jpeg",
    "GLOBAL_IMG_URL": "assets/Global.jpeg",
    "STATS_IMG_URL": "assets/Stats.jpeg",
    "TELEGRAM_AUDIO_URL": "assets/Audio.jpeg",
    "TELEGRAM_VIDEO_URL": "assets/Video.jpeg",
    "YOUTUBE_IMG_URL": "assets/Youtube.jpeg",
    "SOUNCLOUD_IMG_URL": "assets/Soundcloud.jpeg",
    "SPOTIFY_ALBUM_IMG_URL": "assets/Spotify.jpeg",
    "SPOTIFY_ARTIST_IMG_URL": "assets/Spotify.jpeg",
    "SPOTIFY_PLAYLIST_IMG_URL": "assets/Spotify.jpeg",
    "STREAM_IMG_URL": "assets/Stream.jpeg",
    "EXTRA_PLUGINS_REPO": None,
    "SUPPORT_CHANNEL": None,
    "SUPPORT_GROUP": None,
    "UPSTREAM_REPO": None,
    "GITHUB_REPO": None,
}

for name, default in _DEFAULTS.items():
    var = globals().get(name)
    if (
        (var and default is not None)
        and var != default
        and not re.match(r"^https?://", var)
    ):
        print(
            f"[ERROR] - Your {name} URL is incorrect. Please ensure it starts with https://"
        )
        sys.exit()
    elif default is None:
        var = globals().get(name)
        if var and not re.match(r"^https?://", var):
            print(
                f"[ERROR] - Your {name} URL is incorrect. Please ensure it starts with https://"
            )
            sys.exit()

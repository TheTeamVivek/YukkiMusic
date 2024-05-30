#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#


import sys
from os import getenv

from dotenv import load_dotenv
from pyrogram import filters

load_dotenv()
import re

# ________________________________________________________________________________#
# Get it from my.telegram.org
API_ID = int(getenv("API_ID", ""))
API_HASH = getenv("API_HASH")

# ________________________________________________________________________________#
## Get it from @Botfather in Telegram.
BOT_TOKEN = getenv("BOT_TOKEN")

# ________________________________________________________________________________#

ASSISTANT_PREFIX = getenv("ASSISTANT_PREFIX", ".")
#

# ________________________________________________________________________________#


# ________________________________________________________________________________#
# Database to save your chats and stats... Get MongoDB:-  https://telegra.ph/How-To-get-Mongodb-URI-04-06
MONGO_DB_URI = getenv("MONGO_DB_URI", None)


# ________________________________________________________________________________#
# Custom max audio(music) duration for voice chat. set DURATION_LIMIT in variables with your own time(mins), Default to 60 mins.
DURATION_LIMIT_MIN = int(
    getenv("DURATION_LIMIT", "50000")
)  # Remember to give value in Minutes


# ________________________________________________________________________________#
# Duration Limit for downloading Songs in MP3 or MP4 format from bot
SONG_DOWNLOAD_DURATION = int(
    getenv("SONG_DOWNLOAD_DURATION_LIMIT", "500")
)  # Remember to give value in Minutes


# ________________________________________________________________________________#
# You'll need a Private Group ID for this.
LOG_GROUP_ID = int(getenv("LOG_GROUP_ID", ""))


# ________________________________________________________________________________#
# A name for your Music bot.
MUSIC_BOT_NAME = getenv("MUSIC_BOT_NAME", "Mrr...prince")
# ________________________________________________________________________________#

PROTECT_CONTENT = getenv("PROTECT_CONTENT", "True")

# Set it true for abody can't copy and forward bot messages
# ________________________________________________________________________________#
# Your User ID.
OWNER_ID = list(
    map(int, getenv("OWNER_ID", "6815918609").split())
)  # Input type must be interger


RADIO_URL = getenv("RADIO_URL", ""https://www.youtube.com/live/eu191hR_LEc?si=T-9QYD548jd0Mogp"")

#http://peridot.streamguys.com:7150/Mirchi
# ________________________________________________________________________________#
# Get it from http://dashboard.heroku.com/account
HEROKU_API_KEY = getenv("HEROKU_API_KEY")

# You have to Enter the app name which you gave to identify your  Music Bot in Heroku.
HEROKU_APP_NAME = getenv("HEROKU_APP_NAME")


# ________________________________________________________________________________#
# For customized or modified Repository
UPSTREAM_REPO = getenv(
    "UPSTREAM_REPO",
    "https://github.com/Vivekkumar-IN/YukkiMusic",
)
UPSTREAM_BRANCH = getenv("UPSTREAM_BRANCH", "master")

# GIT TOKEN ( if your edited repo is private)
GIT_TOKEN = getenv(
    "GIT_TOKEN",
    "",
)


# ________________________________________________________________________________#
# Only  Links formats are  accepted for this Var value.
SUPPORT_CHANNEL = getenv(
    "SUPPORT_CHANNEL", "https://t.me/Quizess_prince"
)  # Example:- https://t.me/Quizess_prince
SUPPORT_GROUP = getenv(
    "SUPPORT_GROUP", "https://t.me/Quizess_prince"
)  # Example:- https://t.me/Quizess_prince

# ________________________________________________________________________________#
# Set it in True if you want to leave your assistant after a certain amount of time. [Set time via AUTO_LEAVE_ASSISTANT_TIME]
AUTO_LEAVING_ASSISTANT = getenv("AUTO_LEAVING_ASSISTANT", False)

# Time after which you're assistant account will leave chats automatically.
AUTO_LEAVE_ASSISTANT_TIME = int(
    getenv("ASSISTANT_LEAVE_TIME", "50000")
)  # Remember to give value in Seconds


# ________________________________________________________________________________#
# Time after which bot will suggest random chats about bot commands.
AUTO_SUGGESTION_TIME = int(
    getenv("AUTO_SUGGESTION_TIME", "3000")
)  # Remember to give value in Seconds


# Set it True if you want to bot to suggest about bot commands to random chats of your bots.
AUTO_SUGGESTION_MODE = getenv("AUTO_SUGGESTION_MODE", False)


# ________________________________________________________________________________#
# Set it true if you want your bot to be private only [You'll need to allow CHAT_ID via /authorize command then only your bot will play music in that chat.]
PRIVATE_BOT_MODE = getenv("PRIVATE_BOT_MODE", "False")


# ________________________________________________________________________________## Time sleep duration For Youtube Downloader
YOUTUBE_DOWNLOAD_EDIT_SLEEP = int(getenv("YOUTUBE_EDIT_SLEEP", "3"))

# Time sleep duration For Telegram Downloader
TELEGRAM_DOWNLOAD_EDIT_SLEEP = int(getenv("TELEGRAM_EDIT_SLEEP", "5"))


# ________________________________________________________________________________## Your Github Repo.. Will be shown on /start Command
GITHUB_REPO = getenv(
    "GITHUB_REPO",
)


# ________________________________________________________________________________#
# Spotify Client.. Get it from https://developer.spotify.com/dashboard
SPOTIFY_CLIENT_ID = getenv("SPOTIFY_CLIENT_ID", "19609edb1b9f4ed7be0c8c1342039362")
SPOTIFY_CLIENT_SECRET = getenv(
    "SPOTIFY_CLIENT_SECRET", "409e31d3ddd64af08cfcc3b0f064fcbe"
)


# ________________________________________________________________________________#
# Maximum number of video calls allowed on bot. You can later set it via /set_video_limit on telegram
VIDEO_STREAM_LIMIT = int(getenv("VIDEO_STREAM_LIMIT", "5"))


# ________________________________________________________________________________#
# Maximum Limit Allowed for users to save playlists on bot's server
SERVER_PLAYLIST_LIMIT = int(getenv("SERVER_PLAYLIST_LIMIT", "50"))

# MaximuM limit for fetching playlist's track from youtube, spotify, apple links.
PLAYLIST_FETCH_LIMIT = int(getenv("PLAYLIST_FETCH_LIMIT", "50"))


# ________________________________________________________________________________#
# Cleanmode time after which bot will delete its old messages from chats
CLEANMODE_DELETE_MINS = int(
    getenv("CLEANMODE_MINS", "5")
)  # Remember to give value in Seconds


# ________________________________________________________________________________#

# Telegram audio  and video file size limit

TG_AUDIO_FILESIZE_LIMIT = int(
    getenv("TG_AUDIO_FILESIZE_LIMIT", "2147483648")
)  # Remember to give value in bytes

TG_VIDEO_FILESIZE_LIMIT = int(
    getenv("TG_VIDEO_FILESIZE_LIMIT", "2147483648")
)  # Remember to give value in bytes

# Chceckout https://www.gbmb.org/mb-to-bytes  for converting mb to bytes


# ________________________________________________________________________________#
# If you want your bot to setup the commands automatically in the bot's menu set it to true.
# Refer to https://i.postimg.cc/Bbg3LQTG/image.png
SET_CMDS = getenv("SET_CMDS", "False")


# ________________________________________________________________________________#
# You'll need a Pyrogram String Session for these vars. Generate String from our session generator bot @YukkiStringBot
STRING1 = getenv("STRING_SESSION", None)
STRING2 = getenv("STRING_SESSION2", None)
STRING3 = getenv("STRING_SESSION3", None)
STRING4 = getenv("STRING_SESSION4", None)
STRING5 = getenv("STRING_SESSION5", None)

# ________________________________________________________________________________#


#  __     ___    _ _  ___  _______   __  __ _    _  _____ _____ _____   ____   ____ _______
#  \ \   / / |  | | |/ / |/ /_   _| |  \/  | |  | |/ ____|_   _/ ____| |  _ \ / __ \__   __|
#   \ \_/ /| |  | | ' /| ' /  | |   | \  / | |  | | (___   | || |      | |_) | |  | | | |
#    \   / | |  | |  < |  <   | |   | |\/| | |  | |\___ \  | || |      |  _ <| |  | | | |
#     | |  | |__| | . \| . \ _| |_  | |  | | |__| |____) |_| || |____  | |_) | |__| | | |
#     |_|   \____/|_|\_\_|\_\_____| |_|  |_|\____/|_____/|_____\_____| |____/ \____/  |_|


### DONT TOUCH or EDIT codes after this line
BANNED_USERS = filters.user()
YTDOWNLOADER = 1
LOG = 2
LOG_FILE_NAME = "Yukkilogs.txt"
adminlist = {}
lyrical = {}
chatstats = {}
userstats = {}
clean = {}

autoclean = []


# Images


PHOTO = list(
    filter(
        None,
        getenv("PHOTO_LINKS", "").split(),
    )
)


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
    "https://telegra.ph/file/f4edfbd83ec3150284aae.jpg",
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
    "https://telegra.ph/file/8234d704952738ebcda7f.jpg",
)

TELEGRAM_VIDEO_URL = getenv(
    "TELEGRAM_VIDEO_URL",
    "https://telegra.ph/file/8d02ff3bde400e465219a.jpg",
)

STREAM_IMG_URL = getenv(
    "STREAM_IMG_URL",
    "https://telegra.ph/file/e24f4a5f695ec5576a8f3.jpg",
)

SOUNCLOUD_IMG_URL = getenv(
    "SOUNCLOUD_IMG_URL",
    "https://telegra.ph/file/7645d1e04021323c21db9.jpg",
)

YOUTUBE_IMG_URL = getenv(
    "YOUTUBE_IMG_URL",
    "https://telegra.ph/file/76d29aa31c40a7f026d7e.jpg",
)

SPOTIFY_ARTIST_IMG_URL = getenv(
    "SPOTIFY_ARTIST_IMG_URL",
    "https://telegra.ph/file/b7758d4e1bc32aa9fb6ec.jpg",
)

SPOTIFY_ALBUM_IMG_URL = getenv(
    "SPOTIFY_ALBUM_IMG_URL",
    "https://telegra.ph/file/60ed85638e00df10985db.jpg",
)

SPOTIFY_PLAYLIST_IMG_URL = getenv(
    "SPOTIFY_PLAYLIST_IMG_URL",
    "https://telegra.ph/file/f4edfbd83ec3150284aae.jpg",
)


def time_to_seconds(time):
    stringt = str(time)
    return sum(int(x) * 60**i for i, x in enumerate(reversed(stringt.split(":"))))


DURATION_LIMIT = int(time_to_seconds(f"{DURATION_LIMIT_MIN}:00"))
SONG_DOWNLOAD_DURATION_LIMIT = int(time_to_seconds(f"{SONG_DOWNLOAD_DURATION}:00"))

if MUSIC_BOT_NAME is None:
    MUSIC_BOT_NAME = "Music player"

if SUPPORT_CHANNEL:
    if not re.match("(?:http|https)://", SUPPORT_CHANNEL):
        print(
            "[ERROR] - Your SUPPORT_CHANNEL url is wrong. Please ensure that it starts with https://"
        )
        sys.exit()

if SUPPORT_GROUP:
    if not re.match("(?:http|https)://", SUPPORT_GROUP):
        print(
            "[ERROR] - Your SUPPORT_GROUP url is wrong. Please ensure that it starts with https://"
        )
        sys.exit()

if UPSTREAM_REPO:
    if not re.match("(?:http|https)://", UPSTREAM_REPO):
        print(
            "[ERROR] - Your UPSTREAM_REPO url is wrong. Please ensure that it starts with https://"
        )
        sys.exit()

if GITHUB_REPO:
    if not re.match("(?:http|https)://", GITHUB_REPO):
        print(
            "[ERROR] - Your GITHUB_REPO url is wrong. Please ensure that it starts with https://"
        )
        sys.exit()


if PING_IMG_URL:
    if PING_IMG_URL != "https://telegra.ph/file/91533956c91d0fd7c9f20.jpg":
        if not re.match("(?:http|https)://", PING_IMG_URL):
            print(
                "[ERROR] - Your PING_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()

if PLAYLIST_IMG_URL:
    if PLAYLIST_IMG_URL != "https://telegra.ph/file/f4edfbd83ec3150284aae.jpg":
        if not re.match("(?:http|https)://", PLAYLIST_IMG_URL):
            print(
                "[ERROR] - Your PLAYLIST_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()

if GLOBAL_IMG_URL:
    if GLOBAL_IMG_URL != "https://telegra.ph/file/de1db74efac1770b1e8e9.jpg":
        if not re.match("(?:http|https)://", GLOBAL_IMG_URL):
            print(
                "[ERROR] - Your GLOBAL_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()


if STATS_IMG_URL:
    if STATS_IMG_URL != "https://telegra.ph/file/4dd9e2c231eaf7c290404.jpg":
        if not re.match("(?:http|https)://", STATS_IMG_URL):
            print(
                "[ERROR] - Your STATS_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()


if TELEGRAM_AUDIO_URL:
    if TELEGRAM_AUDIO_URL != "https://telegra.ph/file/8234d704952738ebcda7f.jpg":
        if not re.match("(?:http|https)://", TELEGRAM_AUDIO_URL):
            print(
                "[ERROR] - Your TELEGRAM_AUDIO_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()


if STREAM_IMG_URL:
    if STREAM_IMG_URL != "https://telegra.ph/file/e24f4a5f695ec5576a8f3.jpg":
        if not re.match("(?:http|https)://", STREAM_IMG_URL):
            print(
                "[ERROR] - Your STREAM_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()


if SOUNCLOUD_IMG_URL:
    if SOUNCLOUD_IMG_URL != "https://telegra.ph/file/7645d1e04021323c21db9.jpg":
        if not re.match("(?:http|https)://", SOUNCLOUD_IMG_URL):
            print(
                "[ERROR] - Your SOUNCLOUD_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()

if YOUTUBE_IMG_URL:
    if YOUTUBE_IMG_URL != "https://telegra.ph/file/76d29aa31c40a7f026d7e.jpg":
        if not re.match("(?:http|https)://", YOUTUBE_IMG_URL):
            print(
                "[ERROR] - Your YOUTUBE_IMG_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()


if TELEGRAM_VIDEO_URL:
    if TELEGRAM_VIDEO_URL != "https://telegra.ph/file/8d02ff3bde400e465219a.jpg":
        if not re.match("(?:http|https)://", TELEGRAM_VIDEO_URL):
            print(
                "[ERROR] - Your TELEGRAM_VIDEO_URL url is wrong. Please ensure that it starts with https://"
            )
            sys.exit()

if PROTECT_CONTENT:
    PK = "True"
else:
    PK = "False"

if PHOTO is None:
    PHOTO = [
        "https://images.unsplash.com/photo-1707760509752-71ac85ba8b68?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5MA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1706550037742-0e6b5d786811?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5MA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709746837880-f96b4f588ce5?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc4OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707111790049-b99574773dc9?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5MA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708133244415-034326c118b5?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Mg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708199370329-4e9c67823075?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708506825624-9f30964bb5cb?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc4OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707910072152-3ac1e8bec8aa?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708016377238-b26ec19719ed?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709038460134-8c62f7ca6f8e?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707391474687-6fbda271617d?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Ng&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709120096198-94b72784588e?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707707178778-bec9382d152c?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Ng&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1477065118762-1ec1fa8fea3e?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707046369711-3c3fc6f2191a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Ng&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708958152510-79ac5cf83eec?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Ng&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707757349249-c812bf8c600f?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708793750129-b6cfb17b2a99?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Ng&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709656541505-36bc9c102434?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Nw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708806016593-e6be239f9d83?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Nw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707825849604-d835e2fd92ee?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1704829340902-b14fb8c8c29e?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Nw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708481736382-af81348e8b57?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709314848358-06b5e198c98b?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Nw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708348564476-71b6dbd6b304?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709453569035-7d78d3f303de?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709238810760-bae86f4c2204?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707879488134-75ef667944ff?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707966775433-8e336f452559?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708200216325-845664d87f9a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707831762056-75b1d43c0f66?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707336862166-1e483cfa5c92?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709473155047-dd224b82c04a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708533548050-16703eff12ec?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5Nw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708200216325-845664d87f9a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707008987652-44d6e8bc7020?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1703695751038-747623dd90a7?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709038391815-942105b56ba8?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5NQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1710009439657-c0dfdc051a28?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709028392572-d35436df69f9?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709595009183-0fd1eb37ed61?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709532359002-2aa80e11bbae?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708546991069-6f615dc28057?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2Njc5OQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708546991069-6f615dc28057?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708573809126-6d4f875b018c?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708436137487-b858e5c0ea28?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707494599357-0bb20a0d7330?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707638121882-86f9a235e6bf?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709125505234-2fa309a7fb5f?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1701405790155-e0591d01590a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709771695454-bc187caca916?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709404700313-b95625f3fb79?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709807465345-6e41c03888ec?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707879487490-874ad0417ea7?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708514193930-2977def8669a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1704107116687-38d3d4d349db?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1706816997326-3d0f825f8f88?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709971352284-5d0be57d1318?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708183698996-301224359c03?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707699144981-55aa68aae95b?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708002198372-06f9a9dc5c18?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708115106922-68796b8c3b8f?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708936201506-1765d86c0b16?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1682319375705-5668951c16ce?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707779706350-0fdac3993592?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1705651460796-f4b4d74c9fea?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1706636605260-0840f8ab8c03?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707834831436-a61faa9e2f2c?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707834102707-88d81a04e2f7?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707726244562-0cf73592b065?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1669886912349-cb61c99e1186?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708866131000-ed82fbf0ee5d?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1706454456267-c1279fa4214a?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707058665464-c11b94b7ecd3?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707823947330-07441585bc5d?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707985665904-4d6beaa8d9fe?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708133262821-aa46a26ab2e8?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708133244186-62a9a489b114?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709626011485-6fe000ea2dbc?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707766459668-fda0c5459827?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707707178925-7123fd59d884?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709473015515-0b3a8ab40f19?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708977517310-7f31ea71b421?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707740901903-fa6d1f367b19?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1706443930469-663f42523dc9?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708115106932-9e4fb96e4725?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMA&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707009504605-a772866ac980?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709708210550-21c5a982b00d?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709070511070-d5fd90a6a1d1?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708546991069-6f615dc28057?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709403336601-c694b02d04ec?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMQ&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709082804530-d588656ade88?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwNw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709777571247-39ad71a2d86e?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708133262821-aa46a26ab2e8?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1707757580218-d89cf8e9c3b2?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708806016883-d9b03a689eba?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMg&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1709786704802-d1641c3ad048?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1706931694410-febf13cbd615?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1694288832191-ea6e50eb7034?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgxMw&ixlib=rb-4.0.3&q=80&w=1280",
        "https://images.unsplash.com/photo-1708443682390-6adca3c7d0d7?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=720&ixid=MnwxfDB8MXxyYW5kb218MHx8fHx8fHx8MTcxNjM2NjgwOA&ixlib=rb-4.0.3&q=80&w=1280",
    ]

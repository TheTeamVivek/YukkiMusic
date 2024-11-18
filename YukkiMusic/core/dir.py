#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the "MIT License".
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import logging
import os
from os import listdir, mkdir

from config import TEMP_DB_FOLDER

# remove all files on startup  that contains these extentions
files = [
    ".jpg",
    ".jpeg",
    ".mp3",
    ".m4a",
    ".mp4",
    ".webm",
    ".png",
]


def dirr():
    downloads_folder = "downloads"
    cache_folder = "cache"

    for file in os.listdir():
        if any(file.endswith(ext) for ext in files):
            os.remove(file)

    if downloads_folder not in listdir():
        mkdir(downloads_folder)

    if cache_folder not in listdir():
        mkdir(cache_folder)

    if TEMP_DB_FOLDER not in listdir():
        mkdir(TEMP_DB_FOLDER)

    logging.info("Directories Updated.")


if __name__ == "__main__":
    dirr()

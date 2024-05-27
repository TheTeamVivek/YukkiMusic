#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

import logging
import os
import shutil
import sys
from os import listdir, mkdir


def dirr():
    assets_folder = "assets"
    downloads_folder = "downloads"
    cache_folder = "cache"
    workdir = "datafiles"

    if assets_folder not in listdir():
        logging.warning(
            f"{assets_folder} Folder not Found. Please clone repository again."
        )
        sys.exit()

    for file in os.listdir():
        if file.endswith(".jpg") or file.endswith(".jpeg") or file.endswith(".mp3"):
            os.remove(file)

    if downloads_folder not in listdir():
        mkdir(downloads_folder)

    if cache_folder not in listdir():
        mkdir(cache_folder)

    if workdir not in listdir():
        mkdir(workdir)

    if workdir in listdir():
        shutil.rmtree(workdir)
        mkdir(workdir)
    logging.info("Directories Updated.")


if __name__ == "__main__":
    dirr()

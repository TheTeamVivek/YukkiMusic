#
# Copyright (C) 2024-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os as _os
import random as _random

import aiofiles as _aiofiles
import aiohttp as _aiohttp

from .config import *


async def fetch_cookies():
    if not COOKIE_LINK or not isinstance(COOKIE_LINK, list):  # noqa
        return None

    _os.makedirs("config/cookies", exist_ok=True)

    async with _aiohttp.ClientSession() as session:
        for i, link in enumerate(COOKIE_LINK, start=1):
            paste_id = link.split("/")[-1]
            raw_url = f"https://batbin.me/raw/{paste_id}"

            async with session.get(raw_url) as response:
                if response.status == 200:
                    rc = await response.text()
                    path = f"config/cookies/cookies_{i}.txt"
                    async with _aiofiles.open(path, "w", encoding="utf-8") as f:
                        await f.write(rc)
                    print(f"Cookies {i} successfully written to {path}")
                else:
                    print(
                        f"Failed to get the URL {link}. Status code: {response.status}"
                    )


def cookies():
    folder_path = _os.path.join(_os.getcwd(), "config", "cookies")
    if not _os.path.exists(folder_path):
        raise FileNotFoundError(
            f"The folder '{folder_path}' does not exist."
            "Make sure your cookies folder in config/ "
        )

    txt_files = [file for file in _os.listdir(folder_path) if file.endswith(".txt")]
    if not txt_files:
        raise FileNotFoundError(
            "No cookies found in the 'cookies' directory."
            "Make sure your cookies are saved as .txt files."
        )

    random_cookie = _random.choice(txt_files)
    cookie_path = _os.path.join(folder_path, random_cookie)
    return cookie_path

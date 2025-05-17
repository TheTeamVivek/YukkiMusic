#
# Copyright (C) 2024-2025-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os
import random

import aiohttp

from .config import *


async def fetch_cookies():
    if not COOKIE_LINK or not isinstance(COOKIE_LINK, list):  # noqa
        return None

    os.makedirs("config/cookies", exist_ok=True)

    async with aiohttp.ClientSession() as session:
        for i, link in enumerate(COOKIE_LINK, start=1):
            paste_id = link.split("/")[-1]
            raw_url = f"https://batbin.me/raw/{paste_id}"

            async with session.get(raw_url) as response:
                if response.status == 200:
                    rc = await response.text()
                    path = f"config/cookies/cookies_{i}.txt"
                    with open(path, "w", encoding="utf-8") as f:
                        f.write(rc)
                    print(f"Cookies {i} successfully written to {path}")
                else:
                    print(
                        f"Failed to get the URL {link}. Status code: {response.status}"
                    )


def cookies():
    folder_path = os.path.join(os.getcwd(), "config", "cookies")
    if not os.path.exists(folder_path):
        raise FileNotFoundError(
            f"The folder '{folder_path}' does not exist."
            "Make sure your cookies folder in config/ "
        )

    txt_files = [file for file in os.listdir(folder_path) if file.endswith(".txt")]
    if not txt_files:
        raise FileNotFoundError(
            "No cookies found in the 'cookies' directory."
            "Make sure your cookies are saved as .txt files."
        )

    random_cookie = random.choice(txt_files)
    cookie_path = os.path.join(folder_path, random_cookie)
    return cookie_path

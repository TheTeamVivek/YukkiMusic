#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

import aiohttp
from .config import *



async def fetch_cookies():
    if not COOKIE_LINK:
        return None
    paste_id = COOKIE_LINK.split("/")[-1]
    raw_url = f"https://batbin.me/raw/{paste_id}"
    async with aiohttp.ClientSession() as session:
        async with session.get(raw_url) as response:
            if response.status == 200:
                raw_content = await response.text()
                with open("cookies/cookies.txt", "w", encoding="utf-8") as file:
                    file.write(raw_content)

                print("Cookies successfully written")
            else:
                print(f"Failed to get the URL. Status code: {response.status}")
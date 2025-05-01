#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from .apple import Apple
from .resso import Resso
from .savan import Saavn
from .soundcloud import SoundCloud
from .spotify import Spotify
from .telegram import Telegram
from .youtube import YouTube

apple = Apple()
saavn = Saavn()
resso = Resso()
soundcloud = SoundCloud()
spotify = Spotify()
telegram = Telegram()
youtube = YouTube()


async def track(url: str) -> "Track":
    services = [apple, saavn, resso, soundcloud, spotify, telegram, youtube]

    for x in services:
        if await x.valid(url):
            return await x.track(url)

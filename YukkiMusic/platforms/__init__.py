#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
from ..core.enum import SongType, SourceType
from . import carbon
from .apple import Apple
from .resso import Resso
from .savan import Saavn
from .soundcloud import SoundCloud
from .spotify import Spotify
from .telegram import Telegram
from .youtube import YouTube

apple = Apple()
carbon = Carbon()
saavn = Saavn()
resso = Resso()
soundcloud = SoundCloud()
spotify = Spotify()
telegram = Telegram()
youtube = YouTube()


async def valid(url: str) -> SourceType:
    services = [
        (apple, SourceType.APPLE),
        (saavn, SourceType.SAAVN),
        (resso, SourceType.RESSO),
        (soundcloud, SourceType.SOUNDCLOUD),
        (spotify, SourceType.SPOTIFY),
        (youtube, SourceType.YOUTUBE),
    ]

    for service, source_type in services:
        if await service.valid(url):
            return source_type


async def info(
    type: SourceType, **kwargs
) -> (
    dict
):  # todo implement all classes and there info function in this function using SourceType and **kwargs
    pass

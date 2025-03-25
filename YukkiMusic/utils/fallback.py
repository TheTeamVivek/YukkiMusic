# Radhe Radhe
from YukkiMusic.platforms.JioSavan import Saavn  # noqa

from .decorators.asyncify import asyncify


@asyncify
def fallback_download():
    pass


async def fallback_track(name):
    pass

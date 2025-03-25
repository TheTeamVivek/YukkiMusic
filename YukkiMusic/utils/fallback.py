# Radhe Radhe
from YukkiMusic.platforms.JioSavan import Saavn  # noqa

from .decorators.asyncify import asyncify


@asyncify
def download(title, video):
    video = None
    path, _ = await Saavn().download(title)
    return path, video


async def track(name):
    pass

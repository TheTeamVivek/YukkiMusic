# Radhe Radhe
from YukkiMusic.platforms.JioSavan import Saavn

from .decorators.asyncify import asyncify


async def download(title, video):
    video = None
    path, _ = await Saavn().download(title)
    return path, video

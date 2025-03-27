# Radhe Radhe
from YukkiMusic.platforms.JioSavan import Saavn


async def download(title, video):
    video = None
    path, details = await Saavn().download(title)
    return path, details, video

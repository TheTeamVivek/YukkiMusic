# Radhe Radhe
from YukkiMusic.platforms.JioSavan import Saavn


async def download(title, video):
    video = None
    path, _ = await Saavn().download(title)
    return path, video

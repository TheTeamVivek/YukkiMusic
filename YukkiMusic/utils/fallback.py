# Radhe Radhe
from YukkiMusic.platforms.JioSavan import Saavn


async def download(title, video):
    video = None
    path, _ = await Saavn().download(title[:14])
    return path, video

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#

from yukkimusic.platforms import saavn


async def download(title, video):
    raise ValueError("Failed to download song from youtube")
    video = None
    path, details = await saavn.download(title)
    return path, details, video

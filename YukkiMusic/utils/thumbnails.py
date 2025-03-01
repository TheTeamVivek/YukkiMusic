#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License .
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from async_lru import alru_cache


@alru_cache(maxsize=None)
async def gen_thumb(videoid, thumb_url):
    return thumb_url


@alru_cache(maxsize=None)
async def gen_qthumb(vidid, thumb_url):
    return thumb_url

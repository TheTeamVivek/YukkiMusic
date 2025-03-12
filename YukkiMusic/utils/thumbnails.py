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
async def gen_thumb(videoid, thumb_url=None):
    if thumb_url is not None:
        return thumb_url
    from YukkiMusic.core.youtube import search

    try:
        query = f"https://www.youtube.com/watch?v={videoid}"
        results = await search(query)
        return results.thumb
    except Exception:
        return f"https://img.youtube.com/vi/{videoid}/maxresdefault.jpg"


@alru_cache(maxsize=None)
async def gen_qthumb(vidid, thumb_url=None):
    return await gen_thumb(vidid, thumb_url)

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from asyncio import Lock
from functools import wraps

_locks: dict[str, Lock] = {}


def with_lock(key: str):
    """Decorator to run a coroutine with a per-key asyncio.Lock."""

    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            if key not in _locks:
                _locks[key] = Lock()
            async with _locks[key]:
                return await func(*args, **kwargs)

        return wrapper

    return decorator

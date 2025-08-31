import asyncio
from collections import defaultdict
from functools import wraps

_locks = defaultdict(asyncio.Lock)


def with_lock(key_fn_or_str):
    """
    Decorator to run a coroutine with a per-key asyncio.Lock.

    Args:
        key_fn_or_str (str | callable): A string key or a function that
        takes (*args, **kwargs) and returns a key.
    """

    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            key = (
                key_fn_or_str(*args, **kwargs)
                if callable(key_fn_or_str)
                else key_fn_or_str
            )
            async with _locks[key]:
                return await func(*args, **kwargs)

        return wrapper

    return decorator

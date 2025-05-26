import asyncio
from functools import wraps


def asyncify(sync_function):
    @wraps(sync_function)
    async def async_wrapper(*args, **kwargs):
        return await asyncio.to_thread(sync_function, *args, **kwargs)

    return async_wrapper

import asyncio
from functools import partial, wraps


def asyncify(func):
    @wraps(func)
    async def run(*args, **kwargs):
        loop = asyncio.get_running_loop()
        pfunc = partial(func, *args, **kwargs)
        return await loop.run_in_executor(None, pfunc)

    return run

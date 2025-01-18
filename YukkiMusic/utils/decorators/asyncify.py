import asyncio
from concurrent.futures import Executor
from functools import partial
from typing import Any

from decorator import decorator


@decorator
def asyncify(
    func,
    executor: type[Executor] | None = None,
    max_workers: int | None = None,
    *args: Any,
    **kwargs: Any
):
    async def run():
        loop = asyncio.get_running_loop()
        pfunc = partial(func, *args, **kwargs)

        if executor is None:
            return await loop.run_in_executor(None, pfunc)
        else:
            with executor(max_workers=max_workers) as exec:
                return await loop.run_in_executor(exec, pfunc)

    return run

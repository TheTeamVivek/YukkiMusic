import asyncio
from concurrent.futures import Executor
from functools import partial
from typing import Any

from decorator import decorator

__all__ = ["asyncify"]


@decorator
def asyncify(
    func,
    executor: Executor | None = None,
    max_workers: int | None = None,
    *args: Any,
    **kwargs: Any,
):
    async def run():
        pfunc = partial(func, *args, **kwargs)

        if executor is None:
            return await asyncio.to_thread(pfunc)
        else:
            loop = asyncio.get_running_loop()
            with executor(max_workers=max_workers) as exe:
                return await loop.run_in_executor(exe, pfunc)

    return run

import asyncio
from collections.abc import Callable
from functools import partial, wraps
from typing import Any

__all__ = ["asyncify"]


def asyncify(func: Callable) -> Callable[..., Any]:
    @wraps(func)
    def wrapper(*args: Any, **kwargs: Any) -> Any:
        async def run():
            pfunc = partial(func, *args, **kwargs)
            return await asyncio.to_thread(pfunc)

        return run()

    return wrapper

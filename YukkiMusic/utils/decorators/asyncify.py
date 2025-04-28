import asyncio
from functools import wraps, partial
from typing import Any, Callable

__all__ = ["asyncify"]


def asyncify(func: Callable) -> Callable[..., Any]:
    @wraps(func)
    def wrapper(*args: Any, **kwargs: Any) -> Any:
        async def run():
            pfunc = partial(func, *args, **kwargs)
            return await asyncio.to_thread(pfunc)
        return run()
    return wrapper
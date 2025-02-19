import asyncio
import inspect


class Combinator:
    def __init__(self, func):
        self.func = func

    async def __call__(self, event):
        if inspect.iscoroutinefunction(self.func):
            return await self.func(event)
        return await asyncio.to_thread(self.func, event)

    def __and__(self, other):
        async def and_func(event):
            return (await self(event)) and (await other(event))

        return Combinator(and_func)

    def __or__(self, other):
        async def or_func(event):
            return (await self(event)) or (await other(event))

        return Combinator(or_func)

    def __invert__(self):
        async def not_func(event):
            return not (await self(event))

        return Combinator(not_func)


def wrap(func):
    return Combinator(func)


@wrap
def private(event):
    return getattr(event, "is_private", False)

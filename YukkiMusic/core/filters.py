import asyncio
import inspect
import re
from string import get_string
from YukkiMusic.utils.database import get_lang

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


@wrap
def command(commands, use_string=False):
    if isinstance(commands, str):
        commands = [commands]

    async def func(event):
        if use_string:
            try:
                lang = await get_lang(event.chat_id)
                string = get_string(lang)
            except Exception:
                string = get_string("en")

            command_list = [string.get(cmd, cmd) for cmd in commands]
        else:
            command_list = commands

        command_pattern = "|".join([re.escape(cmd) for cmd in command_list])

        user = await event.client.get_me()
        username = re.escape(user.username) if user and user.username else ""

        pattern = re.compile(
            rf"^(?:/)?({command_pattern})(?:@{username})?(?:\s|$)",
            re.IGNORECASE,
        )

        return bool(re.match(pattern, event.text))

    return func

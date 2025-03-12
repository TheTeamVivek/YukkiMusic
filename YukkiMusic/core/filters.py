import asyncio
import inspect
import re
from collections.abc import Callable

from telethon.tl.types import PeerChannel, User

from strings import get_string
from YukkiMusic.utils.database import get_lang


class Filter:
    def __init__(self, func: Callable = None):
        self.func = func

    async def __call__(self, event):
        if self.func is not None:
            return (
                await self.func(event)
                if inspect.iscoroutinefunction(self.func)
                else await asyncio.to_thread(self.func(event))
            )
        return False

    def __and__(self, other):
        "And Filter"

        async def and_filter(event):
            x = await self(event)
            y = await other(event)
            if not x:
                return False
            return x and y

        return Filter(and_filter)

    def __or__(self, other):
        "Or Filter"

        async def or_filter(event):
            x = await self(event)
            y = await other(event)
            if x:
                return True
            return x or y

        return Filter(or_filter)

    def __invert__(self):
        "Invert Filter"

        async def invert_filter(event):
            return not (await self(event))

        return Filter(invert_filter)


def wrap(func):
    "wrap the function by Filter"
    return Filter(func=func)


@wrap
def forwarded(e):
    "Message is forwarded"
    return bool(getattr(e, "forward", None))


@wrap
def new_chat_members(event):  # May be only useable in events.ChatAction
    "Member is joined or added in chat"
    return getattr(event, "user_added", False) or getattr(event, "user_joined", False)


@wrap
def private(event):
    """Check if the chat is private."""
    return getattr(event, "is_private", False)


@wrap
def group(event):
    """Check if the chat is a group or supergroup."""
    return getattr(event, "is_group", False)


@wrap
async def channel(event):
    """Check if the chat is a Channel (not a MegaGroup)."""
    msg = getattr(event, "message", None)
    peer = getattr(msg, "peer_id", None) if msg else None

    if isinstance(peer, PeerChannel):
        entity = await event.client.get_entity(peer)
        return not getattr(entity, "megagroup", False)

    return False


class User(set, Filter):
    """Check if the sender is a specific user."""

    def __init__(self, users: int | str | list[int, str] | None = None):
        users = [] if users is None else users if isinstance(users, list) else [users]

        super().__init__(
            (
                "me"
                if u in ["me", "self"]
                else u.lower().strip("@") if isinstance(u, str) else u
            )
            for u in users
        )

    async def func(self, event):
        sender = await event.get_sender()
        return isinstance(sender, User) and (
            sender.id in self
            or (sender.username and sender.username.lower() in self)
            or ("me" in self and sender.is_self)
        )


@wrap
def command(commands, use_strings=False):
    "Check if the message startswith the provided command"
    if isinstance(commands, str):
        commands = [commands]

    async def func(event):
        text = event.text
        if not text:
            return False

        username = (
            event.client.username.lower()
        )  # Because this event.client is Bot client so username can't be None

        if use_strings:
            lang = await get_lang(event.chat_id)
            lang = get_string(lang)

            commands = {
                lang.get(cmd, cmd) for cmd in commands
            }  # Get the command from string if use_strings is True if the command is not found on string so use the command
        commands = {cmd.lower() for cmd in commands}
        command_pattern = "|".join(map(re.escape, commands))
        pattern = rf"^(?:/)?({command_pattern})(?:@{re.escape(username)})?(?:\s|$)"

        return bool(re.match(pattern, text, flags=re.IGNORECASE))

    return func

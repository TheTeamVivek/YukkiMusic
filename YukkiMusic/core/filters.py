#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio as _asyncio
import inspect as _inspect
import re as _re

from telethon.tl import types as _types

from strings import get_command as _get_command

__all__ = [
    "Filter",
    "wrap",
    "forwarded",
    "new_chat_members",
    "private",
    "group",
    "channel",
    "User",
    "command",
]


class Filter:
    def __init__(self, func):
        self.func = func

    async def __call__(self, event):
        if self.func is not None:
            if _inspect.iscoroutinefunction(self.func):
                return await self.func(event)
            else:
                return await _asyncio.to_thread(self.func, event)
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
    return Filter(func)


def wrap(func):
    "wrap the function by Filter"
    return Filter(func)


@wrap
def forwarded(e):
    "Message is forwarded"
    return bool(getattr(e, "forward", None))


@wrap
def new_chat_members(event):  # May be only usable in events.ChatAction
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

    if isinstance(peer, _types.PeerChannel):
        entity = await event.client.get_entity(peer)
        return not getattr(entity, "megagroup", False)

    return False


class User(set, Filter):
    def __init__(self, users=None):
        users = [] if users is None else users if isinstance(users, list) else [users]

        cleaned = {
            (
                "me"
                if u in ("me", "self")
                else u.lower().strip("@")
                if isinstance(u, str)
                else u
            )
            for u in users
        }

        set.__init__(self, cleaned)
        Filter.__init__(self, self.check_user)

    async def check_user(self, event):
        sender = await event.get_sender()
        return (
            sender.id in self
            or (sender.username and sender.username.lower() in self)
            or ("me" in self and sender.is_self)
        )


user = User


def _normalize_command(value):
    if isinstance(value, list):
        return {v.lower() for v in value if isinstance(v, str)}
    elif isinstance(value, str):
        return {value.lower()}
    return set()


def command(commands, use_strings=False):
    "Check if the message starts with the provided command(s)"
    if isinstance(commands, str):
        commands = [commands]

    @wrap
    async def filter_func(event):
        message_text = event.text
        if not message_text:
            return False

        username = event.client.username.lower()
        final_commands = []

        if use_strings:
            for cmd in commands:
                final_commands.extend(_get_command(cmd))
        else:
            final_commands = commands

        escaped = map(_re.escape, final_commands)
        pattern = rf"^(?:/)?({'|'.join(escaped)})(?:@{username})?(?:\s|$)"
        return bool(_re.match(pattern, message_text, flags=_re.IGNORECASE))

    return filter_func

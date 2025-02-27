import inspect
import re

from telethon.tl.types import PeerChannel, User

from strings import get_string
from YukkiMusic.utils.database import get_lang


class Combinator:
    def __init__(self, func):
        self.func = func

    async def __call__(self, event):
        return (
            await self.func(event)
            if inspect.iscoroutinefunction(self.func)
            else self.func(event)
        )

    def __and__(self, other):
        async def combined(event):
            return (await self(event)) and (await other(event))

        return Combinator(combined)

    def __or__(self, other):
        async def combined(event):
            return (await self(event)) or (await other(event))

        return Combinator(combined)

    def __invert__(self):
        async def inverted(event):
            return not (await self(event))

        return Combinator(inverted)


def wrap(func):
    return Combinator(func)


@wrap
def new_chat_members(event): # May be only useable in events.ChatAction
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
    """Check if the chat is a Channel (not a Mega Group)."""
    msg = getattr(event, "message", None)
    peer = getattr(msg, "peer_id", None) if msg else None

    if isinstance(peer, PeerChannel):
        entity = await event.client.get_entity(peer)
        return not getattr(entity, "megagroup", False)

    return False


@wrap
def user(users):
    """Check if the sender is a specific user."""

    if isinstance(users, (int, str)):
        users = {users}
    else:
        users = set(users)

    normalized_users = {
        str(user).lower().lstrip("@") if isinstance(user, str) else user
        for user in users
    }

    async def check_user(event):
        sender = await event.get_sender()
        if sender is None:
            return False
        if not isinstance(sender, User):
            return False

        user_id = sender.id
        username = sender.username.lower() if sender.username else None

        if "me" in normalized_users or "self" in normalized_users:
            normalized_users.update(
                {event.client.me.id, event.client.me.username.lower()}
                if event.client.me.username
                else {event.client.me.id}
            )

        return user_id in normalized_users or (
            username in normalized_users if username else False
        )

    return check_user


@wrap
def command(commands, use_strings=False):
    if isinstance(commands, str):
        commands = [commands]

    async def func(event):
        text = event.text.lstrip()
        if not text:
            return False

        user = await event.client.get_me()
        username = user.username.lower() if user and user.username else ""

        if use_strings:
            try:
                lang = await get_lang(event.chat_id)
                string = get_string(lang)
            except Exception:
                string = get_string("en")

            command_list = {string.get(cmd, cmd) for cmd in commands}
        else:
            command_list = set(commands)

        pattern = rf"^(?:/)?({'|'.join(map(re.escape, command_list))})(?:@{re.escape(username)})?(?:\s|$)"

        return bool(re.match(pattern, text))

    return func

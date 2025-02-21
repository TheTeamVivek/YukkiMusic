import inspect
import re
from string import get_string

from telethon.tl.types import PeerChannel, User

from YukkiMusic.utils.database import get_lang


class Combinator:
    def __init__(self, func):
        self.func = func

    async def __call__(self, event):
        if inspect.iscoroutinefunction(self.func):
            return await self.func(event)
        return self.func(event)

    def __and__(self, other):
        return Combinator(lambda event: (await self(event)) and (await other(event)))

    def __or__(self, other):
        return Combinator(lambda event: (await self(event)) or (await other(event)))

    def __invert__(self):
        return Combinator(lambda event: not (await self(event)))

def wrap(func):
    return Combinator(func)


@wrap
def private(event):
    "Chat is Private"
    return getattr(event, "is_private", False)


@wrap
def group(event):
    "Chat is Group or Supergroup"
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
    """Check if the sender is a specific user"""

    async def check_user(event):
        if not getattr(event, "sender_id", False):
            return False
        sender = await event.get_sender()

        if not isinstance(sender, User):
            return False

        user_id = sender.id
        username = sender.username.lower() if sender.username else None

        if isinstance(users, (int, str)):
            users_set = {users}
        else:
            users_set = set(users)

        users = set()
        for user in users_set:
            if isinstance(user, int):
                users.add(user)
            else:
                users.add(str(user).lower().lstrip("@"))

        if "me" in users or "self" in users:
            users.add(event.client.me.id)
            if event.client.me.username:
                users.add(event.client.me.username.lower())

        return user_id in users or (username in users if username else False)

    return check_user


@wrap
def command(commands, use_strings=False):
    if isinstance(commands, str):
        commands = [commands]

    async def func(event):
        if use_strings:
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

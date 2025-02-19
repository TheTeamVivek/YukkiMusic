import asyncio
import inspect
import re
import traceback
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime
from string import get_string
from typing import Any

from telethon import TelegramClient, events
from telethon.errors import UserNotParticipantError
from telethon.tl.functions.channels import (
    GetParticipantRequest,
    LeaveChannelRequest,
)
from telethon.tl.functions.messages import (
    DeleteChatUserRequest,
    GetFullChatRequest,
)
from telethon.tl.types import (
    ChannelParticipant,
    ChannelParticipantAdmin,
    ChannelParticipantBanned,
    ChannelParticipantCreator,
    ChannelParticipantLeft,
    ChannelParticipantSelf,
    InputPeerChannel,
    InputPeerChat,
    InputUserSelf,
    PeerChannel,
    PeerChat,
    User,
)

from YukkiMusic.utils.database import get_lang

from ..logging import logger

log = logger(__name__)


@dataclass
class ShellRunResult:
    returncode: int
    stdout: str
    stderr: str


class TelethonClient(TelegramClient):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.__lock = asyncio.Lock()
        self.__tasks = []

    async def run_coro(self, func: Callable, *args, **kwargs):
        if inspect.iscoroutinefunction(func):
            r = await func(*args, **kwrags)
        else:
            r = await asyncio.to_thread(func, *args, **kwargs)
        return r

    async def create_mention(self, user: User, html: bool = False) -> str:
        user_name = f"{user.first_name} {user.last_name or ''}".strip()
        user_id = user.id
        if html:
            return f'<a href="tg://user?id={user_id}">{user_name}</a>'
        return f"[{user_name}](tg://user?id={user_id})"

    async def leave_chat(self, chat_id):
        entity = await self.get_entity(chat_id)
        if isinstance(entity, PeerChannel):
            await self(LeaveChannelRequest(entity))
        elif isinstance(entity, PeerChat):
            await self(DeleteChatUserRequest(entity.id, InputUserSelf()))

    async def get_chat_member(
        self,
        chat_id: int | str,
        user_id: int | str,
    ):
        chat = await self.get_entity(chat_id)
        user = await self.get_entity(user_id)

        status_map = {
            "BANNED": ChannelParticipantBanned,
            "LEFTED": ChannelParticipantLeft,
            "OWNER": ChannelParticipantCreator,
            "ADMIN": ChannelParticipantAdmin,
            "SELF": ChannelParticipantSelf,
            "MEMBER": ChannelParticipant,
        }

        if isinstance(chat, InputPeerChat):
            r = await self(GetFullChatRequest(chat_id=chat.chat_id))

            members = getattr(r.full_chat.participants, "participants", [])

            for member in members:
                if member.user_id == user.user_id:
                    for status, cls in status_map.items():
                        if isinstance(member, cls):
                            return member, status
                    raise UserNotParticipantError

        elif isinstance(chat, InputPeerChannel):
            r = await self(GetParticipantRequest(channel=chat, participant=user))
            participant = r.participant
            for status, cls in status_map.items():
                if isinstance(participant, cls):
                    return participant, status
        else:
            raise ValueError(f'The chat_id "{chat_id}" belongs to a user')

    async def handle_error(self, exc: Exception):  # TODO Make it more brief
        date_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        error_trace = traceback.format_exc()

        error_message = (
            f"**Error:** {type(exc).__name__}\n"
            f"**Date:** {date_time}\n"
            f"**Traceback:**\n{error_trace}"
        )

        await self.send_message(config.LOG_GROUP_ID, error_message)

        try:
            await self.send_message(config.OWNER_ID[0], error_message)
        except Exception:
            pass
        log.error(error_trace)

    async def start(self, *args, **kwargs):
        await super.start(*args, **kwargs)
        me = await self.get_me()
        # pylint: disable=attribute-defined-outside-init
        self.me = me
        self.id = me.id
        self.username = me.username
        self.mention = self.create_mention(me)
        self.name = f"{me.first_name} {me.last_name or ''}".strip()
        # pylint: enable=attribute-defined-outside-init
        asyncio.create_task(self.__task_runner())

    def on_message(self, command, **kwargs):
        def decorator(function):
            kwargs["incoming"] = kwargs.get("incoming", True)
            func = kwargs.get("func")

            async def custom_func(event):
                if func and not await self.run_coro(func, event):
                    return False

                if isinstance(command, str):
                    command_list = [command]
                else:
                    command_list = command

                try:
                    lang = await get_lang(event.chat_id)
                    string = get_string(lang)
                except Exception:
                    string = get_string(lang)

                command_list = [string[cmd] for cmd in command_list]
                command_pattern = "|".join([re.escape(cmd) for cmd in command_list])

                user = await event.client.get_me()
                username = re.escape(user.username) if user and user.username else ""

                pattern = re.compile(
                    rf"^(?:/)?({command_pattern})(?:@{username})?(?:\s|$)",
                    re.IGNORECASE,
                )

                return bool(re.match(pattern, event.text))

            kwargs["func"] = custom_func
            kwargs.pop("pattern", None)

            self.add_event_handler(function, events.NewMessage(**kwargs))
            return function

        return decorator

    async def __task_runner(self):
        while True:
            async with self.__lock:
                if not self.__tasks:
                    return
                tasks = [
                    self.run_coro(func, *args, **kwargs)
                    for func, args, kwargs in self.__tasks
                ]
                self.__tasks.clear()

            results = await asyncio.gather(*tasks, return_exceptions=True)
            for r in results:
                if isinstance(r, Exception):  # Check if the result is an exception
                    await self.handle_error(
                        r
                    )  # Log the error traceback and inform to OWNER
            await asyncio.sleep(0.2)

    async def add_task(self, func: Callable[..., Any], *args, **kwargs):
        async with self.__lock:
            self.__tasks.append((func, args, kwargs))

    async def run(self, command: list):
        process = await asyncio.create_subprocess_exec(
            *command,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )

        stdout, stderr = await process.communicate()

        return ShellRunResult(
            returncode=process.returncode,
            stdout=stdout.decode().strip() if stdout else None,
            stderr=stderr.decode().strip() if stderr else None,
        )

import asyncio
import inspect
import traceback
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime
from functools import wraps

from telethon import TelegramClient, events
from telethon.errors import (
    ChatSendMediaForbiddenError,
    ChatSendPhotosForbiddenError,
    ChatWriteForbiddenError,
    FloodWaitError,
    MessageIdInvalidError,
    MessageNotModifiedError,
    UserNotParticipantError,
)
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

    async def run_coro(self, func: Callable, err: bool = True, *args, **kwargs):
        try:
            if inspect.iscoroutinefunction(func):
                r = await func(*args, **kwargs)
            else:
                r = await asyncio.to_thread(func, *args, **kwargs)
            return r
        except Exception as e:
            if err:
                raise e

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

    def on_message(self, func=None, *args, **kwargs):
        def decorator(function):
            @wraps(function)
            async def wrapper(event):
                try:
                    return await function(event)
                except FloodWaitError as e:
                    log.warning("FloodWait: Sleeping for %d seconds.", e.value)
                    await asyncio.sleep(e.value)
                except (
                    ChatWriteForbiddenError,
                    ChatSendMediaForbiddenError,
                    ChatSendPhotosForbiddenError,
                    MessageNotModifiedError,
                    MessageIdInvalidError,
                ) as e:
                    if isinstance(e, ChatWriteForbiddenError):
                        await self.run_coro(
                            event.chat_id, func=self.leave_chat, err=False
                        )  # using for disable errors

                except events.StopPropagation as e:
                    raise events.StopPropagation from e

                except Exception as e:
                    await self.handle_error(e)

            if func is not None:
                kwargs["func"] = func
            kwargs["incoming"] = kwargs.get("incoming", True)
            self.add_event_handler(wrapper, events.NewMessage(*args, **kwargs))
            return wrapper

        return decorator

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

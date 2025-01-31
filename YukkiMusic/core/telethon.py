import re
from functools import wraps

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


class TelethonClient(TelegramClient):
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

    async def start(self, *arg, **kwarg):
        await self.start(*arg, **kwarg)
        me = await self.get_me()
        self.me = me
        self.id = me.id
        self.username = me.username
        self.mention = self.create_mention(me)
        self.name = f"{me.first_name} {me.last_name or ''}".strip()

    def on_message(self, command, **kwargs):
        def decorator(function):
            @wraps(function)
            async def wrapper(event):
                kwargs["incoming"] = kwargs.get("incoming") or True
                command = [command] if isinstance(command, str) else command
                # command = get_command(command, "en") #todo
                command = [re.escape(cmd) for cmd in command]
                command = "|".join(command)
                username = re.escape(self.username)
                pattern = re.compile(
                    rf"^(?:/)?({command})(?:@{username})?(?:\s|$)", re.IGNORECASE
                )
                kwargs["pattern"] = pattern
                await function(event)

            self.add_event_handler(wrapper, events.NewMessage(**kwargs))
            return wrapper

        return decorator

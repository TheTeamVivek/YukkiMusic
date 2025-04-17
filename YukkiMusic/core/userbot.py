#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import sys
import traceback
from datetime import datetime
from functools import wraps

from pyrogram import Client, StopPropagation
from pyrogram.errors import (
    ChatSendMediaForbidden,
    ChatSendPhotosForbidden,
    ChatWriteForbidden,
    FloodWait,
    MessageIdInvalid,
    MessageNotModified,
)
from pyrogram.handlers import MessageHandler

import config

from ..logging import logger

assistants = []
assistantids = []


class Userbot(Client):
    def __init__(self):
        self.clients = [
            Client(
                f"YukkiString_{i}",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                session_string=session.strip(),
            )
            for i, session in enumerate(config.STRING_SESSIONS, start=1)
        ]
        self.handlers = []

    async def _start(self, client, index):
        logger(__name__).info("Starting Assistant Client %d", index)
        try:
            await client.start()
            assistants.append(index)
            try:
                await client.send_message(config.LOG_GROUP_ID, "Assistant Started")
            except ChatWriteForbidden:
                try:
                    await client.join_chat(config.LOG_GROUP_ID)
                    await client.send_message(config.LOG_GROUP_ID, "Assistant Started")
                except Exception:
                    logger(__name__).error(
                        "Assistant Account %d failed to send message in log group. "
                        "Ensure the assistant is added to the log group.",
                        index,
                    )
                    sys.exit(1)

            get_me = await client.get_me()
            client.username = get_me.username
            client.id = get_me.id
            client.mention = get_me.mention
            assistantids.append(get_me.id)
            client.name = f"{get_me.first_name} {get_me.last_name or ''}".strip()

            # Add stored handlers to the client
            for handler, group in self.handlers:
                client.add_handler(handler, group)

        except Exception as e:
            logger(__name__).error(
                "Assistant Account {index} failed with error: %s xiting...", e
            )
            sys.exit(1)

    async def start(self):
        """Start all clients."""
        tasks = [
            self._start(client, i) for i, client in enumerate(self.clients, start=1)
        ]
        await asyncio.gather(*tasks)

    async def stop(self):
        """Gracefully stop all clients."""
        tasks = [client.stop() for client in self.clients]
        await asyncio.gather(*tasks)

    def on_message(
        self, filters=None, group=0
    ):  # on_message decirator for future Userbot Plugins
        """Decorator for handling messages with error handling."""

        def decorator(func):
            @wraps(func)
            async def wrapper(client, message):
                try:
                    await func(client, message)
                except FloodWait as e:
                    logger(__name__).warning(
                        "FloodWait: Sleeping for %d seconds.", e.value
                    )
                    await asyncio.sleep(e.value)
                except (
                    ChatWriteForbidden,
                    ChatSendMediaForbidden,
                    ChatSendPhotosForbidden,
                    MessageNotModified,
                    MessageIdInvalid,
                ):
                    pass
                except StopPropagation:
                    raise
                except Exception as e:
                    # Detailed error logging
                    date_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                    user_id = message.from_user.id if message.from_user else "Unknown"
                    chat_id = message.chat.id if message.chat else "Unknown"
                    chat_username = (
                        f"@{message.chat.username}"
                        if message.chat.username
                        else "Private Group"
                    )
                    command = (
                        " ".join(message.command)
                        if hasattr(message, "command")
                        else message.text
                    )
                    error_trace = traceback.format_exc()
                    error_message = (
                        f"**Error:** {type(e).__name__}\n"
                        f"**Date:** {date_time}\n"
                        f"**Chat ID:** {chat_id}\n"
                        f"**Chat Username:** {chat_username}\n"
                        f"**User ID:** {user_id}\n"
                        f"**Command/Text:** {command}\n"
                        f"**Traceback:**\n{error_trace}"
                    )
                    await client.send_message(config.LOG_GROUP_ID, error_message)
                    try:
                        await client.send_message(config.OWNER_ID[0], error_message)
                    except Exception:
                        pass

            handler = MessageHandler(wrapper, filters)
            self.handlers.append((handler, group))
            return func

        return decorator


userbot = Userbot()

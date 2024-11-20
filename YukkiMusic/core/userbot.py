#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import sys

from pyrogram import Client
from pyrogram.errors import ChatWriteForbidden
import config

from ..logging import LOGGER

assistants = []
assistantids = []


class Userbot(Client):
    def __init__(self):
        self.clients = []
        session_strings = config.STRING_SESSIONS

        for i, session in enumerate(session_strings, start=1):

            client = Client(
                f"YukkiString{i}",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                in_memory=True,
                no_updates=True,
                session_string=session.strip(),
            )
            self.clients.append(client)

    async def _start(self, client, index):
        LOGGER(__name__).info("Starting Assistant Clients")
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
                    LOGGER(__name__).error(
                    f"Assistant Account {index} has failed to send message in Loggroup Make sure you have added assistsant in Loggroup."
                    )
                    sys.exit(1)

            get_me = await client.get_me()
            client.username = get_me.username
            client.id = get_me.id
            client.mention = get_me.mention
            assistantids.append(get_me.id)
            client.name = f"{get_me.first_name} {get_me.last_name or ''}".strip()

        except Exception as e:
            LOGGER(__name__).error(
                f"Assistant Account {index} failed with error: {str(e)}."
            )
            sys.exit(1)

    async def start(self):
        tasks = []  # List to hold start tasks
        for i, client in enumerate(self.clients, start=1):
            task = self._start(client, i)
            tasks.append(task)
        await asyncio.gather(*tasks)

    async def stop(self):
        """Gracefully stop all clients."""
        tasks = [client.stop() for client in self.clients]
        await asyncio.gather(*tasks)

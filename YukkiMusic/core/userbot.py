#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import os
import sys
from pyrogram import Client, filters
from ..logging import LOGGER
import config
from YukkiMusic.misc import SUDOERS

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
                session_string=session.strip(),
            )
            self.clients.append(client)

    async def _start(self, client, index):
        LOGGER(__name__).info("Starting Assistant Clients")
        try:
            await client.start()
            try:
                await client.join_chat("TheYukki")
                await client.join_chat("YukkiSupport")
                await client.join_chat("TheTeamVivek")
            except:
                pass

            assistants.append(index)  # Mark the assistant as active

            await client.send_message(config.LOG_GROUP_ID, "Assistant Started")

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
        LOGGER(__name__).info("Stopping all assistant clients...")
        tasks = [client.stop() for client in self.clients]
        await asyncio.gather(*tasks)
        LOGGER(__name__).info("All assistant clients stopped.")

#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#

import sys
from pyrogram import Client
import config
from ..logging import LOGGER

assistants = []
assistantids = []

#जय श्री राम प्रभु अगर ये युक्ति काम कर गया तो मैं तो यह से ये वाक्य नहीं हटाउॅगा

class Userbot(Client):
    def __init__(self):
        super().__init__(
            "YukkiUserbot",
            api_id=config.API_ID,
            api_hash=config.API_HASH,
            session_name="yukki_session",
            bot_token=config.BOT_TOKEN
        )
        self.dispatcher = self.add_handler()
        
        if config.STRING1:
            self.one = Client(
                "YukkiString1",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                session_string=str(config.STRING1),
                no_updates=True,
            )
        else:
            self.one = None

        if config.STRING2:
            self.two = Client(
                "YukkiString2",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                session_string=str(config.STRING2),
                no_updates=True,
            )
        else:
            self.two = None

        if config.STRING3:
            self.three = Client(
                "YukkiString3",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                session_string=str(config.STRING3),
                no_updates=True,
            )
        else:
            self.three = None

        if config.STRING4:
            self.four = Client(
                "YukkiString4",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                session_string=str(config.STRING4),
                no_updates=True,
            )
        else:
            self.four = None

        if config.STRING5:
            self.five = Client(
                "YukkiString5",
                api_id=config.API_ID,
                api_hash=config.API_HASH,
                session_string=str(config.STRING5),
                no_updates=True,
            )
        else:
            self.five = None
            
    async def start(self):
        await super().start()
        if self.one:
            try:
                await self.one.join_chat("TeamYM")
                await self.one.join_chat("TheYukki")
                await self.one.join_chat("YukkiSupport")
            except:
                pass
            assistants.append(1)
            try:
                await self.one.send_message(config.LOG_GROUP_ID, "ᴀssɪsᴛᴀɴᴛ sᴛᴀʀᴛᴇᴅ")
            except:
                LOGGER(__name__).error(
                    f"Assistant Account 1 has failed to access the log Group. Make sure that you have added your assistant to your log group and promoted as admin! "
                )
                sys.exit()
            get_me = await self.one.get_me()
            self.one.username = get_me.username
            self.one.id = get_me.id
            self.one.mention = get_me.mention
            assistantids.append(get_me.id)
            if get_me.last_name:
                self.one.name = get_me.first_name + " " + get_me.last_name
            else:
                self.one.name = get_me.first_name
            LOGGER(__name__).info(f"Assistant Started as {self.one.name}")
        if self.two:
            try:
                await self.two.join_chat("TeamYM")
                await self.two.join_chat("TheYukki")
                await self.two.join_chat("YukkiSupport")
            except:
                pass
            assistants.append(2)
            try:
                await self.two.send_message(config.LOG_GROUP_ID, "ᴀssɪsᴛᴀɴᴛ sᴛᴀʀᴛᴇᴅ")
            except:
                LOGGER(__name__).error(
                    f"Assistant Account 2 has failed to access the log Group. Make sure that you have added your assistant to your log group and promoted as admin! "
                )
                sys.exit()
            get_me = await self.two.get_me()
            self.two.username = get_me.username
            self.two.id = get_me.id
            self.two.mention = get_me.mention
            assistantids.append(get_me.id)
            if get_me.last_name:
                self.two.name = get_me.first_name + " " + get_me.last_name
            else:
                self.two.name = get_me.first_name
            LOGGER(__name__).info(f"Assistant Two Started as {self.two.name}")
        if self.three:
            try:
                await self.three.join_chat("TeamYM")
                await self.three.join_chat("TheYukki")
                await self.three.join_chat("YukkiSupport")
            except:
                pass
            assistants.append(3)
            try:
                await self.three.send_message(config.LOG_GROUP_ID, "ᴀssɪsᴛᴀɴᴛ sᴛᴀʀᴛᴇᴅ")
            except:
                LOGGER(__name__).error(
                    f"Assistant Account 3 has failed to access the log Group. Make sure that you have added your assistant to your log group and promoted as admin! "
                )
                sys.exit()
            get_me = await self.three.get_me()
            self.three.username = get_me.username
            self.three.id = get_me.id
            self.three.mention = get_me.mention
            assistantids.append(get_me.id)
            if get_me.last_name:
                self.three.name = get_me.first_name + " " + get_me.last_name
            else:
                self.three.name = get_me.first_name
            LOGGER(__name__).info(f"Assistant Three Started as {self.three.name}")
        if self.four:
            try:
                await self.four.join_chat("TeamYM")
                await self.four.join_chat("TheYukki")
                await self.four.join_chat("YukkiSupport")
            except:
                pass
            assistants.append(4)
            try:
                await self.four.send_message(config.LOG_GROUP_ID, "ᴀssɪsᴛᴀɴᴛ sᴛᴀʀᴛᴇᴅ")
            except:
                LOGGER(__name__).error(
                    f"Assistant Account 4 has failed to access the log Group. Make sure that you have added your assistant to your log group and promoted as admin! "
                )
                sys.exit()
            get_me = await self.four.get_me()
            self.four.username = get_me.username
            self.four.id = get_me.id
            self.four.mention = get_me.mention
            assistantids.append(get_me.id)
            if get_me.last_name:
                self.four.name = get_me.first_name + " " + get_me.last_name
            else:
                self.four.name = get_me.first_name
            LOGGER(__name__).info(f"Assistant Four Started as {self.four.name}")
        if self.five:
            try:
                await self.five.join_chat("TeamYM")
                await self.five.join_chat("TheYukki")
                await self.five.join_chat("YukkiSupport")
            except:
                pass
            assistants.append(5)
            try:
                await self.five.send_message(config.LOG_GROUP_ID, "ᴀssɪsᴛᴀɴᴛ sᴛᴀʀᴛᴇᴅ")
            except:
                LOGGER(__name__).error(
                    f"Assistant Account 5 has failed to access the log Group. Make sure that you have added your assistant to your log group and promoted as admin! "
                )
                sys.exit()
            get_me = await self.five.get_me()
            self.five.username = get_me.username
            self.five.id = get_me.id
            self.five.mention = get_me.mention
            assistantids.append(get_me.id)
            if get_me.last_name:
                self.five.name = get_me.first_name + " " + get_me.last_name
            else:
                self.five.name = get_me.first_name
            LOGGER(__name__).info(f"Assistant Five Started as {self.five.name}")
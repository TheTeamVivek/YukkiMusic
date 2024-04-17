import asyncio

from config import (
    AUTO_GCAST,
    AUTO_GCAST_MSG,
    LOGGER_ID,
    AUTO_GCAST_DELAY_TIME,
)
from YukkiMusic import app
from YukkiMusic.utils.database import get_served_users
from pyrogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from pyrogram.errors import FloodWait


MESSAGE = f"""нєу, ɪ ᴀᴍ {app.mention}

✰ I'ᴍ ᴀ ᴛᴇʟᴇɢʀᴀᴍ sᴛʀᴇᴀᴍɪɴɢ ʙᴏᴛ ᴡɪᴛʜ sᴏᴍᴇ ᴜsᴇғᴜʟ ғᴇᴀᴛᴜʀᴇs.

Sᴜᴘᴘᴏʀᴛɪɴɢ ᴘʟᴀᴛғᴏʀᴍs :
➪ ᴀᴘᴘʟᴇ
➪ ʀᴇssᴏ
➪ Sᴏᴜɴᴅᴄʟᴏᴜᴅ
➪ Sᴘᴏᴛɪғʏ
➪ ʏᴏᴜᴛᴜʙᴇ
➪ ᴛᴇʟᴇɢʀᴀᴍ [ ᴀᴜᴅɪᴏ + ᴠɪᴅᴇᴏ ʟᴏᴄᴀʟ ғɪʟᴇ]

✰ ᴀᴅs ғʀᴇᴇ ᴍᴜsɪᴄ ʙᴏᴛ ʙᴀsᴇᴅ ᴏɴ ʏᴜᴋᴋɪ's ʀᴇᴘᴏ ᴡɪᴛʜ ᴇxᴛʀᴀ ғᴇᴀᴛᴜʀᴇs ᴀɴᴅ ғɪxᴇᴅ ʙᴜɢ's

✰ Fᴇᴇʟ ғʀᴇᴇ ᴛᴏ ᴀᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘs."""

BUTTON = InlineKeyboardMarkup(
    [
        [
            InlineKeyboardButton(f"᯽ 𝙺ɪᴅɴᴀᴘ 𝙼ᴇ ᯽", url=f"https://t.me/{app.username}?startgroup=true")
        ]
    ]
)

MSG = AUTO_GCAST_MSG if AUTO_GCAST_MSG else MESSAGE

TEXT = """**ᴀᴜᴛᴏ ɢᴄᴀsᴛ ɪs ᴇɴᴀʙʟᴇᴅ sᴏ ᴀᴜᴛᴏ ɢᴄᴀsᴛ/ʙʀᴏᴀᴅᴄᴀsᴛ ɪs ᴅᴏɪɴ ɪɴ ᴀʟʟ  ᴄᴏɴᴛɪɴᴜᴏᴜsʟʏ ᴛᴏ ᴀʟʟ ᴜsᴇʀs. **\n**ɪᴛ ᴄᴀɴ ʙᴇ sᴛᴏᴘᴘᴇᴅ ʙʏ ᴘᴜᴛ ᴠᴀʀɪᴀʙʟᴇ [ᴀᴜᴛᴏ_ɢᴄᴀsᴛ = (ᴋᴇᴇᴘ ʙʟᴀɴᴋ & ᴀɴᴅ sᴇᴛ ᴛᴏ ғᴀʟsᴇ)]**"""

async def send_notice():
    try:
        await app.send_message(LOGGER_ID, TEXT)
    except :
        pass

async def send_message_to_users():
    try:
        users = await get_served_users()

        for i in users:
            user_id = i.get('user_id')
            if isinstance(user_id, int): 
                try:
                    await app.send_message(user_id, text=MSG, reply_markup=BUTTON)
                except FloodWait as e:
                    await asyncio.sleep(e.value)
                except:
                    pass
    except:
        pass

async def continuous_broadcast():
    await send_notice() 
    while True:
    	try:
            await send_message_to_users()
        except:
            pass
        await asyncio.sleep(AUTO_GCAST_DELAY_TIME)
        
if AUTO_GCAST == str(True):
    asyncio.create_task(continuous_broadcast())
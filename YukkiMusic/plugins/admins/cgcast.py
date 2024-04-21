import asyncio
from pyrogram import filters
from pyrogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_served_chats,
    get_served_users,
)
@app.on_message(filters.command(["gchats", "guser"]) & SUDOERS)
async def CGAST(_, message: Message):
    query = f"""нєу, ɪ ᴀᴍ {app.mention}

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
                InlineKeyboardButton(
                    f"᯽ 𝙺ɪᴅɴᴀᴘ 𝙼ᴇ ᯽", url=f"https://t.me/{app.username}?startgroup=true"
                )
            ]
        ]
    )

    served_users = []
    susers = await get_served_users()
    for user in susers:
        served_users.append(int(user["user_id"]))
    for i in served_users:
        try:
            await app.send_message(i, text=query, reply_markup=BUTTON)
        except FloodWait as e:
            await asyncio.sleep(e.value)
        except:
            pass
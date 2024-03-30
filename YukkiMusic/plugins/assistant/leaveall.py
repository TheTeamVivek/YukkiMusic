import asyncio

from pyrogram import filters
from pyrogram.errors import FloodWait
from pyrogram.types import Message

from YukkiMusic.misc import SUDOERS
from YukkiMusic import app
from YukkiMusic.utils.database import get_client

@app.on_message(filters.command(["leaveall", "assleaveall"]) & filters.user(OWNER_ID))
async def ass_leaveall(_, message: Message):
    for num in assistants:
          client = await get_client(num)
    lear = await client.send_message(f"» sᴛᴀʀᴛᴇᴅ ʟᴇᴀᴠɪɴɢ ᴄʜᴀᴛs...")
    left = 0
    failed = 0
    chats = []
    async for dialog in app2.get_dialogs():
        chats.append(int(dialog.chat.id))
    schat = (await app.get_chat(SUNAME)).id
    for i in chats:
        if i in (-1001686672798, int(schat)):
            continue
        try:
            await app2.leave_chat(int(i))
            left += 1
        except FloodWait as e:
            flood_time = int(e.value)
            if flood_time > 200:
                continue
            await asyncio.sleep(flood_time)
        except Exception:
            continue
            failed += 1
    try:
        await lear.edit_text(
            f"<u>**» {ASS_MENTION} sᴜᴄᴄᴇssғᴜʟʟʏ ʟᴇғᴛ ᴄʜᴀᴛs :**</u>\n\n**ʟᴇғᴛ :** `{left}`\n**ғᴀɪʟᴇᴅ :** `{failed}`"
        )
    except:
        await message.reply_text(
            f"<u>**» {ASS_MENTION} sᴜᴄᴄᴇssғᴜʟʟʏ ʟᴇғᴛ ᴄʜᴀᴛs :**</u>\n\n**ʟᴇғᴛ :** `{left}`\n**ғᴀɪʟᴇᴅ :** `{failed}`"
        )
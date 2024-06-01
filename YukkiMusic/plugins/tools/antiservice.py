from pyrogram import filters

from YukkiMusic import app
from YukkiMusic.utils.permissions import adminsOnly
from YukkiMusic.core.mongo import mongodb
from config import BANNED_USERS
antiservicedb = mongodb.antiservice

async def is_antiservice_on(chat_id: int) -> bool:
    chat = await antiservicedb.find_one({"chat_id": chat_id})
    if not chat:
        return True
    return False


async def antiservice_on(chat_id: int):
    is_antiservice = await is_antiservice_on(chat_id)
    if is_antiservice:
        return
    return await antiservicedb.delete_one({"chat_id": chat_id})


async def antiservice_off(chat_id: int):
    is_antiservice = await is_antiservice_on(chat_id)
    if not is_antiservice:
        return
    return await antiservicedb.insert_one({"chat_id": chat_id})
    

@app.on_message(filters.command(["antiservice", "cleanmode"]) & ~filters.private & ~BANNED_USERS)
@adminsOnly("can_change_info")
async def anti_service(_, message):
    if len(message.command) != 2:
        return await message.reply_text(
            "Usᴀɢᴇ: /antiservice [enable | disable]"
        )
    status = message.text.split(None, 1)[1].strip()
    status = status.lower()
    chat_id = message.chat.id
    if status == "enable":
        await antiservice_on(chat_id)
        await message.reply_text(
            "**Eɴᴀʙʟᴇᴅ AɴᴛɪSᴇʀᴠɪᴄᴇ Sʏsᴛᴇᴍ**.\n I ᴡɪʟʟ Dᴇʟᴇᴛᴇ Sᴇʀᴠɪᴄᴇ Mᴇssᴀɢᴇs ғʀᴏᴍ Nᴏᴡ ᴏɴ."
        )
    elif status == "disable":
        await antiservice_off(chat_id)
        await message.reply_text(
            "**Dɪsᴀʙʟᴇᴅ AɴᴛɪSᴇʀᴠɪᴄᴇ Sʏsᴛᴇᴍ.**\n I ᴡᴏɴ'ᴛ Bᴇ Dᴇʟᴇᴛɪɴɢ Sᴇʀᴠɪᴄᴇ Mᴇssᴀɢᴇ ғʀᴏᴍ Nᴏᴡ ᴏɴ."
        )
    else:
        await message.reply_text(
            "Unknown Suffix, Use /antiservice [enable|disable]"
        )


@app.on_message(filters.service, group=11)
async def delete_service(_, message):
    chat_id = message.chat.id
    try:
        if await is_antiservice_on(chat_id):
            return await message.delete()
    except Exception:
        pass

__MODULE__ = "AɴᴛɪSᴇʀᴠɪᴄᴇ"
__HELP__ = """

Pʟᴜɢɪɴ ᴛᴏ ᴅᴇʟᴇᴛᴇ sᴇʀᴠɪᴄᴇ ᴍᴇssᴀɢᴇs ɪɴ ᴀ ᴄʜᴀᴛ!

/antiservice [ᴇɴᴀʙʟᴇ|ᴅɪsᴀʙʟᴇ]
"""
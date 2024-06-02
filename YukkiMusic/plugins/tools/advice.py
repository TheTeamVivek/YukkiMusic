from pyrogram import filters
from TheApi import api as aapi

from config import LOG_GROUP_ID
from YukkiMusic import api, app


@app.on_message(filters.command("advice"))
async def advice(_, message):
    A = await message.reply_text("...")
    res = aapi.advice()
    await A.edit(res["results"])


@app.on_message(filters.command("astronomical"))
async def advice(_, message):
    a = await api.astronomy()
    if a["success"]:
        c = a["date"]
        url = a["imageUrl"]
        b = a["explanation"]
        caption = f"Tᴏᴅᴀʏ's [{c}] ᴀsᴛʀᴏɴᴏᴍɪᴄᴀʟ ᴇᴠᴇɴᴛ:\n\n{b}"
        await message.reply_photo(url, caption=caption)
    else:
        await message.reply_photo("ᴛʀʏ ᴀғᴛᴇʀ sᴏᴍᴇ ᴛɪᴍᴇ")
        await app.send_message(LOG_GROUP_ID, "/astronomical not working")


__MODULE__ = "Aᴅᴠɪᴄᴇ"
__HELP__ = """
/advice - Gᴇᴛ ʀᴀɴᴅᴏᴍ ᴀᴅᴠɪᴄᴇ
/astronomical - ᴛᴏ ɢᴇᴛ ᴛᴏᴅᴀʏ's ᴀsᴛʀᴏɴᴏᴍɪᴄᴀʟ  ғᴀᴄᴛ"""

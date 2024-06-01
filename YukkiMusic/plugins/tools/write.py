from pyrogram import filters
from TheApi import api

from YukkiMusic import app


@app.on_message(filters.command(["write"]))
async def write(client, message):
    if message.reply_to_message and message.reply_to_message.text:
        txt = message.reply_to_message.text
    elif len(message.command) > 1:
        txt = message.text.split(None, 1)[1]
    else:
        return await message.reply(
            "Pʟᴇᴀsᴇ ʀᴇᴘʟʏ ᴛᴏ ᴍᴇssᴀɢᴇ ᴏʀ ᴡʀɪᴛᴇ ᴀғᴛᴇʀ ᴄᴏᴍᴍᴀɴᴅ ᴛᴏ ᴜsᴇ ᴡʀɪᴛᴇ CMD"
        )
    nan = await message.reply_text("Pʀᴏᴄᴇssɪɴɢ...")
    try:
        img = api.write(txt)
        await message.reply_photo(img)
        await nan.delete()
    except Exception as e:
        await nan.edit(e)

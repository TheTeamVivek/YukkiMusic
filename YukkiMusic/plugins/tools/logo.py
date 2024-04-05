import requests
from YukkiMusic import app
from pyrogram import filters
from pyrogram.types import Message

@app.on_message(filters.command(["logo", "logomaker"], prefixes=["'", '!', '/']))
def handle_logo_command(client, message: Message):
    if len(message.command) != 2:
            return await message.reply_text(
                "**» ɢɪᴠᴇ ᴀ  ɴᴀᴍᴇ ᴛᴏ ᴄʀᴇᴀᴛᴇ ʟᴏɢᴏ..**"
            )
    name = message.text.split(maxsplit=1)[1]
    base_url = "https://logomaker.apinepdev.workers.dev/?logoname={}".format(name)

    try:
        response = requests.get(base_url)
        response.raise_for_status()
        if 'logo' in response.json():
            logo_url = response.json()['logo']
            
            message.reply_photo(logo_url)
        else:
            message.reply_text("Fᴀɪʟᴇᴅ ᴛᴏ ғᴇᴛᴄʜ")
    except Exception as e:
        print(f"Error fetching or sending image: {e}")
        message.reply_text("Fᴀɪʟᴇᴅ ᴛᴏ ғᴇᴛᴄʜ ᴏʀ sᴇɴᴅ ɪᴍᴀɢᴇ.")
        await app.send_message(LOG_GROUP_ID,text=f"ᴇʀʀᴏʀ ᴏɴ logo.py \n ᴇʀʀᴏʀ ɪs {e}")
        
"""
ᴛʜɪs ʙᴏᴛ ᴄᴀɴ ᴄʀᴇᴀᴛᴇ sᴏᴍᴇ ʙᴇᴀᴜᴛɪғᴜʟ ᴀɴᴅ ᴀᴛᴛʀᴀᴄᴛɪᴠᴇ ʟᴏɢᴏ ғᴏʀ ʏᴏᴜʀ ᴘʀᴏғɪʟᴇ ᴘɪᴄs.

❍ /logo (Text) *:* ᴄʀᴇᴀᴛᴇ ᴀ ʟᴏɢᴏ ᴏғ ʏᴏᴜʀ ɢɪᴠᴇɴ ᴛᴇxᴛ ᴡɪᴛʜ ʀᴀɴᴅᴏᴍ ᴠɪᴇᴡ.
"""
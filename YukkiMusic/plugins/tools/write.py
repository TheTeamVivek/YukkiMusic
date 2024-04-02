from pyrogram import Client, filters
import requests
from YukkiMusic import app 

@app.on_message(filters.command(["write","note"]))
def write_text(client, message):
    if len(message.command) < 2:
        message.reply_text("**ᴜsᴀsɢᴇ**\ɴ`/write jai shree ram`")
        return
    text = " ".join(message.command[1:])
    photo_url = "https://apis.xditya.me/write?text=" + text
    app.send_photo(
        chat_id=message.chat.id,
        photo=photo_url,
        caption="ʜᴇʀᴇ ɪs ʏᴏᴜʀ ɴᴏᴛᴇs"
    )
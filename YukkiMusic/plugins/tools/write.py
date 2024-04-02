from pyrogram import Client, filters
import requests
from YukkiMusic import app 

@app.on_message(filters.command(["w","n"]))
def write_text(client, message):
    if len(message.command) < 2:
        message.reply_text("Please provide text after the /write command.\nMust be greater than 2 words")
        return
    text = " ".join(message.command[1:])
    photo_url = "https://apis.xditya.me/write?text=" + text
    app.send_photo(
        chat_id=message.chat.id,
        photo=photo_url,
        caption="Here is the Notes!"
    )
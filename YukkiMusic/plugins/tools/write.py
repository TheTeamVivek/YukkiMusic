from urllib.parse import quote

from pyrogram import Client, filters
from YukkiMusic import app 

@app.on_message(filters.command(["write","note"]))
async def write_text(client, message):
    if len(message.command) < 2:
        await message.reply_text("**Usage**:- `/write jai shree ram`")
        return
    # Join all command arguments into a single string
    text = " ".join(message.command[1:])
    # Encode the text to replace spaces with %20
    encoded_text = quote(text)
    photo_url = "https://apis.xditya.me/write?text=" + encoded_text
    await app.send_photo(
        chat_id=message.chat.id,
        photo=photo_url,
        caption="Here is your note"
    )
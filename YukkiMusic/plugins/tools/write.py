from urllib.parse import quote

from pyrogram import Client, filters
from YukkiMusic import app


@app.on_message(filters.command(["write", "note"]))
async def write_text(client, message):
    if len(message.command) < 2:
        await message.reply_text("**Usage**:- `/write jai shree ram`")
        return

    if message.reply_to_message and message.reply_to_message.text:
        text = message.reply_to_message.text
    else:

        text = " ".join(message.command[1:])

    encoded_text = quote(text)
    photo_url = "https://apis.xditya.me/write?text=" + encoded_text
    await app.send_photo(
        chat_id=message.chat.id, photo=photo_url, caption="Here is your note"
    )

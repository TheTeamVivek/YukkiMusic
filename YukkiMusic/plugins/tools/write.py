from pyrogram import Client, filters
from YukkiMusic import app 

@app.on_message(filters.command(["write","note"]))
async def write_text(client, message):
    if message.reply_to_message:
        text = message.reply_to_message.text
    if len(message.command) < 2 or not message.command[1]:
        await message.reply_text("**Usage**:- `/write <your text>`")
        return
    
    text = " ".join(message.command[1:])
    
    photo_url = "https://apis.xditya.me/write?text=" + text
    
    await app.send_photo(
        chat_id=message.chat.id,
        photo=photo_url,
        caption="Here is your note"
    )
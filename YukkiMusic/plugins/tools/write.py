from urllib.parse import quote
from pyrogram import Client, filters
from YukkiMusic import app 

@app.on_message(filters.command(["write","note"]))
async def write_text(client, message):
    if len(message.command) < 2:
        await message.reply_text("**ᴜsᴀsɢᴇ**:- `/write jai shree ram`")
        return
    user_input = " ".join(message.command[1:])
    text = quote(user_input)
    photo_url = "https://apis.xditya.me/write?text=" + text
    await app.send_photo(
        chat_id=message.chat.id,
        photo=photo_url,
        caption="ʜᴇʀᴇ ɪs ʏᴏᴜʀ ɴᴏᴛᴇs"
    )
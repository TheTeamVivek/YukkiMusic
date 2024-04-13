from pyrogram import filters

from YukkiMusic import app


@app.on_message(filters.command(["qr"]))
async def write_text(client, message):
    if len(message.command) < 2:
        await message.reply_text("**Usage**:- `/qr https://t.me/vivekkumar07089`")
        return
    text = " ".join(message.command[1:])
    photo_url = "https://apis.xditya.me/qr/gen?text=" + text
    await app.send_photo(
        chat_id=message.chat.id, photo=photo_url, caption="Here is your qrcode"
    )

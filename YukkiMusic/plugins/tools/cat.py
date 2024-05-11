import requests
from pyrogram import filters
from pyrogram.types import Message
from YukkiMusic import app
from config import BANNED_USERS

@app.on_message(filters.command("cat") & ~BANNED_USERS)
async def cat(c, m: Message):
    r = requests.get("https://api.thecatapi.com/v1/images/search")
    if r.status_code == 200:
        data = r.json()
        cat_url = data[0]["url"]
        if cat_url.endswith(".gif"):
            await m.reply_animation(cat_url, caption="meow")
        else:
            await m.reply_photo(cat_url, caption="meow")
    else:
        await m.reply_text("Failed to fetch cat picture 🙀")

@app.on_callback_query(filters.regex("refresh_cat") & ~BANNED_USERS)
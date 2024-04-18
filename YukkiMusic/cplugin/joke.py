import requests
from pyrogram import Client, filters
from pyrogram.enums import ParseMode

JOKE_API_ENDPOINT = (
    "https://hindi-jokes-api.onrender.com/jokes?api_key=93eeccc9d663115eba73839b3cd9"
)


@Client.on_message(filters.command("joke"))
async def get_joke(_, message):
    response = requests.get(JOKE_API_ENDPOINT)
    r = response.json()
    joke_text = r["jokeContent"]
    await message.reply_text(joke_text, parse_mode=ParseMode.HTML)

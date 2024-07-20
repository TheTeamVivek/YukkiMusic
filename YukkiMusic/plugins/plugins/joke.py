import requests
from pyrogram import filters
from pyrogram.enums import ParseMode
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from YukkiMusic import app

JOKE_API_ENDPOINT = (
    "https://hindi-jokes-api.onrender.com/jokes?api_key=93eeccc9d663115eba73839b3cd9"
)


@app.on_message(filters.command("joke"))
async def get_joke(_, message):
    response = requests.get(JOKE_API_ENDPOINT)
    r = response.json()
    joke_text = r["jokeContent"]
    refresh_button = InlineKeyboardButton("ʀᴇғʀᴇsʜ", callback_data=f"refresh_joke")
    keyboard = InlineKeyboardMarkup(inline_keyboard=[[refresh_button]])
    await message.reply_text(
        joke_text, reply_markup=keyboard, parse_mode=ParseMode.HTML
    )


@app.on_callback_query(filters.regex(r"refresh_joke"))
async def refresh_joke(_, query):
    await query.answer()
    response = requests.get(JOKE_API_ENDPOINT)
    r = response.json()
    new_joke_text = r["jokeContent"]
    await query.message.edit_text(
        new_joke_text,
        reply_markup=InlineKeyboardMarkup(
            [[InlineKeyboardButton("ʀᴇғʀᴇsʜ", callback_data=f"refresh_joke")]]
        ),
        parse_mode=ParseMode.HTML,
    )

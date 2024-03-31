import requests
import uuid
from  import db
from pyrogram import Client, filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from YukkiMusic import app
from pyrogram.enums import ParseMode

JOKE_API_ENDPOINT = 'https://hindi-jokes-api.onrender.com/jokes?api_key=93eeccc9d663115eba73839b3cd9'

@app.on_message(filters.command("joke"))
async def joke(_, message):
    response = requests.get(JOKE_API_ENDPOINT)
    r = response.json()
    joke_text = r['jokeContent']


    joke_id = str(uuid.uuid4())
    await db.tap_count.insert_one({"joke_id": joke_id, "tap_count": 0, "last_tapped_user": None})

    refresh_button = InlineKeyboardButton("ʀᴇғʀᴇsʜ", callback_data=f"refresh_joke_{joke_id}")
    keyboard = InlineKeyboardMarkup(inline_keyboard=[[refresh_button]])

    await message.reply_text(joke_text, reply_markup=keyboard,parse_mode=ParseMode.HTML)

@app.on_callback_query(filters.regex(r"refresh_joke_(\S+)"))
async def refresh_joke(_, query):

    joke_id = query.data.split("_")[-1]

    joke_record = await db.tap_count.find_one({"joke_id": joke_id})
    if not joke_record:
        return



    await db.tap_count.update_one({"joke_id": joke_id}, {"$inc": {"tap_count": 1}})


    tap_count = joke_record["tap_count"] + 1
    last_tapped_user = query.from_user.mention if query.from_user else "Unknown"

    await db.tap_count.update_one({"joke_id": joke_id}, {"$set": {"last_tapped_user": last_tapped_user}})


    await query.answer()


    response = requests.get(JOKE_API_ENDPOINT)
    r = response.json()
    new_joke_text = r['jokeContent']

    await query.message.edit_text(new_joke_text + f"\n\nᴛᴏᴛᴀʟ ᴛᴀᴘ: {tap_count}\nʟᴀsᴛ ᴛᴀᴘᴘᴇᴅ ʙʏ: {last_tapped_user}", reply_markup=InlineKeyboardMarkup([[InlineKeyboardButton("ʀᴇғʀᴇsʜ", callback_data=f"refresh_joke_{joke_id}")]]),parse_mode=ParseMode.HTML)
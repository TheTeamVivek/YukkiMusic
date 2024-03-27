import asyncio
from pyrogram import Client, filters
from pyrogram.types import (CallbackQuery, InlineKeyboardButton,
                            InlineKeyboardMarkup, InlineQueryResultArticle,
                            InlineQueryResultPhoto, InputTextMessageContent,
                            Message)

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.core.userbot import Userbot

userbot = Userbot()
ASSISTANT_PREFIX = "."


@app.on_message(
    filters.command("pfp", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
)
async def set_pfp(client, message):
    if not message.reply_to_message or not message.reply_to_message.photo:
        return await eor(message, text="Reply to a photo.")
    photo = await message.reply_to_message.download()
    try:
        await userbot.set_profile_photo(photo=photo)
        await eor(message, text="Successfully Changed PFP.")
    except Exception as e:
        await eor(message, text=e)


@app.on_message(
    filters.command("bio", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
)
async def set_bio(client, message):
    if len(message.command) == 1:
        return await eor(message, text="Give some text to set as bio.")
    elif len(message.command) > 1:
        bio = message.text.split(None, 1)[1]
        try:
            await userbot.update_profile(bio=bio)
            await eor(message, text="Changed Bio.")
        except Exception as e:
            await eor(message, text=e)
    else:
        return await eor(message, text="Give some text to set as bio.")

async def vivek():
    await userbot.start()

asyncio.create_task(vivek())
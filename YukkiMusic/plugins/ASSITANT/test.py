from pyrogram import Client, filters
from YukkiMusic import userbot

async def start(client:Client, message):
    await message.reply_text("yeah")


userbot.add_handler(start, filters.command(["start"], prefixes=["."]))

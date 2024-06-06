from pyrogram import Client, filters
from YukkiMusic import userbot

async def start(client:Client, message):
    await message.reply_text("yeah")


start_command_filter = filters.command("start", prefixes=["."])

userbot.add_handler(start, start_command_filter)


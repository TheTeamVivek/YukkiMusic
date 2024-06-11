import os
import shutil

from pyrogram import filters

from YukkiMusic import app, userbot
from YukkiMusic.misc import SUDOERS
from functools import wraps
from pyrogram.handlers import MessageHandler
from pyrogram import filters

def userbot_command(command, prefixes=[".", "/"]):
    def decorator(func):
        @wraps(func)
        async def wrapper(client, message):
            return await func(client, message)
        
        for userbot_client in [userbot.one, userbot.two, userbot.three, userbot.four, userbot.five]:
            if userbot_client:
                userbot_client.add_handler(
                    MessageHandler(wrapper, filters.command(command, prefixes=prefixes))
                )
        return wrapper
    return decorator



@userbot_command("clea")
@app.on_message(filters.command("clea") & SUDOERS)
async def clean(_, message):
    A = await message.reply_text("ᴄʟᴇᴀɴɪɴɢ ᴛᴇᴍᴘ ᴅɪʀᴇᴄᴛᴏʀɪᴇs...")
    dir = "downloads"
    dir1 = "cache"
    shutil.rmtree(dir)
    shutil.rmtree(dir1)
    os.mkdir(dir)
    os.mkdir(dir1)
    await A.edit("ᴛᴇᴍᴘ ᴅɪʀᴇᴄᴛᴏʀɪᴇs ᴀʀᴇ ᴄʟᴇᴀɴᴇᴅ")

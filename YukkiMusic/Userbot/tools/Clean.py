from pyrogram import filters
from .. import userbot_command

@userbot_command("start")
async def start(_ ,m):
    await m.reply_text("i am working")

 
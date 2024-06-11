import os
import shutil

from pyrogram import filters
from typing import Union, List
from YukkiMusic import app, userbot
from YukkiMusic.misc import SUDOERS
from pyrogram import filters

def commandx(commands: Union[str, List[str]]):
    return filters.command(commands, ['.', '!'])


@userbot.one.on_message(commandx("clea"))
async def clean(_, message):
    A = await message.reply_text("ᴄʟᴇᴀɴɪɴɢ ᴛᴇᴍᴘ ᴅɪʀᴇᴄᴛᴏʀɪᴇs...")
    dir = "downloads"
    dir1 = "cache"
    shutil.rmtree(dir)
    shutil.rmtree(dir1)
    os.mkdir(dir)
    os.mkdir(dir1)
    await A.edit("ᴛᴇᴍᴘ ᴅɪʀᴇᴄᴛᴏʀɪᴇs ᴀʀᴇ ᴄʟᴇᴀɴᴇᴅ")

async def h():
    @userbot.one.on_message(commandx("clea"))
    async def clean(_, message):
        A = await message.reply_text("ᴄʟᴇᴀɴɪɴɢ ᴛᴇᴍᴘ ᴅɪʀᴇᴄᴛᴏʀɪᴇs...")
        dir = "downloads"
        dir1 = "cache"
        shutil.rmtree(dir)
        shutil.rmtree(dir1)
        os.mkdir(dir)
        os.mkdir(dir1)
        await A.edit("ᴛᴇᴍᴘ ᴅɪʀᴇᴄᴛᴏʀɪᴇs ᴀʀᴇ ᴄʟᴇᴀɴᴇᴅ")

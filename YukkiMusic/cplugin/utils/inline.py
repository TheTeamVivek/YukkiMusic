from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from YukkiMusic import app
from config import *

pm_buttons = [
    [
        InlineKeyboardButton(
            text="ᴀᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘ",
            url=f"https://t.me/{app.username}?startgroup=true",
        )
    ],
    [InlineKeyboardButton(text="ʜᴇʟᴩ & ᴄᴏᴍᴍᴀɴᴅs", callback_data="clone_help")],
    [
        InlineKeyboardButton(text="❄ ᴄʜᴀɴɴᴇʟ ❄", url=SUPPORT_CHANNEL),
        InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=SUPPORT_GROUP),
    ],
]


gp_buttons = [
    [
        InlineKeyboardButton(
            text="ᴀᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘ",
            url=f"https://t.me/{app.username}?startgroup=true",
        )
    ],
    [
        InlineKeyboardButton(text="❄ ᴄʜᴀɴɴᴇʟ ❄", url=SUPPORT_CHANNEL),
        InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=SUPPORT_GROUP),
    ],
]

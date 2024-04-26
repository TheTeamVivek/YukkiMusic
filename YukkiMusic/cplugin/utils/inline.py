from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from ..start import APP_USERNAME

pm_buttons = [
    [
        InlineKeyboardButton(
            text="ᴀᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘ",
            url=f"https://t.me/{APP_USERNAME}?startgroup=true",
        )
    ],
    [InlineKeyboardButton(text="ʜᴇʟᴩ & ᴄᴏᴍᴍᴀɴᴅs", callback_data="clone_help")],
    [
        InlineKeyboardButton(text="❄ ᴄʜᴀɴɴᴇʟ ❄", url=SUPPORT_CHANNEL),
        InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=SUPPORT_GROUP),
    ],
    [
        InlineKeyboardButton(text="🥀 ᴅᴇᴠᴇʟᴏᴩᴇʀ 🥀", user_id=OWNER_ID),
    ],
]


gp_buttons = [
    [
        InlineKeyboardButton(
            text="ᴀᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘ",
            url=f"https://t.me/{APP_USERNAME}?startgroup=true",
        )
    ],
    [
        InlineKeyboardButton(text="❄ ᴄʜᴀɴɴᴇʟ ❄", url=SUPPORT_CHANNEL),
        InlineKeyboardButton(text="✨ sᴜᴩᴩᴏʀᴛ ✨", url=SUPPORT_GROUP),
    ],
    [
        InlineKeyboardButton(text="🥀 ᴅᴇᴠᴇʟᴏᴩᴇʀ 🥀", user_id=OWNER_ID),
    ],
]

from pyrogram import filters

from config import BANNED_USERS
from pyrogram.types InlineKeyboardMarkup, InlineKeyboardButton
"""from asyncio.exceptions import TimeoutError

try:
    Api_id =  await message.ask("give me you api_id",timeout=2)
    Api_hash =  await message.ask("give me you api_hash")
    print(Api_hash.text)
    print(Api_id.text)
except TimeoutError as e:"""


from config import BANNED_USERS
from YukkiMusic import app

keyboard = InlineKeyboardMarkup(
    [
        [
            InlineKeyboardButton(text="ᴘʏʀᴏɢʀᴀᴍ v1", callback_data="session_pyro1"),
            InlineKeyboardButton(text="ᴘʏʀᴏɢʀᴀᴍ v2", callback_data="session_pyro2"),
        ],
        [InlineKeyboardButton(text="ᴛᴇʟᴇᴛʜᴏɴ", callback_data="session_tele")],
        [
            InlineKeyboardButton(
                text="ᴘʏʀᴏɢʀᴀᴍ ʙᴏᴛ v1", callback_data="session_pyrobot1"
            ),
            InlineKeyboardButton(
                text="ᴘʏʀᴏɢʀᴀᴍ ʙᴏᴛ v2", callback_data="session_pyrobot2"
            ),
        ],
        [InlineKeyboardButton(text="〆 ᴄʟᴏsᴇ 〆", callback_data="close")],
    ]
)


@app.on_message(filters.command("session") & ~BANNED_USERS)
async def session(c, m):
    await m.reply_text(
        "ᴄʜᴏᴏsᴇ ʙᴇʟᴏᴡ ʙᴜᴛᴛᴏɴ ᴛᴏ ɢᴇɴᴇʀᴀᴛᴇ sᴛʀɪɴɢ sᴇssɪᴏɴ", reply_markup=keyboard
    )

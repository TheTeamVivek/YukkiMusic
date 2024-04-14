from pykeyboard import InlineKeyboard
from pyrogram.types import InlineKeyboardButton as Ikb

def ikb(data: dict, row_width: int = 2):
    return keyboard(data.items(), row_width=row_width)
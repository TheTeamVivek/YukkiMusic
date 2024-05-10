from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup


helpmenu = InlineKeyboardMarkup(
            [
                [InlineKeyboardButton(text="ᴘʟᴀʏ", callback_data="clone_cb play")
InlineKeyboardButton(text="ᴛᴇʟᴇɢʀᴀᴘʜ", callback_data="clone_cb telegraph")],
                [
                    InlineKeyboardButton(text="ʙᴀᴄᴋ", callback_data="clone_home"),
                    InlineKeyboardButton(text="ᴄʟᴏsᴇ", callback_data="close"),
                ],
            ],
        )




buttons = InlineKeyboardMarkup(
    [
        [
            InlineKeyboardButton(text="▷", callback_data="resume_cb"),
            InlineKeyboardButton(text="II", callback_data="pause_cb"),
            InlineKeyboardButton(text="‣‣I", callback_data="skip_cb"),
            InlineKeyboardButton(text="▢", callback_data="end_cb"),
        ]
    ]
)

close_key = InlineKeyboardMarkup(
    [[InlineKeyboardButton(text="✯ ᴄʟᴏsᴇ ✯", callback_data="close")]]
)

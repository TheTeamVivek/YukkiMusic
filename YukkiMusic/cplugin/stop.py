from pyrogram import Client, filters
from pyrogram.types import Message

from .play import pytgcalls
from .utils import _clear_, admin_check, close_key


@Client.on_message(filters.command(["stop", "end"]) & filters.group)
@admin_check
async def stop_str(client, message: Message, _):
    try:
        await message.delete()
    except:
        pass
    try:
        await _clear_(message.chat.id)
        await pytgcalls.leave_group_call(message.chat.id)
    except:
        pass

    return await message.reply_text(
        text=f"➻ **sᴛʀᴇᴀᴍ ᴇɴᴅᴇᴅ/sᴛᴏᴩᴩᴇᴅ** ❄\n│ \n└ʙʏ : {message.from_user.mention} 🥀",
        reply_markup=close_key,
    )

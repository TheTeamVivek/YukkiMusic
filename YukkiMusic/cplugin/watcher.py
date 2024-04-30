from pyrogram import Client, filters
from pyrogram.types import Message
from .utils.active import _clear_
from .play import pytgcalls

welcome = 20
close = 30


@Client.on_message(filters.video_chat_started, group=welcome)
@Client.on_message(filters.video_chat_ended, group=close)
async def welcome(_, message: Message):
    try:
        await _clear_(message.chat.id)
        await pytgcalls.leave_group_call(message.chat.id)
    except:
        pass

from pyrogram import Client, filters
from pyrogram.types import Message
from pytgcalls import filters as call_filters
import logging
from pytgcalls.types import Update, ChatUpdate
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


@pytgcalls.on_update(call_filters.chat_update(ChatUpdate.Status.LEFT_CALL))
@pytgcalls.on_update(call_filters.stream_end)
async def handler(client: PyTgCalls, update: Update):
    try:
        await _clear_(update.chat_id)
    except Exception as e:
        logging.exception(e)
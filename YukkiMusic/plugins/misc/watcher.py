from pyrogram import filters
from pyrogram.types import Message

from YukkiMusic import app
from YukkiMusic.core.call import Yukki


@app.on_message(filters.video_chat_started, group=20)
@app.on_message(filters.video_chat_ended, group=30)
async def vc_close_open(_, message: Message):
    await Yukki.force_stop_stream(message.chat.id)

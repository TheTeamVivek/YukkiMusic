import requests
from MukeshAPI import api
from pyrogram import filters
from pyrogram.enums import ChatAction

from YukkiMusic import app

x = None


@app.on_message(filters.command("gemini"))
async def gemini_handler(client, message):
    global x
    await app.send_chat_action(message.chat.id, ChatAction.TYPING)
    if (
        message.text.startswith(f"/gemini@{app.username}")
        and len(message.text.split(" ", 1)) > 1
    ):
        user_input = message.text.split(" ", 1)[1]
    elif message.reply_to_message and message.reply_to_message.text:
        user_input = message.reply_to_message.text
    else:
        if len(message.command) > 1:
            user_input = " ".join(message.command[1:])
        else:
            await message.reply_text("ᴇxᴀᴍᴘʟᴇ :- `/gemini who is lord ram`")
            return

    try:
        response = api.gemini(user_input)
        await app.send_chat_action(message.chat.id, ChatAction.TYPING)
        x = response["results"]
        await message.reply_text(f"{x} ", quote=True)
    except requests.exceptions.RequestException as e:
        pass

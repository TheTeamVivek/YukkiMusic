import requests
from pyrogram import filters
from pyrogram.enums import ChatAction

from YukkiMusic import app


@app.on_message(
    filters.command(
        ["chatgpt", "ai", "ask"], prefixes=["+", ".", "/", "-", "?", "$", "#", "&"]
    )
)
async def chatgpt_chat(bot, message):
    if len(message.command) < 2 and not message.reply_to_message:
        await message.reply_text(
            "Example:\n\n`/ai write simple website code using html css, js?`"
        )
        return

    if message.reply_to_message and message.reply_to_message.text:
        user_input = message.reply_to_message.text
    else:
        user_input = " ".join(message.command[1:])

    try:
        response = requests.get(
            f"https://chatgpt.apinepdev.workers.dev/?question={user_input}"
        )
        if response.status_code == 200:
            await bot.send_chat_action(message.chat.id, ChatAction.TYPING)
            result = response.json()["answer"]
            await message.reply_text(f"{result}", quote=True)
        else:
            pass
    except requests.exceptions.RequestException as e:
        pass

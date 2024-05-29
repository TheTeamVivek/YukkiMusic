from YukkiMusic import userbot
from pyrogram import filters, Client

@userbot.one.on_message(filters.text)
def handle_message(client, message):
    message.reply_text("Hello, world!")
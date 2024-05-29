from pyrogram import filters

from YukkiMusic import userbot


@userbot.one.on_message(filters.text)
def handle_message(client, message):
    message.reply_text("Hello, world!")

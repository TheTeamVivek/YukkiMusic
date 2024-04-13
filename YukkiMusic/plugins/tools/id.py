from pyrogram import filters

from YukkiMusic import app


# Define a handler for /id command
@app.on_message(filters.command("i", prefixes="/"))
async def get_id(client, message):
    # Check if the message is sent in a group
    if message.chat.type == "group" or message.chat.type == "supergroup":
        await message.reply_text(f"The ID of this group is: {message.chat.id}")
    # Check if the message is sent in a private chat
    elif message.chat.type == "private":
        await message.reply_text(f"Your ID is: {message.from_user.id}")

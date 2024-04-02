from pyrogram import Client, filters
import requests
from YukkiMusic import app 

@app.on_message(filters.command(["w","n"]))
def write_text(client, message):
    # Check if there is text following the command
    if len(message.command) < 1:
        # Reply to the user asking to provide text
        message.reply_text("Please provide text after the /write command.")
        return
    
    # Extract the text after /write command
    text = " ".join(message.command[1:])
    
    # Replace <text> in the URL with the user input message
    photo_url = "https://apis.xditya.me/write?text=" + text
    
    app.send_photo(
        chat_id=message.chat.id,
        photo=photo_url,
        caption="Here is the Notes!"
    )
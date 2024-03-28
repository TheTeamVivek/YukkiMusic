import os
from inspect import getfullargspec
from pyrogram import Client, filters
from pyrogram.types import InputPhoto

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import get_client

ASSISTANT_PREFIX = "."


@app.on_message(
    filters.command("pfp", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
)
async def set_pfp(client, message):
    from YukkiMusic.core.userbot import assistants
    if not message.reply_to_message or not message.reply_to_message.photo:
        return await message.reply_text("Reply to a photo.")
    
    try:
        for num in assistants:
            assistant_client = await Client(num)
            await assistant_client.start()
            # Retrieve profile photos
            profile_photos = await assistant_client.get_profile_photos("me")
            # Delete each profile photo
            for photo in profile_photos:
                await assistant_client.delete_profile_photos(photo.file_id)
            
            photo = await message.reply_to_message.download()
            await assistant_client.set_profile_photo(photo=InputPhoto(photo))
            await message.reply_text("Successfully Changed PFP.")
            os.remove(photo)
            await assistant_client.stop()
    
    except Exception as e:
        await message.reply_text(str(e))

@app.on_message(
    filters.command("bio", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
)
async def set_bio(client, message):
    if len(message.command) == 1:
        return await eor(message, text="Give some text to set as bio.")
    elif len(message.command) > 1:
        userbot = await get_client(1)
        bio = message.text.split(None, 1)[1]
        try:
            await userbot.update_profile(bio=bio)
            await eor(message, text="Changed Bio.")
        except Exception as e:
            await eor(message, text=e)
    else:
        return await eor(message, text="Give some text to set as bio.")

async def eor(msg: Message, **kwargs):
    func = (
        (msg.edit_text if msg.from_user.is_self else msg.reply)
        if msg.from_user
        else msg.reply
    )
    spec = getfullargspec(func.__wrapped__).args
    return await func(**{k: v for k, v in kwargs.items() if k in spec})
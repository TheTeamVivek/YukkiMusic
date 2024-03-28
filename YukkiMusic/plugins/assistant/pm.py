import asyncio
from YukkiMusic.core.userbot import assistants
from YukkiMusic.misc import SUDOERS
from pyrogram import filters
from pyrogram.types import Message
from inspect import getfullargspec


ASSISTANT_PREFIX="."


async def initialize_clients():
    client = await get_client(num)
    return client

async def main():
    client = await initialize_clients()
    return client

@client.on_message(
    filters.command("setname", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
)
async def set_bio(client, message):
    from YukkiMusic.core.userbot import assistants
    if len(message.command) == 1:
        return await eor(message, text="Give some text to set as name.")
    elif len(message.command) > 1:
        for num in assistants:
            client = await get_client(num)
            name = message.text.split(None, 1)[1]
        try:
            await client.update_profile(first_name=name)
            await eor(message, text=f"name Changed to {name} .")
        except Exception as e:
            await eor(message, text=e)
    else:
        return await eor(message, text="Give some text to set as name.")


async def eor(msg: Message, **kwargs):
    func = (
        (msg.edit_text if msg.from_user.is_self else msg.reply)
        if msg.from_user
        else msg.reply
    )
    spec = getfullargspec(func.__wrapped__).args
    return await func(**{k: v for k, v in kwargs.items() if k in spec})


asyncio.create_task(main())

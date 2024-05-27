import asyncio
import logging
import random
import re

from pyrogram import Client, filters
from pyrogram.raw.functions.messages import DeleteHistory
from pyrogram.types import Message

from YukkiMusic import app
from YukkiMusic.core.userbot import assistants
from YukkiMusic.utils.database import get_client


@app.on_message(filters.command(["sg", "History"]))
async def sg(client: Client, message: Message):
    if len(message.text.split()) < 2 and not message.reply_to_message:
        return await message.reply("sg username/id/reply")
    if message.reply_to_message:
        args = message.reply_to_message.from_user.id
    else:
        args = message.text.split()[1:]
        if not args:
            return await message.reply(
                "Please provide a username, ID, or reply to a message."
            )
        args = args[0]
    lol = await message.reply("<code>Processing...</code>")
    if args:
        try:
            user = await client.get_users(f"{args}")
        except Exception:
            return await lol.edit("<code>Please specify a valid user!</code>")
    bo = ["sangmata_bot", "sangmata_beta_bot"]
    sg = random.choice(bo)
    aj = random.choice(assistants)
    ubot = await get_client(aj)

    try:
        a = await ubot.send_message(sg, f"{user.id}")
        await a.delete()
    except Exception as e:
        return await lol.edit(str(e))
    await asyncio.sleep(1)

    async for stalk in ubot.search_messages(a.chat.id):
        if stalk.text is None:
            continue
        if not stalk:
            await message.reply("The bot encountered an issue. Please try again later")
        elif stalk:
            await message.reply(f"{stalk.text}")
            break

    try:
        user_info = await ubot.resolve_peer(sg)
        await ubot.send(DeleteHistory(peer=user_info, max_id=0, revoke=True))
    except Exception:
        pass

    await lol.delete()


@app.on_message(filters.command(["truecaller"]))
async def sg(client: Client, message: Message):
    if len(message.text.split()) < 2 and not message.reply_to_message:
        return await message.reply("sg username/id/reply")
    if message.reply_to_message and message.reply_to_message.text:
        input = message.reply_to_message.text
    else:
        input = " ".join(message.command[1:])

    if not re.match(r"^\+[1-9]\d{1,14}$", input):
        return await message.reply_text(
            "Please provide a valid phone number including country code."
        )

    lol = await message.reply("<code>Processing...</code>")
    tr = ["@TrueCaller_Z_Bot"]
    cli = random.choice(assistants)
    ubot = await get_client(cli)

    try:
        a = await ubot.send_message(tr, f"{input}")
        await a.delete()
    except Exception as e:
        return await lol.edit(str(e))
    await asyncio.sleep(1)

    async for stalk in ubot.search_messages(a.chat.id):
        if stalk.text is None:
            continue
        if not stalk:
            await message.reply("The bot encountered an issue. Please try again later")
        elif stalk:
            await message.reply(f"{stalk.text}")
            break

    try:
        user_info = await ubot.resolve_peer(sg)
        await ubot.send(DeleteHistory(peer=user_info, max_id=0, revoke=True))
    except Exception as e:
        logging.exception(e)

    await lol.delete()

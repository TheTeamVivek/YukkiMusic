from io import BytesIO
from traceback import format_exc

from pyrogram import filters
from pyrogram.types import Message

from config import BANNED_USERS
from YukkiMusic import app, arq
from YukkiMusic.utils.error import capture_err

__MODULE__ = "Quotly"
__HELP__ = """
/q - To quote a message.
/q [INTEGER] - To quote more than 1 messages.
/q r - to quote a message with it's reply

"""


async def quotify(messages: list):
    response = await arq.quotly(messages)
    if not response.ok:
        return [False, response.result]
    sticker = response.result
    sticker = BytesIO(sticker)
    sticker.name = "sticker.webp"
    return [True, sticker]


def getArg(message: Message) -> str:
    arg = message.text.strip().split(None, 1)[1].strip()
    return arg


def isArgInt(message: Message) -> list:
    count = getArg(message)
    try:
        count = int(count)
        return [True, count]
    except ValueError:
        return [False, 0]


@app.on_message(filters.command("q") & ~filters.group & ~BANNED_USERS)
async def song_commad_group(client, message: Message, _):
    await message.reply_text("buddy you can use this command in my pm/dm")


@app.on_message(filters.command("q") & ~filters.private & ~BANNED_USERS)
@capture_err
async def quotly_func(client, message: Message):
    if not message.reply_to_message:
        return await message.reply_text("Reply to a message to quote it.")
    if not message.reply_to_message.text:
        return await message.reply_text("Replied message has no text, can't quote it.")
    m = await message.reply_text("Quoting Messages")
    if len(message.command) < 2:
        messages = [message.reply_to_message]

    elif len(message.command) == 2:
        arg = isArgInt(message)
        if arg[0]:
            if arg[1] < 2 or arg[1] > 10:
                return await m.edit("Argument must be between 2-10.")

            count = arg[1]

            # Fetching 5 extra messages so that we can ignore media
            # messages and still end up with correct offset
            messages = [
                i
                for i in await client.get_messages(
                    message.chat.id,
                    range(
                        message.reply_to_message.id,
                        message.reply_to_message.id + (count + 5),
                    ),
                    replies=0,
                )
                if not i.empty and not i.media
            ]
            messages = messages[:count]
        else:
            if getArg(message) != "r":
                return await m.edit(
                    "Incorrect Argument, Pass **'r'** or **'INT'**, **EX:** __/q 2__"
                )
            reply_message = await client.get_messages(
                message.chat.id,
                message.reply_to_message.id,
                replies=1,
            )
            messages = [reply_message]
    else:
        return await m.edit("Incorrect argument, check quotly module in help section.")
    try:
        if not message:
            return await m.edit("Something went wrong.")

        sticker = await quotify(messages)
        if not sticker[0]:
            await message.reply_text(sticker[1])
            return await m.delete()
        sticker = sticker[1]
        await message.reply_sticker(sticker)
        await m.delete()
        sticker.close()
    except Exception as e:
        await m.edit(
            "Something went wrong while quoting messages,"
            + " This error usually happens when there's a "
            + " message containing something other than text,"
            + " or one of the messages in-between are deleted."
        )
        e = format_exc()
        print(e)

import asyncio
from inspect import getfullargspec
from pyrogram import Client, filters
from pyrogram.types import Message

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (approve_pmpermit, disapprove_pmpermit, is_on_off,
                            is_pmpermit_approved)
from config import ASSISTANT_PREFIX, LOG_GROUP_ID
from YukkiMusic.core.userbot import Userbot

class CustomUserbot(Userbot):
    async def on_message(self, message: Message):
        if message.chat.type == "private" and message.from_user.id in SUDOERS:
            await self.awaiting_message(message)
        await super().on_message(message)

    async def awaiting_message(self, message: Message):
        if await is_on_off(5):
            try:
                await self.forward_messages(
                    chat_id=LOG_GROUP_ID,
                    from_chat_id=message.from_user.id,
                    message_ids=message.message_id,
                )
            except Exception as err:
                pass
        user_id = message.from_user.id
        if await is_pmpermit_approved(user_id):
            return
        async for m in self.iter_history(user_id, limit=6):
            if m.reply_markup:
                await m.delete()
        flood_key = str(user_id)
        if flood_key in flood:
            flood[flood_key] += 1
        else:
            flood[flood_key] = 1
        if flood[flood_key] > 5:
            await message.reply_text("Spam Detected. User Blocked")
            await self.send_message(
                LOG_GROUP_ID,
                f"**Spam Detected, User Blocked**\n\n- **Blocked User:** {message.from_user.mention}\n- **User ID:** {message.from_user.id}",
            )
            return await self.block_user(user_id)
        await message.reply_text(
            f"Hello, I am {app.mention}'s Assistant.\n\nPlease don't spam here, else you'll get blocked.\nFor more help, start with: {app.mention}"
        )

    async def eor(self, msg: Message, **kwargs):
        func = (
            (msg.edit_text if msg.from_user.is_self else msg.reply)
            if msg.from_user
            else msg.reply
        )
        spec = getfullargspec(func.__wrapped__).args
        return await func(**{k: v for k, v in kwargs.items() if k in spec})

    async def start(self):
        await super().start()

userbot = CustomUserbot()

@userbot.on_message(
    filters.command("approve", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
    & ~filters.via_bot
)
@userbot.on_message(
    filters.command("approve", prefixes=ASSISTANT_PREFIX)
    & filters.user("me")
    & ~filters.via_bot
)
async def pm_approve(client, message):
    if not message.reply_to_message:
        return await userbot.eor(
            message, text="Reply to a user's message to approve."
        )
    user_id = message.reply_to_message.from_user.id
    if await is_pmpermit_approved(user_id):
        return await userbot.eor(message, text="User is already approved to PM.")
    await approve_pmpermit(user_id)
    await userbot.eor(message, text="User is approved to PM.")

@userbot.on_message(
    filters.command("disapprove", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
    & ~filters.via_bot
)
@userbot.on_message(
    filters.command("disapprove", prefixes=ASSISTANT_PREFIX)
    & filters.user("me")
    & ~filters.via_bot
)
async def pm_disapprove(client, message):
    if not message.reply_to_message:
        return await userbot.eor(
            message, text="Reply to a user's message to disapprove."
        )
    user_id = message.reply_to_message.from_user.id
    if not await is_pmpermit_approved(user_id):
        await userbot.eor(message, text="User is already disapproved to PM.")
        async for m in userbot.iter_history(user_id, limit=6):
            if m.reply_markup:
                try:
                    await m.delete()
                except Exception:
                    pass
        return
    await disapprove_pmpermit(user_id)
    await userbot.eor(message, text="User is disapproved to PM.")

@userbot.on_message(
    filters.command("block", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
    & ~filters.via_bot
)
@userbot.on_message(
    filters.command("block", prefixes=ASSISTANT_PREFIX)
    & filters.user("me")
    & ~filters.via_bot
)
async def block_user_func(client, message):
    if not message.reply_to_message:
        return await userbot.eor(message, text="Reply to a user's message to block.")
    user_id = message.reply_to_message.from_user.id
    await userbot.eor(message, text="Successfully blocked the user")
    await userbot.block_user(user_id)

@userbot.on_message(
    filters.command("unblock", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
    & ~filters.via_bot
)
@userbot.on_message(
    filters.command("unblock", prefixes=ASSISTANT_PREFIX)
    & filters.user("me")
    & ~filters.via_bot
)
async def unblock_user_func(client, message):
    if not message.reply_to_message:
        return await userbot.eor(
            message, text="Reply to a user's message to unblock."
        )
    user_id = message.reply_to_message.from_user.id
    await userbot.unblock_user(user_id)
    await userbot.eor(message, text="Successfully unblocked the user")

@userbot.on_message(
    filters.command("pfp", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
    & ~filters.via_bot
)
@userbot.on_message(
    filters.command("pfp", prefixes=ASSISTANT_PREFIX)
    & filters.user("me")
    & ~filters.via_bot
)
async def set_pfp(client, message):
    if not message.reply_to_message or not message.reply_to_message.photo:
        return await userbot.eor(message, text="Reply to a photo.")
    photo = await message.reply_to_message.download()
    try:
        await userbot.set_profile_photo(photo=photo)
        await userbot.eor(message, text="Successfully Changed PFP.")
    except Exception as e:
        await userbot.eor(message, text=str(e))

@userbot.on_message(
    filters.command("bio", prefixes=ASSISTANT_PREFIX)
    & SUDOERS
    & ~filters.via_bot
)
@userbot.on_message(
    filters.command("bio", prefixes=ASSISTANT_PREFIX)
    & filters.user("me")
    & ~filters.via_bot
)
async def set_bio(client, message):
    if len(message.command) == 1:
        return await userbot.eor(message, text="Give some text to set as bio.")
    elif len(message.command) > 1:
        bio = message.text.split(None, 1)[1]
        try:
            await userbot.update_profile(bio=bio)
            await userbot.eor(message, text="Changed Bio.")
        except Exception as e:
            await userbot.eor(message, text=str(e))
    else:
        return await userbot.eor(message, text="Give some text to set as bio.")


async def main():
    await userbot.start()

asyncio.run(main())
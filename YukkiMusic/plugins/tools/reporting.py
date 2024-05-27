"""from pyrogram import filters
from YukkiMusic import app
from YukkiMusic.utils.error import capture_err
from pyrogram.enums import ChatMemberStatus, ChatMembersFilter
from config import adminlist
import logging


@app.on_message(
    (filters.command("report") | filters.command(["admins", "admin"], prefixes="@"))
    & ~filters.private
)
@capture_err
async def report_user(_, message):
    try:
        a = await app.get_chat_member(message.chat.id, app.id)
        if (
            a.status == ChatMemberStatus.ADMINISTRATOR
            or a.status == ChatMemberStatus.OWNER
        ):
            return
    except Exception as e:
        logging.exception(e)

    if len(message.text.split()) <= 1 and not message.reply_to_message:
        return await message.reply_text("Reply to a message to report that user.")

    reply = message.reply_to_message if message.reply_to_message else message
    reply_id = reply.from_user.id if reply.from_user else reply.sender_chat.id
    user_id = message.from_user.id if message.from_user else message.sender_chat.id

    list_of_admins = adminlist.get(message.chat.id)
    linked_chat = (await app.get_chat(message.chat.id)).linked_chat
    if linked_chat is not None:
        if (
            reply_id in list_of_admins
            or reply_id == message.chat.id
            or reply_id == linked_chat.id
        ):
            return await message.reply_text(
                "Do you know that the user you are replying is an admin ?"
            )
    else:
        if reply_id in list_of_admins or reply_id == message.chat.id:
            return await message.reply_text(
                "Do you know that the user you are replying is an admin ?"
            )

    user_mention = (
        reply.from_user.mention if reply.from_user else reply.sender_chat.title
    )
    text = f"Reported {user_mention} to admins!."
    admin_data = [
        i
        async for i in app.get_chat_members(
            chat_id=message.chat.id, filter=ChatMembersFilter.ADMINISTRATORS
        )
    ]  # will it give floods ???
    for admin in admin_data:
        if admin.user.is_bot or admin.user.is_deleted:
            # return bots or deleted admins
            continue
        text += f"[\u2063](tg://user?id={admin.user.id})"

    await reply.reply_text(text)"""

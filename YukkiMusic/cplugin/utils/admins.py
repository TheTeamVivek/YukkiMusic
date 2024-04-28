from typing import Callable

from pyrogram import Client
from pyrogram.enums import ChatMemberStatus
from pyrogram.types import CallbackQuery, Message

from YukkiMusic.misc import SUDOERS

from .active import is_active_chat

def admin_check(func: Callable) -> Callable:
    async def non_admin(_, message: Message):
        if not await is_active_chat(message.chat.id):
            return await message.reply_text("ʙᴏᴛ ɪsɴ'ᴛ sᴛʀᴇᴀᴍɪɴɢ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ.")

        if message.from_user.id in SUDOERS:
            return await func(_, message)

        check = await app.get_chat_member(message.chat.id, message.from_user.id)
        if check.status not in [ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR]:
            return await message.reply_text(
                "» ʏᴏᴜ'ʀᴇ ɴᴏᴛ ᴀɴ ᴀᴅᴍɪɴ ʙᴀʙʏ, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs."
            )

        admin = (
            await app.get_chat_member(message.chat.id, message.from_user.id)
        ).privileges
        if admin.can_manage_video_chats:
            return await func(_, message)
        else:
            return await message.reply_text(
                "» ʏᴏᴜ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ᴘᴇʀᴍɪssɪᴏɴs ᴛᴏ ᴍᴀɴᴀɢᴇ ᴠɪᴅᴇᴏᴄʜᴀᴛs, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs."
            )

    return non_admin



def admin_check_cb(func: Callable) -> Callable:
    async def cb_non_admin(client, query: CallbackQuery, _) -> None:
        if not await is_active_chat(query.message.chat.id):
            return await query.answer(
                "ʙᴏᴛ ɪsɴ'ᴛ sᴛʀᴇᴀᴍɪɴɢ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ.", show_alert=True
            )

        if query.from_user.id in SUDOERS:
            return await func(client, query, _)

        try:
            check = await client.get_chat_member(
                query.message.chat.id, query.from_user.id
            )
        except:
            return
        if check.status not in [ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR]:
            return await query.answer(
                "» ʏᴏᴜ'ʀᴇ ɴᴏᴛ ᴀɴ ᴀᴅᴍɪɴ ʙᴀʙʏ, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs.",
                show_alert=True,
            )

        admin = (
            await client.get_chat_member(query.message.chat.id, query.from_user.id)
        ).privileges
        if admin.can_manage_video_chats:
            return await func(client, query, _)
        else:
            return await query.answer(
                "» ʏᴏᴜ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ᴘᴇʀᴍɪssɪᴏɴs ᴛᴏ ᴍᴀɴᴀɢᴇ ᴠɪᴅᴇᴏᴄʜᴀᴛs, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs.",
                show_alert=True,
            )

    return cb_non_admin

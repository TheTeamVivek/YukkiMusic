from pyrogram import Client, filters
from pyrogram.types import Message
from pyrogram.enums import ChatMemberStatus, ChatType
from YukkiMusic import app

BOT_ID = app.id


@Client.on_message(filters.new_chat_members)
@Client.on_message(filters.new_chat_members, group=-1)
async def welcome(client: Client, message: Message):
    i = await client.get_me()
    if chat_type != ChatType.SUPERGROUP:
        await message.reply_text(
            "ᴘʟᴇᴀsᴇ ᴄᴏɴᴠᴇʀᴛ ʏᴏᴜʀ ɢʀᴏᴜᴘ ᴛᴏ ᴀ sᴜᴘᴇʀɢʀᴏᴜᴘ ᴏʀ ᴍᴀᴋᴇ ʏᴏᴜʀ ɢʀᴏᴜᴘ ʜɪsᴛᴏʀʏ ᴠɪɪsɪʙʟᴇ sᴏ ɪ ᴡᴏʀᴋ ᴘᴇʀғᴇᴄᴛʟʏ"
        )
        return await client.leave_chat(message.chat.id)
    a = await client.get_chat_member(message.chat.id, i.id)
    if a.status != ChatMemberStatus.ADMINISTRATOR:
        await message.reply_text(
            "ᴘʟᴇᴀsᴇ ᴍᴀᴋᴇ ᴍᴇ ᴀɴ ᴀᴅᴍɪɴ ᴡɪᴛʜ **ɪɴᴠɪᴛᴇ ᴜsᴇʀ** ᴘᴇʀᴍɪssɪᴏɴ ᴛᴏ ᴘʟᴀʏ ᴍᴜsɪᴄ"
        )
        return await client.leave_chat(message.chat.id)
    try:
        b = await client.get_chat_member(message.chat.id, BOT_ID)
        if b.status in [
            ChatMemberStatus.OWNER,
            ChatMemberStatus.ADMINISTRATOR,
            ChatMemberStatus.MEMBER,
        ]:
            await message.reply_text(
                "sᴏʀʀʏ! ᴍʏ ᴍᴀsᴛᴇʀ ʙᴏᴛ ɪs ᴀʟʀᴇᴀᴅʏ ʜᴇʀᴇ sᴏ ɪ ᴀᴍ ʟᴇᴀᴠᴇɪɴɢ"
            )
            return await client.leave_chat(message.chat.id)
    except Exception:
        pass


@Client.on_message(filters.video_chat_started, group=20)
@Client.on_message(filters.video_chat_ended, group=30)
@Client.on_message(filters.new_chat_members)
@Client.on_message()
async def leave(client: Client, message: Message):
    if (
        message.from_user
        and message.from_user.id == BOT_ID
        or message.from_user
        and message.from_user.id == CLONES
    ):
        await client.send_message(
            message.chat.id,
            "My Friend bot aur master bot is here so I can't stay here anymore thankyou",
        )
        await client.leave_chat(message.chat.id)

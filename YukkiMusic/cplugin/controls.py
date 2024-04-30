import logging
from pytgcalls.types import MediaStream, AudioQuality


from pyrogram import Client, filters
from pyrogram.types import Message
from pyrogram.enums import ChatMemberStatus
from .play import pytgcalls
from .utils import (
    close_key,
    is_streaming,
    stream_off,
    stream_on,
    is_active_chat,
)
from .utils.active import _clear_
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.thumbnails import gen_thumb
from YukkiMusic.misc import clonedb


@Client.on_message(filters.command(["pause", "resume", "end", "stop"]) & filters.group)
async def pause_str(client, message: Message):
    id = await client.get_me()
    try:
        await message.delete()
    except:
        pass
    if not await is_active_chat(message.chat.id, id.id):
        return await message.reply_text("ʙᴏᴛ ɪsɴ'ᴛ sᴛʀᴇᴀᴍɪɴɢ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ.")
    check = await client.get_chat_member(message.chat.id, message.from_user.id)

    if (
        check.status not in [ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR]
    ) or message.from_user.id not in SUDOERS:

        return await message.reply_text(
            "» ʏᴏᴜ'ʀᴇ ɴᴏᴛ ᴀɴ ᴀᴅᴍɪɴ ʙᴀʙʏ, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs."
        )

    admin = (
        await client.get_chat_member(message.chat.id, message.from_user.id)
    ).privileges
    if not admin.can_manage_video_chats:
        return await message.reply_text(
            "» ʏᴏᴜ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ᴘᴇʀᴍɪssɪᴏɴs ᴛᴏ ᴍᴀɴᴀɢᴇ ᴠɪᴅᴇᴏᴄʜᴀᴛs, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs."
        )
    if message.text.lower() == "/pause":
        if not await is_streaming(message.chat.id, id.id):
            return await message.reply_text(
                "ᴅɪᴅ ʏᴏᴜ ʀᴇᴍᴇᴍʙᴇʀ ᴛʜᴀᴛ ʏᴏᴜ ʀᴇsᴜᴍᴇᴅ ᴛʜᴇ sᴛʀᴇᴀᴍ ?"
            )
        await pytgcalls.pause_stream(message.chat.id)
        await stream_off(message.chat.id, id.id)
        return await message.reply_text(
            text=f"➻ sᴛʀᴇᴀᴍ ᴩᴀᴜsᴇᴅ 🥺\n└ʙʏ : {message.from_user.mention} 🥀",
        )
    elif message.text.lower() == "/resume":

        if await is_streaming(message.chat.id, id.id):
            return await message.reply_text(
                "ᴅɪᴅ ʏᴏᴜ ʀᴇᴍᴇᴍʙᴇʀ ᴛʜᴀᴛ ʏᴏᴜ ᴘᴀᴜsᴇᴅ ᴛʜᴇ sᴛʀᴇᴀᴍ ?"
            )
        await stream_on(message.chat.id, id.id)
        await pytgcalls.resume_stream(message.chat.id)
        return await message.reply_text(
            text=f"➻ sᴛʀᴇᴀᴍ ʀᴇsᴜᴍᴇᴅ 💫\n│ \n└ʙʏ : {message.from_user.mention} 🥀",
        )
    elif message.text.lower() == "/end" or message.text.lower() == "/stop":
        try:
            await _clear_(message.chat.id, id.id)
            await pytgcalls.leave_group_call(message.chat.id)
        except:
            pass

        return await message.reply_text(
            text=f"➻ **sᴛʀᴇᴀᴍ ᴇɴᴅᴇᴅ/sᴛᴏᴩᴩᴇᴅ** ❄\n│ \n└ʙʏ : {message.from_user.mention} 🥀",
        )


@Client.on_message(filters.command(["skip", "next"]) & filters.group)
async def skip_str(client: Client, message: Message):
    i = await client.get_me()
    try:
        await message.delete()
    except:
        pass
    if not await is_active_chat(message.chat.id, i.id):
        return await message.reply_text("ʙᴏᴛ ɪsɴ'ᴛ sᴛʀᴇᴀᴍɪɴɢ ᴏɴ ᴠɪᴅᴇᴏᴄʜᴀᴛ.")
    check = await client.get_chat_member(message.chat.id, message.from_user.id)

    if (
        check.status not in [ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR]
    ) or message.from_user.id not in SUDOERS:

        return await message.reply_text(
            "» ʏᴏᴜ'ʀᴇ ɴᴏᴛ ᴀɴ ᴀᴅᴍɪɴ ʙᴀʙʏ, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs."
        )

    admin = (
        await client.get_chat_member(message.chat.id, message.from_user.id)
    ).privileges
    if not admin.can_manage_video_chats:
        return await message.reply_text(
            "» ʏᴏᴜ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ᴘᴇʀᴍɪssɪᴏɴs ᴛᴏ ᴍᴀɴᴀɢᴇ ᴠɪᴅᴇᴏᴄʜᴀᴛs, ᴘʟᴇᴀsᴇ sᴛᴀʏ ɪɴ ʏᴏᴜʀ ʟɪᴍɪᴛs."
        )
    get = clonedb.get((message.chat.id, i.id))
    if not get:
        try:
            await _clear_(message.chat.id, i.id)
            await pytgcalls.leave_group_call(message.chat.id)
            await message.reply_text(
                text=f"➻ sᴛʀᴇᴀᴍ sᴋɪᴩᴩᴇᴅ 🥺\n│ \n└ʙʏ : {message.from_user.mention} 🥀\n\n**» ɴᴏ ᴍᴏʀᴇ ǫᴜᴇᴜᴇᴅ ᴛʀᴀᴄᴋs ɪɴ** {message.chat.title}, **ʟᴇᴀᴠɪɴɢ ᴠɪᴅᴇᴏᴄʜᴀᴛ.**",
            )
        except:
            return
    else:
        title = get[0]["title"]
        duration = get[0]["duration"]
        file_path = get[0]["file_path"]
        videoid = get[0]["videoid"]
        req_by = get[0]["req"]
        user_id = get[0]["user_id"]
        get.pop(0)

        stream = MediaStream(file_path, audio_parameters=AudioQuality.HIGH)
        try:
            await pytgcalls.change_stream(
                message.chat.id,
                stream,
            )
        except:
            await _clear_(message.chat.id, i.id)
            return await pytgcalls.leave_group_call(message.chat.id)

        await message.reply_text(
            text=f"➻ sᴛʀᴇᴀᴍ sᴋɪᴩᴩᴇᴅ 🥺\n│ \n└ʙʏ : {message.from_user.mention} 🥀",
        )
        img = await gen_thumb(videoid)
        return await message.reply_photo(
            photo=img,
            caption=f"**➻ sᴛᴀʀᴛᴇᴅ sᴛʀᴇᴀᴍɪɴɢ**\n\n‣ **ᴛɪᴛʟᴇ :** [{title[:27]}](https://t.me/{i.username}?start=info_{videoid})\n‣ **ᴅᴜʀᴀᴛɪᴏɴ :** `{duration}` ᴍɪɴᴜᴛᴇs\n‣ **ʀᴇǫᴜᴇsᴛᴇᴅ ʙʏ :** {req_by}",
        )

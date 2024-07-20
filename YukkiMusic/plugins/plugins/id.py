from pyrogram import filters
from YukkiMusic import app


@app.on_message(filters.command("id"))
async def get_id(client, message):
    try:
        if not message.reply_to_message and message.chat:
            await message.reply(
                f"ᴜsᴇʀ <b>{message.from_user.first_name}'s</b> ɪᴅ ɪs <code>{message.from_user.id}</code>.\nᴛʜɪs ᴄʜᴀᴛ's ɪᴅ ɪs: <code>{message.chat.id}</code>."
            )
        elif not message.reply_to_message.sticker or message.reply_to_message is None:
            if message.reply_to_message.forward_from_chat:
                await message.reply(
                    f"Tʜᴇ ғᴏʀᴡᴀʀᴅᴇᴅ {str(message.reply_to_message.forward_from_chat.type)[9:].lower()}, {message.reply_to_message.forward_from_chat.title} ʜᴀs ᴀɴ ID ᴏғ <code>{message.reply_to_message.forward_from_chat.id}</code>"
                )

            elif message.reply_to_message.forward_from:
                await message.reply(
                    f"Tʜᴇ ғᴏʀᴡᴀʀᴅᴇᴅ ᴜsᴇʀ, {message.reply_to_message.forward_from.first_name} ʜᴀs ᴀɴ ID ᴏғ <code>{message.reply_to_message.forward_from.id}</code>."
                )

            elif message.reply_to_message.forward_sender_name:
                await message.reply(
                    "Sᴏʀʀʏ, I ɴᴇᴠᴇʀ sᴀᴡ ᴛʜᴀᴛ ᴜsᴇʀ's ᴍᴇssᴀɢᴇ ᴏʀ I ᴀᴍ ᴜɴᴀʙʟᴇ ᴛᴏ ғᴇᴛᴄʜ ᴛʜᴇ ID."
                )
            else:
                await message.reply(
                    f"ᴜsᴇʀ {message.reply_to_message.from_user.first_name}'s ID ɪs <code>{message.reply_to_message.from_user.id}</code>."
                )
        elif message.reply_to_message.sticker:
            if message.reply_to_message.forward_from_chat:
                await message.reply(
                    f"Tʜᴇ ғᴏʀᴡᴀʀᴅᴇᴅ {str(message.reply_to_message.forward_from_chat.type)[9:].lower()}, {message.reply_to_message.forward_from_chat.title} ʜᴀs ᴀɴ ID ᴏғ <code>{message.reply_to_message.forward_from_chat.id}</code> \nᴀɴᴅ ᴛʜᴇ ʀᴇᴘʟɪᴇᴅ sᴛɪᴄᴋᴇʀ ID ɪs <code>{message.reply_to_message.sticker.file_id}</code>"
                )

            elif message.reply_to_message.forward_from:
                await message.reply(
                    f"Tʜᴇ ғᴏʀᴡᴀʀᴅᴇᴅ ᴜsᴇʀ, {message.reply_to_message.forward_from.first_name} ʜᴀs ᴀɴ ID ᴏғ <code>{message.reply_to_message.forward_from.id}</code> \nᴀɴᴅ ᴛʜᴇ ʀᴇᴘʟɪᴇᴅ sᴛɪᴄᴋᴇʀ ID ɪs <code>{message.reply_to_message.sticker.file_id}</code>."
                )

            elif message.reply_to_message.forward_sender_name:
                await message.reply(
                    "Sᴏʀʀʏ, I ɴᴇᴠᴇʀ sᴀᴡ ᴛʜᴀᴛ ᴜsᴇʀ's ᴍᴇssᴀɢᴇ ᴏʀ I ᴀᴍ ᴜɴᴀʙʟᴇ ᴛᴏ ғᴇᴛᴄʜ ᴛʜᴇ ID."
                )

            else:
                await message.reply(
                    f"ᴜsᴇʀ {message.reply_to_message.from_user.first_name}'s ID ɪs <code>{message.reply_to_message.from_user.id}</code>\n ᴀɴᴅ ᴛʜᴇ ʀᴇᴘʟɪᴇᴅ sᴛɪᴄᴋᴇʀ ID ɪs <code>{message.reply_to_message.sticker.file_id}</code>."
                )
        else:
            await message.reply(
                f"User {message.reply_to_message.from_user.first_name}'s ᴜsᴇʀ ID ɪs <code>{message.reply_to_message.from_user.id}</code>."
            )
    except Exception as r:
        await message.reply(f"Aɴ ᴇʀʀᴏʀ ᴏᴄᴄᴜʀʀᴇᴅ ᴡʜɪʟᴇ ɢᴇᴛᴛɪɴɢ ᴛʜᴇ ID. {r}")


__MODULE__ = "Usᴇʀɪᴅ"
__HELP__ = """
**ɪᴅ ʀᴇᴛʀɪᴇᴠᴇʀ:**

• `/id`: Retrieve user and chat IDs.
"""

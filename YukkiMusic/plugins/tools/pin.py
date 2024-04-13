from pyrogram import enums, filters

from YukkiMusic import app
from YukkiMusic.utils.filter import admin_filter


@app.on_message(filters.command("pin") & admin_filter)
async def pin(_, message):
    replied = message.reply_to_message
    message.chat.title
    chat_id = message.chat.id
    user_id = message.from_user.id
    message.from_user.mention

    if message.chat.type == enums.ChatType.PRIVATE:
        await message.reply_text("**ᴛʜɪs ᴄᴏᴍᴍᴀɴᴅ ᴡᴏʀᴋs ᴏɴʟʏ ᴏɴ ɢʀᴏᴜᴘs !**")
    elif not replied:
        await message.reply_text("**ʀᴇᴘʟʏ ᴛᴏ ᴀ ᴍᴇssᴀɢᴇ ᴛᴏ ᴘɪɴ ɪᴛ !**")
    else:
        user_stats = await app.get_chat_member(chat_id, user_id)
        if user_stats.privileges.can_pin_messages and message.reply_to_message:
            try:
                await message.reply_to_message.pin()
                await message.reply_text(f"**ᴍᴇssᴀɢᴇ ᴘɪɴɴᴇᴅ!**")
            except Exception as e:
                await message.reply_text(str(e))


# ------------------------------------------------------------------------------- #


@app.on_message(filters.command("unpin") & admin_filter)
async def unpin(_, message):
    replied = message.reply_to_message
    message.chat.title
    chat_id = message.chat.id
    user_id = message.from_user.id
    message.from_user.mention

    if message.chat.type == enums.ChatType.PRIVATE:
        await message.reply_text("**ᴛʜɪs ᴄᴏᴍᴍᴀɴᴅ ᴡᴏʀᴋs ᴏɴʟʏ ᴏɴ ɢʀᴏᴜᴘs !**")
    elif not replied:
        await message.reply_text("**ʀᴇᴘʟʏ ᴛᴏ ᴀ ᴍᴇssᴀɢᴇ ᴛᴏ ᴜɴᴘɪɴ ɪᴛ !**")
    else:
        user_stats = await app.get_chat_member(chat_id, user_id)
        if user_stats.privileges.can_pin_messages and message.reply_to_message:
            try:
                await message.reply_to_message.unpin()
                await message.reply_text(f"**ᴍᴇssᴀɢᴇ ᴜɴᴘɪɴɴᴇᴅ !**")
            except Exception as e:
                await message.reply_text(str(e))


# ------------------------------------------------------------------------------- #
@app.on_message(filters.command("unpinall") & admin_filter)
async def unpin(_, message):
    chat_id = message.chat.id
    user_id = message.from_user.id

    if message.chat.type == enums.ChatType.PRIVATE:
        await message.reply_text("**ᴛʜɪs ᴄᴏᴍᴍᴀɴᴅ ᴡᴏʀᴋs ᴏɴʟʏ ᴏɴ ɢʀᴏᴜᴘs !**")
    else:
        user_stats = await app.get_chat_member(chat_id, user_id)
        if user_stats.privileges.can_pin_messages:
            try:
                await app.unpin_all_chat_messages(chat_id)
                await message.reply_text(f"**ᴀʟʟ ᴍᴇssᴀɢᴇ ᴜɴᴘɪɴɴᴇᴅ !**")
            except Exception as e:
                await message.reply_text(str(e))

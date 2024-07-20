from YukkiMusic import app
from pyrogram import filters
import nekos


@app.on_message(filters.command("slap"))
async def slap(client, message):
    try:
        if message.reply_to_message:
            await message.reply_video(
                nekos.img("slap"),
                caption=f"{message.from_user.mention} sʟᴀᴘᴘᴇᴅ {message.reply_to_message.from_user.mention}",
            )
        else:
            await message.reply_video(nekos.img("slap"))
    except Exception as e:
        await message.reply_text(f"Error: {e}")


__HELP__ = """
Aᴠᴀɪʟᴀʙʟᴇ ᴄᴏᴍᴍᴀɴᴅs:
- /sʟᴀᴘ: Sʟᴀᴘs sᴏᴍᴇᴏɴᴇ. Iғ ᴜsᴇᴅ ᴀs ᴀ ʀᴇᴘʟʏ, sʟᴀᴘs ᴛʜᴇ ʀᴇᴘʟɪᴇᴅ ᴜsᴇʀ.
"""
__MODULE__ = "Sʟᴀᴘ"

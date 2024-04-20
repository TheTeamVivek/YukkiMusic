import requests
from pyrogram import Client
from pyrogram import filters
from MukeshAPI import api


@Client.on_message(filters.command("hastag"))
async def hastag(client, message):

    try:
        text = message.text.split(" ", 1)[1]
        res = api.hashtag(text)
        results = " ".join(res)
        hashtags = results.replace(",", "").replace("[", "").replace("]", "")

    except IndexError:
        return await message.reply_text("Example:\n\n/hastag python")

    await message.reply_text(
        f"ʜᴇʀᴇ ɪs ʏᴏᴜʀ  ʜᴀsᴛᴀɢ :\n<pre>{hashtags}</pre>", quote=True
    )


help = """
Yᴏᴜ ᴄᴀɴ ᴜsᴇ ᴛʜɪs ʜᴀsʜᴛᴀɢ ɢᴇɴᴇʀᴀᴛᴏʀ ᴡʜɪᴄʜ ᴡɪʟʟ ɢɪᴠᴇ ʏᴏᴜ ᴛʜᴇ ᴛᴏᴘ 𝟹𝟶 ᴀɴᴅ ᴍᴏʀᴇ ʜᴀsʜᴛᴀɢs ʙᴀsᴇᴅ ᴏғғ ᴏғ ᴏɴᴇ ᴋᴇʏᴡᴏʀᴅ sᴇʟᴇᴄᴛɪᴏɴ.
° /hastag enter word to generate hastag.
°Exᴀᴍᴘʟᴇ:  /hastag python """

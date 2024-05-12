import os
from telegraph import upload_file
from pyrogram import Client, filters


@Client.on_message(filters.command(["tgm", "telegraph"]))
def ul(_, message):
    reply = message.reply_to_message
    if not reply:
        message.reply("Reply to a media")
    if reply.media:
        i = message.reply("𝐌𝙰𝙺ing 𝐀 𝐋𝙸𝙽𝙺...")
        try:
            path = reply.download()
            fk = upload_file(path)
            for x in fk:
                url = "https://telegra.ph" + x
            i.edit(f" 🇾ᴏᴜʀ🇹ᴇʟᴇɢʀᴀᴘʜ {url}")
            os.remove(path)
        except Exception as e:
            i.edit(f"❌Error \n{e}")

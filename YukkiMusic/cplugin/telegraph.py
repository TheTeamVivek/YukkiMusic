import os
from telegraph import upload_file
from pyrogram import Client, filters
from YukkiMusic import app


@Client.on_message(filters.command(["tgm", "tgt"]))
def ul(_, message):
    reply = message.reply_to_message
    if reply.media:
        i = message.reply("𝐌𝙰𝙺𝙴 𝐀 𝐋𝙸𝙽𝙺...")
        path = reply.download()
        fk = upload_file(path)
        for x in fk:
            url = "https://telegra.ph" + x

        i.edit(f" 🇾ᴏᴜʀ🇹ᴇʟᴇɢʀᴀᴘʜ {url}")
        os.remove(path)
    else:
        message.reply("Reply to a media")

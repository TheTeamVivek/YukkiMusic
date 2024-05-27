from io import BytesIO
from os import path, remove
from time import time

import img2pdf
from PIL import Image
from pyrogram import filters
from pyrogram.types import Message

from YukkiMusic import app
from YukkiMusic.utils.error import capture_err
from YukkiMusic.utils.sections import section


async def convert(
    main_message: Message,
    reply_messages,
    status_message: Message,
    start_time: float,
):
    m = status_message

    documents = []

    for message in reply_messages:
        if not message.document:
            return await m.edit("ɴᴏᴛ ᴀ ᴅᴏᴄᴜᴍᴇɴᴛ!")

        if message.document.mime_type.split("/")[0] != "image":
            return await m.edit("ɪɴᴠᴀʟɪᴅ ᴍɪᴍᴇ ᴛʏᴘᴇ!")

        if message.document.file_size > 5000000:
            return await m.edit("sɪᴢᴇ ᴛᴏᴏ ʟᴀʀɢᴇ, ᴀʙᴏʀᴛᴇᴅ!")
        documents.append(await message.download())

    for img_path in documents:
        img = Image.open(img_path).convert("RGB")
        img.save(img_path, "JPEG", quality=100)

    pdf = BytesIO(img2pdf.convert(documents))
    pdf.name = "Pdf by YukkiMusic.pdf"

    if len(main_message.command) >= 2:
        names = main_message.text.split(None, 1)[1]
        if not names.endswith(".pdf"):
            pdf.name = names + ".pdf"
        else:
            pdf.name = names

    elapsed = round(time() - start_time, 2)

    await main_message.reply_document(
        document=pdf,
        caption=section(
            "IMG2PDF",
            body={
                "Title": pdf.name,
                "Size": f"{pdf.__sizeof__() / (10 ** 6)}MB",
                "Pages": len(documents),
                "Took": f"{elapsed}s",
            },
        ),
    )

    await m.delete()
    pdf.close()
    for file in documents:
        if path.exists(file):
            remove(file)


@app.on_message(filters.command("pdf"))
@capture_err
async def img_to_pdf(_, message: Message):
    reply = message.reply_to_message
    if not reply:
        return await message.reply(
            "ʀᴇᴘʟʏ ᴛᴏ ᴀᴍ ɪᴍᴀɢᴇ (ᴀs ᴅᴏᴄᴜᴍᴇɴᴛ) ᴏʀ sᴇɴᴅ ᴍᴇ ᴍᴜʟᴛɪᴘʟᴇ ᴘʜᴏᴛᴏs ᴀᴛ ᴏɴᴇ ᴛɪᴍᴇ."
        )

    m = await message.reply_text("ᴄᴏɴᴠᴇʀᴛɪɴɢ.....")
    start_time = time()

    if reply.media_group_id:
        messages = await app.get_media_group(
            message.chat.id,
            reply.id,
        )
        return await convert(message, messages, m, start_time)

    return await convert(message, [reply], m, start_time)

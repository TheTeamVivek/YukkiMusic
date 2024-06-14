from pyrogram.types import Message
from re import findall
import datetime
from YukkiMusic.utils.database import is_pnote_on, get_note
from YukkiMusic.utils.keyboard import ikb
from YukkiMusic.utils.functions import extract_text_and_keyb
from YukkiMusic import app


async def send_notes(message: Message, chat_id, text):
    if not text:
        return
    _note = await get_note(chat_id, text)
    if not _note:
        return

    type = _note["type"]
    data = _note["data"]
    file_id = _note.get("file_id")
    keyb = None

    if data:
        if "{app.mention}" in data:
            data = data.replace("{app.mention}", app.mention)
        if "{GROUPNAME}" in data:
            data = data.replace("{GROUPNAME}", (await app.get_chat(chat_id)).title)
        if "{NAME}" in data:
            data = data.replace("{NAME}", message.from_user.mention)
        if "{ID}" in data:
            data = data.replace("{ID}", f"`message.from_user.id`")
        if "{FIRSTNAME}" in data:
            data = data.replace("{FIRSTNAME}", message.from_user.first_name)
        if "{SURNAME}" in data:
            sname = (
                message.from_user.last_name if message.from_user.last_name else "None"
            )
            data = data.replace("{SURNAME}", sname)
        if "{USERNAME}" in data:
            susername = (
                message.from_user.username if message.from_user.username else "None"
            )
            data = data.replace("{USERNAME}", susername)
        if "{DATE}" in data:
            DATE = datetime.datetime.now().strftime("%Y-%m-%d")
            data = data.replace("{DATE}", DATE)
        if "{WEEKDAY}" in data:
            WEEKDAY = datetime.datetime.now().strftime("%A")
            data = data.replace("{WEEKDAY}", WEEKDAY)
        if "{TIME}" in data:
            TIME = datetime.datetime.now().strftime("%H:%M:%S")
            data = data.replace("{TIME}", f"{TIME} UTC")

        if findall(r"\[.+\,.+\]", data):
            keyboard = extract_text_and_keyb(ikb, data)
            if keyboard:
                data, keyb = keyboard

    replied_message = message.reply_to_message
    if replied_message:
        replied_user = (
            replied_message.from_user
            if replied_message.from_user
            else replied_message.sender_chat
        )
        if replied_user.id != from_user.id:
            message = replied_message

    await get_reply(message, type, file_id, data, keyb)


async def get_reply(message, type, file_id, data, keyb):
    if type == "text":
        await message.reply_text(
            text=data,
            reply_markup=keyb,
            disable_web_page_preview=True,
        )
    if type == "sticker":
        await message.reply_sticker(
            sticker=file_id,
        )
    if type == "animation":
        await message.reply_animation(
            animation=file_id,
            caption=data,
            reply_markup=keyb,
        )
    if type == "photo":
        await message.reply_photo(
            photo=file_id,
            caption=data,
            reply_markup=keyb,
        )
    if type == "document":
        await message.reply_document(
            document=file_id,
            caption=data,
            reply_markup=keyb,
        )
    if type == "video":
        await message.reply_video(
            video=file_id,
            caption=data,
            reply_markup=keyb,
        )
    if type == "video_note":
        await message.reply_video_note(
            video_note=file_id,
        )
    if type == "audio":
        await message.reply_audio(
            audio=file_id,
            caption=data,
            reply_markup=keyb,
        )
    if type == "voice":
        await message.reply_voice(
            voice=file_id,
            caption=data,
            reply_markup=keyb,
        )

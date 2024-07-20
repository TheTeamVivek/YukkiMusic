import os

from pyrogram import enums, filters
from pyrogram.types import Message

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import is_gbanned_user

n = "\n"
w = " "


def bold(x):
    return f"**{x}:** "


def bold_ul(x):
    return f"**--{x}:**-- "


def mono(x):
    return f"`{x}`{n}"


def section(
    title: str,
    body: dict,
    indent: int = 2,
    underline: bool = False,
) -> str:
    text = (bold_ul(title) + n) if underline else bold(title) + n

    for key, value in body.items():
        if value is not None:
            text += (
                indent * w
                + bold(key)
                + (
                    (value[0] + n)
                    if isinstance(value, list) and isinstance(value[0], str)
                    else mono(value)
                )
            )
    return text


async def userstatus(user_id):
    try:
        user = await app.get_users(user_id)
        x = user.status
        if x == enums.UserStatus.RECENTLY:
            return "Recently."
        elif x == enums.UserStatus.LAST_WEEK:
            return "Last week."
        elif x == enums.UserStatus.LONG_AGO:
            return "Long time ago."
        elif x == enums.UserStatus.OFFLINE:
            return "Offline."
        elif x == enums.UserStatus.ONLINE:
            return "Online."
    except BaseException:
        return "**sᴏᴍᴇᴛʜɪɴɢ ᴡʀᴏɴɢ ʜᴀᴘᴘᴇɴᴇᴅ !**"


async def get_user_info(user, already=False):
    if not already:
        user = await app.get_users(user)
    if not user.first_name:
        return ["Deleted account", None]
    user_id = user.id
    online = await userstatus(user_id)
    username = user.username
    first_name = user.first_name
    mention = user.mention("Link")
    dc_id = user.dc_id
    photo_id = user.photo.big_file_id if user.photo else None
    is_gbanned = await is_gbanned_user(user_id)
    is_sudo = user_id in SUDOERS
    is_premium = user.is_premium
    body = {
        "ɴᴀᴍᴇ": [first_name],
        "ᴜsᴇʀɴᴀᴍᴇ": [("@" + username) if username else "Null"],
        "ɪᴅ": user_id,
        "ᴅᴄ ɪᴅ": dc_id,
        "ᴍᴇɴᴛɪᴏɴ": [mention],
        "ᴘʀᴇɪᴍɪᴜᴍ": is_premium,
        "ʟᴀsᴛ sᴇᴇɴ": online,
    }
    caption = section("ᴜsᴇʀ ɪɴғᴏ", body)
    return [caption, photo_id]


async def get_chat_info(chat):
    chat = await app.get_chat(chat)
    username = chat.username
    link = f"[Link](t.me/{username})" if username else "Null"
    photo_id = chat.photo.big_file_id if chat.photo else None
    info = f"""
❅─────✧❅✦❅✧─────❅
             ✦ ᴄʜᴀᴛ ɪɴғᴏ ✦

➻ ᴄʜᴀᴛ ɪᴅ ‣ {chat.id}
➻ ɴᴀᴍᴇ ‣ {chat.title}
➻ ᴜsᴇʀɴᴀᴍᴇ ‣ {chat.username}
➻ ᴅᴄ ɪᴅ ‣ {chat.dc_id}
➻ ᴅᴇsᴄʀɪᴘᴛɪᴏɴ  ‣ {chat.description}
➻ ᴄʜᴀᴛᴛʏᴘᴇ ‣ {chat.type}
➻ ɪs ᴠᴇʀɪғɪᴇᴅ ‣ {chat.is_verified}
➻ ɪs ʀᴇsᴛʀɪᴄᴛᴇᴅ ‣ {chat.is_restricted}
➻ ɪs ᴄʀᴇᴀᴛᴏʀ ‣ {chat.is_creator}
➻ ɪs sᴄᴀᴍ ‣ {chat.is_scam}
➻ ɪs ғᴀᴋᴇ ‣ {chat.is_fake}
➻ ᴍᴇᴍʙᴇʀ's ᴄᴏᴜɴᴛ ‣ {chat.members_count}
➻ ʟɪɴᴋ ‣ {link}


❅─────✧❅✦❅✧─────❅"""

    return info, photo_id


@app.on_message(filters.command("info"))
async def info_func(_, message: Message):
    if message.reply_to_message:
        user = message.reply_to_message.from_user.id
    elif not message.reply_to_message and len(message.command) == 1:
        user = message.from_user.id
    elif not message.reply_to_message and len(message.command) != 1:
        user_input = message.text.split(None, 1)[1]
        if user_input.isdigit():
            user = int(user_input)
        elif user_input.startswith("@"):
            user = user_input
        else:
            return await message.reply_text(
                "ᴘʟᴇᴀsᴇ ᴘʀᴏᴠɪᴅᴇ ᴀ ᴜsᴇʀ's ᴜsᴇʀ ɪᴅ ᴏʀ ᴜsᴇʀɴᴀᴍᴇ ᴏʀ ʀᴇᴘʟʏ ᴛᴏ ᴀ ᴜsᴇʀ ᴛᴏ ɢᴇᴛ ɪɴғᴏ"
            )

    m = await message.reply_text("ᴘʀᴏᴄᴇssɪɴɢ...")

    try:
        info_caption, photo_id = await get_user_info(user)
    except Exception as e:
        return await m.edit(str(e))

    if not photo_id:
        return await m.edit(info_caption, disable_web_page_preview=True)
    photo = await app.download_media(photo_id)

    await message.reply_photo(photo, caption=info_caption, quote=False)
    await m.delete()
    os.remove(photo)


@app.on_message(filters.command("chatinfo"))
async def chat_info_func(_, message: Message):
    splited = message.text.split()
    if len(splited) == 1:
        chat = message.chat.id
        if chat == message.from_user.id:
            return await message.reply_text("**Usage:**/chat_info [USERNAME|ID]")
    else:
        chat = splited[1]
    try:
        m = await message.reply_text("Processing")

        info_caption, photo_id = await get_chat_info(chat)
        if not photo_id:
            return await m.edit(info_caption, disable_web_page_preview=True)

        photo = await app.download_media(photo_id)
        await message.reply_photo(photo, caption=info_caption, quote=False)

        await m.delete()
        os.remove(photo)
    except Exception as e:
        await m.edit(e)


__MODULE__ = "Iɴғᴏ"
__HELP__ = """
**ᴜsᴇʀ & ᴄʜᴀᴛ ɪɴғᴏʀᴍᴀᴛɪᴏɴ:**

• `/info`: Gᴇᴛ ɪɴғᴏʀᴍᴀᴛɪᴏɴ ᴀʙᴏᴜᴛ ᴛʜᴇ ᴜsᴇʀ. Usᴇʀɴᴀᴍᴇ, ID, ᴀɴᴅ ᴍᴏʀᴇ.
• `/chatinfo [ᴜsᴇʀɴᴀᴍᴇ|ɪᴅ]`: Gᴇᴛ ɪɴғᴏʀᴍᴀᴛɪᴏɴ ᴀʙᴏᴜᴛ ᴛʜᴇ ᴄʜᴀᴛ. ᴍᴇᴍʙᴇʀ ᴄᴏᴜɴᴛ, ɪs ᴠᴇʀɪғɪᴇᴅ, ɪɴᴠɪᴛᴇ ʟɪɴᴋ, ᴀɴᴅ ᴍᴏʀᴇ.
"""

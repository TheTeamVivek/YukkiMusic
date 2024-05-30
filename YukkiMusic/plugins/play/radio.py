from pyrogram import filters
from pyrogram.types import Message

from config import BANNED_USERS, adminlist
from strings import get_string
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import get_cmode, get_lang, get_playmode, get_playtype
from YukkiMusic.utils.logger import play_logs
from YukkiMusic.utils.stream.stream import stream


@app.on_message(
    filters.command(["radio", "cradio", "vradio"]) & filters.group & ~BANNED_USERS
)
async def radio(
    client,
    message: Message,
):
    language = await get_lang(message.chat.id)
    _ = get_string(language)
    playmode = await get_playmode(message.chat.id)
    playty = await get_playtype(message.chat.id)
    if playty != "Everyone":
        if message.from_user.id not in SUDOERS:
            admins = adminlist.get(message.chat.id)
            if not admins:
                return await message.reply_text(_["admin_18"])
            else:
                if message.from_user.id not in admins:
                    return await message.reply_text(_["play_4"])
    if message.command[0][0] == "c":
        chat_id = await get_cmode(message.chat.id)
        if chat_id is None:
            return await message.reply_text(_["setting_12"])
        try:
            chat = await app.get_chat(chat_id)
        except:
            return await message.reply_text(_["cplay_4"])
        channel = chat.title
    else:
        chat_id = message.chat.id
        channel = None

    video = None
    if message.command[0][0] == "v":
        video = True
    else:
        if "-v" in message.text:
            video = True
        else:
            video = None

    mystic = await message.reply_text(
        _["play_2"].format(channel) if channel else _["play_1"]
    )
    try:
        await stream(
            _,
            mystic,
            message.from_user.id,
            # RADIO_URL,
            "https://www.youtube.com/live/eu191hR_LEc?si=T-9QYD548jd0Mogp",
            chat_id,
            message.from_user.mention,
            message.chat.id,
            video=video,
            streamtype="index",
        )
    except Exception as e:
        ex_type = type(e).__name__
        err = e if ex_type == "AssistantErr" else _["general_3"].format(ex_type)
        return await mystic.edit_text(err)
    return await play_logs(message, streamtype="M3u8 or Index Link")

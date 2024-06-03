from bing_image_urls import bing_image_urls
from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    InputMediaPhoto,
)
from requests import get

from config import BANNED_USERS
from YukkiMusic import app
from YukkiMusic.utils.image import gen_image


@app.on_message(
    filters.command(["pinterest", "image"], prefixes=["/", "!", "."]) & ~BANNED_USERS
)
async def pinterest(_, message):
    command = message.text.split()[0][1:]
    chat_id = message.chat.id

    try:
        query = message.text.split(None, 1)[1]
    except:
        return await message.reply("**ɢɪᴠᴇ ɪᴍᴀɢᴇ ɴᴀᴍᴇ ғᴏʀ sᴇᴀʀᴄʜ 🔍**")
    if command == "pinterest":
        images = get(f"https://pinterest-api-one.vercel.app/?q={query}").json()
        media_group = []
        msg = await message.reply(f"sᴇᴀʀᴄʜɪɴɢ ɪᴍᴀɢᴇs ғʀᴏᴍ ᴘɪɴᴛᴇʀᴇᴛs...")
        for url in images["images"][:7]:

            media_group.append(InputMediaPhoto(media=url))
        try:
            await msg.edit("Uᴘʟᴏᴀᴅɪɴɢ....")
            await app.send_media_group(
                chat_id=chat_id, media=media_group, reply_to_message_id=message.id
            )
            return await msg.delete()

        except Exception as e:
            return await msg.edit(f"ᴇʀʀᴏʀ : {e}")

    elif command == "image":
        images = bing_image_urls(query, limit=7)
        BING = []

        msg = await message.reply(f"sᴇᴀʀᴄʜɪɴɢ ɪᴍᴀɢᴇs ғʀᴏᴍ ʙɪɴɢ...")
        for url in images:

            BING.append(InputMediaPhoto(media=url))

        try:
            await msg.edit("Uᴘʟᴏᴀᴅɪɴɢ....")
            await app.send_media_group(
                chat_id=chat_id, media=BING, reply_to_message_id=message.id
            )
            return await msg.delete()

        except Exception as e:
            return await msg.edit(f"ᴇʀʀᴏʀ : {e}")


re_keyboard = InlineKeyboardMarkup(
    [
        [InlineKeyboardButton(text="Rᴇғʀᴇsʜ", callback_data="randomimagerefresh")],
        [InlineKeyboardButton(text="〆 ᴄʟᴏsᴇ 〆", callback_data="close")],
    ]
)


@app.on_message(filters.command(["rimage", "randomimage"]) & ~BANNED_USERS)
async def wall(client, message):
    img = gen_image()
    await message.reply_photo(img, reply_markup=re_keyboard)


@app.on_callback_query(filters.regex("randomimagerefresh") & ~BANNED_USERS)
async def refresh_cat(c, m: CallbackQuery):
    img = gen_image()
    await m.edit_message_media(
        InputMediaPhoto(media=img),
        reply_markup=re_keyboard,
    )


__MODULE__ = "Iᴍᴀɢᴇ"
__HELP__ = """/pinterest [ǫᴜᴇʀʏ] - ᴛᴏ ɢᴇᴛ ᴛᴏᴘ 7 ɪᴍᴀɢᴇs ғʀᴏᴍ ᴘɪɴᴛᴇʀᴇsᴛ
/image [ǫᴜᴇʀʏ] - ᴛᴏ ɢᴇᴛ ᴛᴏᴘ ɪᴍᴀɢᴇs ғʀᴏᴍ ʙɪɴɢ
/wall | /wallpaper - [ǫᴜᴇʀʏ] - ᴛᴏ ɢᴇᴛ ʀᴇǫᴜᴇsᴛᴇᴅ ᴡᴀʟᴘᴀᴘᴇʀ
/rimage | /randomimage - ᴛᴏ ɢᴇᴛ ʀᴀɴᴅᴏᴍ ɪᴍᴀɢᴇ
/cat - ɢᴇᴛ ʀᴀɴᴅᴏᴍ ᴄᴀᴛ ɪᴍᴀɢᴇs
/dog - ɢᴇᴛ ʀᴀɴᴅᴏᴍ ᴅᴏɢ ɪᴍᴀɢᴇs
"""

from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup
from YukkiMusic import app
from config import START_IMG_URL, SUPPORT_CHANNEL, SUPPORT_GROUP
from pyrogram.errors import ChatAdminRequired, UserNotParticipant
from pyrogram.enums import ChatMemberStatus

@app.on_callback_query(filters.regex("source_code"))
async def gib_repo_callback(_, callback_query):
    if app.username == "TprinceMusicBot":
        try:
            get = await app.get_chat_member(-1002159045835, callback_query.from_user.id)
        except ChatAdminRequired:
            return await callback_query.message.edit("Asᴋ ɪɴ Sᴜᴘᴘᴏʀᴛ Cʜᴀᴛ ғᴏʀ ᴛʜɪs", reply_markup=InlineKeyboardMarkup([[InlineKeyboardButton(text="ʙᴀᴄᴋ", callback_data="settingsback_helper"), InlineKeyboardButton(text="ᴄʟᴏsᴇ", callback_data="close")]]))
        except UserNotParticipant:
            return await callback_query.message.edit("Pʟᴇᴀsᴇ ɪᴏɪɴ ᴏᴜʀ ᴄʜᴀɴɴᴇʟ ᴛʜᴇɴ ɪ ᴛʜɪɴᴋ", reply_markup=InlineKeyboardMarkup([[InlineKeyboardButton(text="ᴄʜᴀɴɴᴇʟ", url=SUPPORT_CHANNEL), InlineKeyboardButton(text="ᴄʟᴏsᴇ", callback_data="close")]]))
        if get.status == ChatMemberStatus.LEFT:
            return await callback_query.message.edit("Pʟᴇᴀsᴇ ɪᴏɪɴ ᴏᴜʀ ᴄʜᴀɴɴᴇʟ ᴛʜᴇɴ ɪ ᴛʜɪɴᴋ", reply_markup=InlineKeyboardMarkup([[InlineKeyboardButton(text="ᴄʜᴀɴɴᴇʟ", url=SUPPORT_CHANNEL), InlineKeyboardButton(text="ᴄʟᴏsᴇ", callback_data="close")]]))
        else:
            return await callback_query.message.edit("ᴡᴏ ᴅᴀʀᴀsᴀʟ ᴍᴀɪɴ ɪss [ʀᴇᴘᴏ](https://github.com/TeamYukki/YukkiMusicBot) sᴇ ʙᴀɴᴀ ʜᴜɴ ᴛᴏ ᴀᴀᴘ ɪsssᴇ ᴜsᴇ ᴋᴀʀ sᴀᴋᴛᴇ ʜᴏ ᴘʜɪʟᴀʟ ᴀɢᴀʀ ᴀᴀᴘᴋᴏ ʀᴇᴀʟ ʀᴇᴘᴏ ᴄʜᴀɪʏᴇ ᴛᴏ sᴜᴘᴘᴏʀᴛ ᴄʜᴀᴛ ᴍᴇɪɴ ᴀᴀᴘ ᴘᴜᴄʜ sᴀᴋᴛᴇ ʜᴏ  😅😅😅😅",disable_web_page_preview=True, reply_markup=InlineKeyboardMarkup([[InlineKeyboardButton(text="ɢʀᴏᴜᴘ", url=SUPPORT_GROUP), InlineKeyboardButton(text="ʙᴀᴄᴋ", callback_data="settingsback_helper")]]))

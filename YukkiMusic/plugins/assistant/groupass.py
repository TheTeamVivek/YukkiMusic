from config import PK
from pyrogram import filters
from pyrogram.types import Message

from YukkiMusic import app
from YukkiMusic.utils.database import get_assistant, save_assistant, set_assistant
from YukkiMusic.misc import SUDOERS
from YukkiMusic.core.userbot import assistants 
from YukkiMusic.utils.filter import admin_filter
from YukkiMusic.utils.assistant import get_assistant_details, assistant

@app.on_message(filters.command("changeassistant") & admin_filter)
async def assis_change(_, message: Message):
	avt = await assistant()
	if avt == True:
		return await message.reply_text("sᴏʀʀʏ sɪʀ! ɪɴ ʙᴏᴛ sᴇʀᴠᴇʀ ᴏɴʟʏ ᴏɴʀ ᴀssɪsᴛᴀɴᴛ ᴀᴠᴀɪʟᴀʙʟᴇ ᴛʜᴇʀᴇғᴏʀᴇ ʏᴏᴜ ᴄᴀɴᴛ ᴄʜᴀɴɢᴇ ᴀssɪsᴛᴀɴᴛ")
    usage = f"**ᴅᴇᴛᴇᴄᴛᴇᴅ ᴡʀᴏɴɢ ᴄᴏᴍᴍᴀɴᴅ ᴜsᴀsɢᴇ \n**ᴜsᴀsɢᴇ:**\n/changeassistant - ᴛᴏ ᴄʜᴀɴɢᴇ ʏᴏᴜʀ ᴄᴜʀʀᴇɴᴛ ɢʀᴏᴜᴘ's ᴀssɪsᴛᴀɴᴛ ᴛᴏ ʀᴀɴᴅᴏᴍ ᴀssɪsᴛᴀɴᴛ ɪɴ ʙᴏᴛ sᴇʀᴠᴇʀ"
    if len(message.command) > 2:
        return await message.reply_text(usage)
    a = await get_assistant(message.chat.id)
    DETAILS = f"ʏᴏᴜʀ ᴄʜᴀᴛ's ᴀssɪsᴛᴀɴᴛ ʜᴀs ʙᴇᴇɴ ᴄʜᴀɴɢᴇᴅ ғʀᴏᴍ [{a.name}](https://t.me/{a.username}) "
    try:
    	await a.leave_chat(message.chat.id)
    except:
    	pass
    b = await set_assistant(message.chat.id)
    DETAILS += f"ᴛᴏ [{b.name}](https://t.me/{b.username})"
    try:
    	await b.join_chat(message.chat.id)
    except:
    	pass
    await message.reply_text(DETAILS, disable_web_page_preview = True, protect_content=PK)

@app.on_message(filters.command("setassistant") & admin_filter)
async def assis_set(_, message: Message):
	avt = await assistant()
	if avt == True:
		return await message.reply_text("sᴏʀʀʏ sɪʀ! ɪɴ ʙᴏᴛ sᴇʀᴠᴇʀ ᴏɴʟʏ ᴏɴᴇ ᴀssɪsᴛᴀɴᴛ ᴀᴠᴀɪʟᴀʙʟᴇ ᴛʜᴇʀᴇғᴏʀᴇ ʏᴏᴜ ᴄᴀɴ'ᴛ ᴄʜᴀɴɢᴇ ᴀssɪsᴛᴀɴᴛ")
    usage = await get_assistant_details()
    if len(message.command) != 2:
        return await message.reply_text(usage, disable_web_page_preview = True, protect_content=PK)
    query = message.text.split(None, 1)[1].strip()
    if query not in assistants:
        return await message.reply_text(usage, disable_web_page_preview = True, protect_content=PK)
    a = await get_assistant(message.chat.id)
    try:
    	await a.leave_chat(message.chat.id)
    except:
    	pass
    await save_assistant(message.chat.id, query)
    b = await get_assistant(message.chat.id)
    try:
    	await b.join_chat(message.chat.id)
    except:
    	pass
    
    DETAILS = f""" ʏᴏᴜʀ ᴄʜᴀᴛ's  ɴᴇᴡ ᴀssɪsᴛᴀɴᴛ ᴅᴇᴛᴀɪʟs:
		ᴀssɪsᴛᴀɴᴛ ɴᴀᴍᴇ :- {a.name}
		ᴀssɪsᴛᴀɴᴛ ᴜsᴇʀɴᴀᴍᴇ :- {a.username}
		ᴀssɪsᴛᴀɴᴛ ɪᴅ:- @{a.id}"""
    return await message.reply_text(DETAILS, disable_web_page_preview = True, protect_content=PK)


@app.on_message(filters.command("checkassistant") & filters.group & admin_filter)
async def check_ass(_, message: Message):
    assistant = await get_assistant(message.chat.id)
    DETAILS = f""" ʏᴏᴜʀ ᴄʜᴀᴛ's ᴀssɪsᴛᴀɴᴛ ᴅᴇᴛᴀɪʟs:
		ᴀssɪsᴛᴀɴᴛ ɴᴀᴍᴇ :- {assistant.name}
		ᴀssɪsᴛᴀɴᴛ ᴜsᴇʀɴᴀᴍᴇ :- {assistant.username}
		ᴀssɪsᴛᴀɴᴛ ɪᴅ:- @{assistant.id}
	"""
    await message.reply_text(
        DETAILS, disable_web_page_preview = True, protect_content=PK
    )
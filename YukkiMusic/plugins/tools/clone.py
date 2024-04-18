import re
import logging
from pymongo import MongoClient
from pyrogram import Client, filters
from pyrogram.types import Message
from pyrogram.errors.exceptions.bad_request_400 import AccessTokenExpired, AccessTokenInvalid
from config import API_ID, API_HASH
from config import MONGO_DB_URI
from YukkiMusic import app

mongo_client = MongoClient(MONGO_DB_URI)
mongo_db = mongo_client["Yukkicloned"]
mongo_collection = mongo_db["Yukkiclone"]


@app.on_message(filters.command("clone") & filters.private)
async def clone_txt(client, message):
    await message.reply_text(f"<b>ʜᴇʟʟᴏ {message.from_user.mention} 👋 </b>\n\n1) sᴇɴᴅ <code>/newbot</code> ᴛᴏ @BotFather\n2) ɢɪᴠᴇ ᴀ ɴᴀᴍᴇ ꜰᴏʀ ʏᴏᴜʀ ʙᴏᴛ.\n3) ɢɪᴠᴇ ᴀ ᴜɴɪǫᴜᴇ ᴜsᴇʀɴᴀᴍᴇ.\n4) ᴛʜᴇɴ ʏᴏᴜ ᴡɪʟʟ ɢᴇᴛ ᴀ ᴍᴇssᴀɢᴇ ᴡɪᴛʜ ʏᴏᴜʀ ʙᴏᴛ ᴛᴏᴋᴇɴ.\n5) ꜰᴏʀᴡᴀʀᴅ ᴛʜᴀᴛ ᴍᴇssᴀɢᴇ ᴛᴏ ᴍᴇ.\n\nᴛʜᴇɴ ɪ ᴀᴍ ᴛʀʏ ᴛᴏ ᴄʀᴇᴀᴛᴇ ᴀ ᴄᴏᴘʏ ʙᴏᴛ ᴏғ ᴍᴇ ғᴏʀ ʏᴏᴜ ᴏɴʟʏ 😌")

@app.on_message((filters.regex(r'\d[0-9]{8,10}:[0-9A-Za-z_-]{35}')) & filters.private)
async def on_clone(client, message):  
    try:
        user_id = message.from_user.id
        bot_token = re.findall(r'\d[0-9]{8,10}:[0-9A-Za-z_-]{35}', message.text, re.IGNORECASE)
        bot_token = bot_token[0] if bot_token else None
        bot_id = re.findall(r'\d[0-9]{8,10}', message.text)
        bots = list(mongo_db.bots.find())
        bot_tokens = None

        for bot in bots:
            bot_tokens = bot['token']

        forward_from_id = message.forward_from.id if message.forward_from else None
        if bot_tokens == bot_token and forward_from_id == 93372553:
            await message.reply_text("**©️ ᴛʜɪs ʙᴏᴛ ɪs ᴀʟʀᴇᴀᴅʏ ᴄʟᴏɴᴇᴅ ʙᴀʙʏ 🐥**")
            return

        if not forward_from_id != 93372553:
            msg = await message.reply_text("**ᴡᴀɪᴛ ᴀ ᴍɪɴᴜᴛᴇ ɪ ᴀᴍ ʙᴏᴏᴛɪɴɢ ʏᴏᴜʀ ʙᴏᴛ..... ❣️**")
            try:
                ai = Client(
                    f"{bot_token}", API_ID, API_HASH,
                    bot_token=bot_token,
                    plugins = dict(root="YukkiMusic.plugins.tools"),
                )
                
                await ai.start()
                bot = await ai.get_me()
                details = {
                    'bot_id': bot.id,
                    'is_bot': True,
                    'user_id': user_id,
                    'name': bot.first_name,
                    'token': bot_token,
                    'username': bot.username
                }
                mongo_db.bots.insert_one(details)
                await msg.edit_text(f"<b>sᴜᴄᴄᴇssғᴜʟʟʏ ᴄʟᴏɴᴇᴅ ʏᴏᴜʀ ʙᴏᴛ: @{bot.username}.\n\nʏᴏᴜ ᴄᴀɴ ᴀʟsᴏ sᴇᴛ ʏᴏᴜʀ sʜᴏʀᴛɴᴇʀ ɪɴ ʏᴏᴜʀ ᴄʟᴏɴᴇᴅ ʙᴏᴛ ғᴏʀ ᴍᴏʀᴇ ɪɴғᴏ sᴛᴀʀᴛ ʏᴏᴜʀ ᴄʟᴏɴᴇᴅ ʙᴏᴛ</b>")
            except BaseException as e:
                logging.exception("Error while cloning bot.")
                await msg.edit_text(f"⚠️ <b>Bot Error:</b>\n\n<code>{e}</code>\n\n**Kindly forward this message to @hhhhhvivek to get assistance.**")
    except Exception as e:
        logging.exception("Error while handling message.")

@app.on_message(filters.command(["delcloned","deletecloned"]) & filters.private)
async def delete_cloned_bot(client, message):
    try:
        bot_token = re.findall(r'\d[0-9]{8,10}:[0-9A-Za-z_-]{35}', message.text, re.IGNORECASE)
        bot_token = bot_token[0] if bot_token else None
        bot_id = re.findall(r'\d[0-9]{8,10}', message.text)

        mongo_collection = mongo_db.bots
        
        cloned_bot = mongo_collection.find_one({"token": bot_token})
        if cloned_bot:
            mongo_collection.delete_one({"token": bot_token})
            await message.reply_text("**🤖 ᴛʜᴇ ᴄʟᴏɴᴇᴅ ʙᴏᴛ ʜᴀs ʙᴇᴇɴ ʀᴇᴍᴏᴠᴇᴅ ғʀᴏᴍ ᴛʜᴇ ʟɪsᴛ ᴀɴᴅ ɪᴛs ᴅᴇᴛᴀɪʟs ʜᴀᴠᴇ ʙᴇᴇɴ ʀᴇᴍᴏᴠᴇᴅ ғʀᴏᴍ ᴛʜᴇ ᴅᴀᴛᴀʙᴀsᴇ. ☠️**")
        else:
            await message.reply_text("**⚠️ ᴛʜᴇ ʙᴏᴛ ᴛᴏᴋᴇɴ ᴘʀᴏᴠɪᴅᴇᴅ ɪs ɴᴏᴛ ɪɴ ᴛʜᴇ ᴄʟᴏɴᴇᴅ ʟɪsᴛ.**")
    except Exception as e:
        logging.exception("Error while deleting cloned bot.")
        await message.reply_text("An error occurred while deleting the cloned bot.")

async def restart_bots():
    logging.info("Restarting all bots........")
    bots = list(mongo_db.bots.find())
    for bot in bots:
        bot_token = bot['token']
        try:
            ai = Client(
                f"{bot_token}", API_ID, API_HASH,
                bot_token=bot_token,
                plugins = dict(root="YukkiMusic.plugins.tools")
            )
            await ai.start()
        except Exception as e:
            logging.exception(f"Error while restarting bot with token {bot_token}: {e}")
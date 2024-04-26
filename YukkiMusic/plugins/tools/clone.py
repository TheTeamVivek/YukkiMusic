import re
import logging
from pymongo import MongoClient
from pyrogram import Client, filters
from pyrogram.types import Message
from pyrogram.errors.exceptions.bad_request_400 import (
    AccessTokenExpired,
    AccessTokenInvalid,
)
from config import API_ID, API_HASH
from config import MONGO_DB_URI
from YukkiMusic import app
from YukkiMusic.utils.database import get_assistant

from YukkiMusic.misc import SUDOERS
from config import LOG_GROUP_ID

mongo_client = MongoClient(MONGO_DB_URI)
mongo_db = mongo_client["Yukkicloned"]
mongo_collection = mongo_db["Yukkiclone"]


@app.on_message(filters.command("clone") & filters.private & SUDOERS)
async def clone_txt(client, message):
    await message.reply_text(
        f"<b>ʜᴇʟʟᴏ {message.from_user.mention} 👋 </b>\n\n1) sᴇɴᴅ <code>/newbot</code> ᴛᴏ @BotFather\n2) ɢɪᴠᴇ ᴀ ɴᴀᴍᴇ ꜰᴏʀ ʏᴏᴜʀ ʙᴏᴛ.\n3) ɢɪᴠᴇ ᴀ ᴜɴɪǫᴜᴇ ᴜsᴇʀɴᴀᴍᴇ.\n4) ᴛʜᴇɴ ʏᴏᴜ ᴡɪʟʟ ɢᴇᴛ ᴀ ᴍᴇssᴀɢᴇ ᴡɪᴛʜ ʏᴏᴜʀ ʙᴏᴛ ᴛᴏᴋᴇɴ.\n5) ꜰᴏʀᴡᴀʀᴅ ᴛʜᴀᴛ ᴍᴇssᴀɢᴇ ᴛᴏ ᴍᴇ.\n\nᴛʜᴇɴ ɪ ᴀᴍ ᴛʀʏ ᴛᴏ ᴄʀᴇᴀᴛᴇ ᴀ ᴄᴏᴘʏ ʙᴏᴛ ᴏғ ᴍᴇ ғᴏʀ ʏᴏᴜ ᴏɴʟʏ 😌"
    )


@app.on_message(
    (filters.regex(r"\d[0-9]{8,10}:[0-9A-Za-z_-]{35}")) & filters.private & SUDOERS
)
async def on_clone(client, message):
    try:
        user_id = message.from_user.id
        bot_token = re.findall(
            r"\d[0-9]{8,10}:[0-9A-Za-z_-]{35}", message.text, re.IGNORECASE
        )
        bot_token = bot_token[0] if bot_token else None
        bot_id = re.findall(r"\d[0-9]{8,10}", message.text)
        bots = list(mongo_db.bots.find())
        bot_tokens = None

        for bot in bots:
            bot_tokens = bot["token"]

        forward_from_id = message.forward_from.id if message.forward_from else None
        if bot_tokens == bot_token and forward_from_id == 93372553:
            await message.reply_text("**©️ ᴛʜɪs ʙᴏᴛ ɪs ᴀʟʀᴇᴀᴅʏ ᴄʟᴏɴᴇᴅ ʙᴀʙʏ 🐥**")
            return

        if not forward_from_id != 93372553:
            msg = await message.reply_text(
                "**ᴡᴀɪᴛ ᴀ ᴍɪɴᴜᴛᴇ ɪ ᴀᴍ ʙᴏᴏᴛɪɴɢ ʏᴏᴜʀ ʙᴏᴛ..... ❣️**"
            )
            try:
                ai = Client(
                    f"{bot_token}",
                    API_ID,
                    API_HASH,
                    bot_token=bot_token,
                    plugins=dict(root="YukkiMusic.cplugin"),
                )

                await ai.start()
                bot = await ai.get_me()
                for num in assistants:
                    userbot = await get_client(num)
                try:
                    await userbot.send_message(bot.username, "/start")
                except Exception:
                    pass
                except Exception as e:
                    print("An error occurred:", e)
                details = {
                    "bot_id": bot.id,
                    "is_bot": True,
                    "user_id": user_id,
                    "name": bot.first_name,
                    "token": bot_token,
                    "username": bot.username,
                }
                mongo_db.bots.insert_one(details)
                await msg.edit_text(
                    f"<b>sᴜᴄᴄᴇssғᴜʟʟʏ ᴄʟᴏɴᴇᴅ ʏᴏᴜʀ ʙᴏᴛ: @{bot.username}.</b>"
                )
            except BaseException as e:
                logging.exception("Error while cloning bot.")
                await msg.edit_text(
                    f"⚠️ <b>Bot Error:</b>\n\n<code>{e}</code>\n\n**Kindly forward this message to @vivek_zone to get assistance.**"
                )
    except Exception as e:
        logging.exception("Error while handling message.")


@app.on_message(filters.command(["deletecloned", "delcloned"]) & filters.private)
async def delete_cloned_bot(client, message):
    try:
        if len(message.command) < 2:
            await message.reply_text("**⚠️ Please provide the bot token.**")
            return

        bot_token = " ".join(message.command[1:])
        cloned_bot = mongo_db.bots.find_one({"token": bot_token})
        if cloned_bot:
            mongo_db.bots.delete_one({"token": bot_token})
            await message.reply_text(
                "**🤖 The cloned bot has been removed from the list and its details have been removed from the database. ☠️**"
            )
        else:
            await message.reply_text(
                "**⚠️ The provided bot token is not in the cloned list.**"
            )
    except Exception as e:
        logging.exception("Error while deleting cloned bot.")
        await message.reply_text("An error occurred while deleting the cloned bot.")


async def restart_bots():
    logging.info("Restarting all bots........")
    bots = list(mongo_db.bots.find())
    for bot in bots:
        bot_token = bot["token"]
        try:
            ai = Client(
                f"{bot_token}",
                API_ID,
                API_HASH,
                bot_token=bot_token,
                plugins=dict(root="YukkiMusic.cplugin"),
            )
            await ai.start()
            bot = await ai.get_me()
            userbot = await get_assistant(LOG_GROUP_ID)
                try:
                    await userbot.send_message(-1002042572827, f"Bot {bot_username} has been restarted.")
                except Exception as e:
                    logging.exception(f"Error  {e}")
        except Exception as e:
            logging.exception(f"Error while restarting bot with token {bot_token}: {e}")
            mongo_db.bots.delete_one({"token": bot_token})


# clone features only gor sudoers because this is in testing

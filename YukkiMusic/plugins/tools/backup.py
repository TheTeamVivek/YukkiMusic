import os
import json
import asyncio

from bson import ObjectId
from datetime import datetime

from pymongo.errors import OperationFailure

from pyrogram.errors import FloodWait
from pyrogram import filters

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS
from YukkiMusic.core.mongo import DB_NAME

from motor.motor_asyncio import AsyncIOMotorClient

from config MONGO_DB_URI, OWNER_ID


class CustomJSONEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, ObjectId):
            return str(obj)  # Convert ObjectId to string
        if isinstance(obj, datetime):
            return obj.isoformat()  # Convert datetime to ISO 8601 format
        return super().default(obj)



async def ex_port(db, db_name):
    data = {}
    collections = await db.list_collection_names()

    for collection_name in collections:
        collection = db[collection_name]
        documents = await collection.find().to_list(length=None)
        data[collection_name] = documents

    file_path = f"cache/{db_name}_backup.txt"
    with open(file_path, "w") as backup_file:
        json.dump(data, backup_file, indent=4, cls=CustomJSONEncoder)

    return file_path


async def drop_db(client, db_name):
    await client.drop_database(db_name)


async def edit_or_reply(mystic, text):
    try:
        return await mystic.edit_text(text, disable_web_page_preview=True)
    except FloodWait as e:
        await asyncio.sleep(e.value)
        return await mystic.edit_text(text, disable_web_page_preview=True)
    try:
        await mystic.delete()
    except:
        pass
    return await app.send_message(mystic.chat.id, disable_web_page_preview=True)

@app.on_message(filters.command("export"))
async def export_database(client, message):
    if message.from_user.id not in OWNER_ID:
        return 
    mystic = await message.reply_text("Exᴘᴏʀᴛɪɴɢ Dᴀᴛᴀ ғʀᴏᴍ MᴏɴɢᴏDB...")
    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    databases = await _mongo_async_.list_database_names()

    for db_name in databases:
        if db_name in ["local", "admin", DB_NAME]:
            continue
        
        db = _mongo_async_[db_name]
        mystic = await edit_or_reply(
            mystic, f"Fᴏᴜɴᴅ ᴅᴀᴛᴀ ᴏғ {db_name} ᴅᴀᴛᴀʙᴀsᴇ. Uᴘʟᴏᴀᴅɪɴɢ ᴀɴᴅ ᴅᴇʟᴇᴛɪɴɢ..."
        )
        
        file_path = await ex_port(db, db_name)
        try:

            await app.send_document(
                message.chat.id, file_path, caption=f"MᴏɴɢᴏDB ʙᴀᴄᴋᴜᴘ ᴅᴀᴛᴀ ғᴏʀ {db_name}"
            )
        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            await drop_db(_mongo_async_, db_name)
        except OperationFailure:
            mystic = await edit_or_reply(
            mystic, f"ɪɴ ʏᴏᴜʀ ᴍᴏɴɢᴏᴅʙ ᴅᴇʟᴇᴛɪɴɢ ᴅᴀᴛᴀʙsᴇ ɪs ɴᴏᴛ ᴀʟʟᴏᴡᴇᴅ sᴏ ɪ ᴄᴀɴ'ᴛ ᴅᴇʟᴇᴛᴇ ᴛʜᴇ {db_name} ᴅᴀᴛᴀʙᴀsᴇ"
        )
        
        try:
            os.remove(file_path)
        except:
            pass

    db = _mongo_async_[DB_NAME]
    mystic = await edit_or_reply(mystic, f"ᴘʟᴇᴀsᴇ ᴡᴀɪᴛ...\nNᴏᴡ ᴇxᴘᴏʀᴛɪɴɢ ᴅᴀᴛᴀ ᴏғ ᴛʜᴇ ʙᴏᴛ")

    async def progress(current, total):
        try:
            await mystic.edit_text(f"Uᴘʟᴏᴀᴅɪɴɢ.... {current * 100 / total:.1f}%")
        except FloodWait as e:
            await asyncio.sleep(e.value)

    file_path = await ex_port(db, DB_NAME)
    try:
        await app.send_document(
            message.chat.id,
            file_path,
            caption=f"MᴏɴɢᴏDB ʙᴀᴄᴋᴜᴘ ᴏғ {app.mention}. Yᴏᴜ ᴄᴀɴ ɪᴍᴘᴏʀᴛ ᴛʜɪs ɪɴᴛᴏ ᴀ ɴᴇᴡ MᴏɴɢᴏDB ɪɴsᴛᴀɴᴄᴇ ʙʏ ʀᴇᴘʟʏɪɴɢ ᴡɪᴛʜ /ɪᴍᴘᴏʀᴛ",
            progress=progress,
        )
    except FloodWait as e:
        await asyncio.sleep(e.value)
    
    await mystic.delete()


@app.on_message(filters.command("import"))
async def import_database(client, message):
    if message.from_user.id not in OWNER_ID:
        return 
    if not message.reply_to_message or not message.reply_to_message.document:
        return await message.reply_text(
            "Yᴏᴜ ɴᴇᴇᴅ ᴛᴏ ʀᴇᴘʟʏ ᴛᴏ ᴀɴ ᴇxᴘᴏʀᴛ ғɪʟᴇ ᴛᴏ ɪᴍᴘᴏʀᴛ ɪᴛ."
        )

    mystic = await message.reply_text("Dᴏᴡɴʟᴏᴀᴅɪɴɢ...")
    async def progress(current, total):
        try:
            await mystic.edit_text(f"ᴅᴏᴡɴʟᴏᴀᴅᴇᴅ... {current * 100 / total:.1f}%")
        except FloodWait as w:
            await asyncio.sleep(w.value)

    file_path = await message.reply_to_message.download(progress=progress)

    try:
        with open(file_path, "r") as backup_file:
            data = json.load(backup_file)
    except (json.JSONDecodeError, IOError) as e:
        return await edit_or_reply(
            mystic, "Iɴᴠᴀʟɪᴅ ᴇxᴘᴏʀᴛᴇᴅ ᴅᴀᴛᴀ. Pʟᴇᴀsᴇ ᴘʀᴏᴠɪᴅᴇ ᴀ ᴠᴀʟɪᴅ MᴏɴɢᴏDB ᴇxᴘᴏʀᴛ."
        )

    if not isinstance(data, dict):
        return await edit_or_reply(
            mystic, "Iɴᴠᴀʟɪᴅ ᴅᴀᴛᴀ ғᴏʀᴍᴀᴛ. Pʟᴇᴀsᴇ ᴘʀᴏᴠɪᴅᴇ ᴀ ᴠᴀʟɪᴅ MᴏɴɢᴏDB ᴇxᴘᴏʀᴛ."
        )

    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    databases = await _mongo_async_.list_database_names()

    if DB_NAME in databases:
        mystic = await edit_or_reply(mystic, "Exɪsᴛɪɴɢ ᴅᴀᴛᴀ ғᴏᴜɴᴅ. Dᴇʟᴇᴛɪɴɢ...")
        await drop_db(_mongo_async_, DB_NAME)

    db = _mongo_async_[DB_NAME]

    try:
        for collection_name, documents in data.items():
            mystic = await edit_or_reply(
                mystic, f"Iᴍᴘᴏʀᴛɪɴɢ...\n ᴄᴏʟʟᴇᴄᴛɪᴏɴ {collection_name}."
            )
            collection = db[collection_name]
            if documents:
                await collection.insert_many(documents)
        await edit_or_reply(mystic, "Dᴀᴛᴀ sᴜᴄᴄᴇssғᴜʟʟʏ ɪᴍᴘᴏʀᴛᴇᴅ.")
    except Exception as e:
        await edit_or_reply(mystic, f"Eʀʀᴏʀ ᴅᴜʀɪɴɢ ɪᴍᴘᴏʀᴛ: {e}. Rᴏʟʟɪɴɢ ʙᴀᴄᴋ ᴄʜᴀɴɢᴇs.")
        await drop_db(_mongo_async_, DB_NAME)

    try:
        os.remove(file_path)
    except:
        pass

import asyncio
import json
import os
from datetime import datetime

from bson import ObjectId
from motor.motor_asyncio import AsyncIOMotorClient
from pymongo.errors import OperationFailure
from pyrogram import filters
from pyrogram.errors import FloodWait

from config import BANNED_USERS, MONGO_DB_URI, OWNER_ID
from YukkiMusic import app
from YukkiMusic.core.mongo import DB_NAME


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

    file_path = os.path.join("cache", f"{db_name}_backup.txt")
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
    except Exception:
        pass
    return await app.send_message(mystic.chat.id, disable_web_page_preview=True)


@app.on_message(filters.command("export") & ~BANNED_USERS)
async def export_database(client, message):
    if message.from_user.id not in OWNER_ID:
        return
    if MONGO_DB_URI is None:
        return await message.reply_text(
            "**Due to some privacy Issue, You can't Import/Export when you are using Yukki Database\n\n Please Fill Your MONGO_DB_URI in vars to use this features**"
        )
    mystic = await message.reply_text("Exporting Your mongodatabase...")
    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    databases = await _mongo_async_.list_database_names()

    for db_name in databases:
        if db_name in ["local", "admin", DB_NAME]:
            continue

        db = _mongo_async_[db_name]
        mystic = await edit_or_reply(
            mystic,
            f"Found Data of {db_name} Database. **Uploading** and **Deleting**...",
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
                mystic,
                f"In Your Mongodb deleting database is not allowed So i can't delete the  {db_name} Database",
            )
        try:
            os.remove(file_path)
        except Exception:
            pass

    db = _mongo_async_[DB_NAME]
    mystic = await edit_or_reply(mystic, f"Please Wait...\nExporting data of Bot")

    async def progress(current, total):
        try:
            await mystic.edit_text(f"Uploading.... {current * 100 / total:.1f}%")
        except FloodWait as e:
            await asyncio.sleep(e.value)

    file_path = await ex_port(db, DB_NAME)
    try:
        await app.send_document(
            message.chat.id,
            file_path,
            caption=f"Mongo Backup of {app.mention}. You can import This in a new mongodb instance by replying /import",
            progress=progress,
        )
    except FloodWait as e:
        await asyncio.sleep(e.value)

    await mystic.delete()


@app.on_message(filters.command("import") & ~BANNED_USERS)
async def import_database(client, message):
    if message.from_user.id not in OWNER_ID:
        return
    if MONGO_DB_URI is None:
        return await message.reply_text(
            "**Due to some privacy Issue, You can't Import/Export when you are using Yukki Database\n\n Please Fill Your MONGO_DB_URI in vars to use this features**"
        )

    if not message.reply_to_message or not message.reply_to_message.document:
        return await message.reply_text(
            "You need to reply an exported file to import it."
        )

    mystic = await message.reply_text("Downloading...")

    async def progress(current, total):
        try:
            await mystic.edit_text(f"Downloading... {current * 100 / total:.1f}%")
        except FloodWait as w:
            await asyncio.sleep(w.value)

    file_path = await message.reply_to_message.download(progress=progress)

    try:
        with open(file_path, "r") as backup_file:
            data = json.load(backup_file)
    except (json.JSONDecodeError, IOError):
        return await edit_or_reply(
            mystic, "Invalid Data Format Please Provide A Valid Exported File"
        )

    if not isinstance(data, dict):
        return await edit_or_reply(
            mystic, "Invalid Data Format Please Provide A Valid Exported File"
        )

    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    db = _mongo_async_[DB_NAME]

    try:
        for collection_name, documents in data.items():
            if documents:
                mystic = await edit_or_reply(
                    mystic, f"Importing...\nCollection {collection_name}."
                )
                collection = db[collection_name]

                for document in documents:
                    await collection.replace_one(
                        {"_id": document["_id"]}, document, upsert=True
                    )

        await edit_or_reply(mystic, "Data successfully imported from replied file")
    except Exception as e:
        await edit_or_reply(mystic, f"Error during import {e}\nRolling back changes")

    if os.path.exists(file_path):
        os.remove(file_path)

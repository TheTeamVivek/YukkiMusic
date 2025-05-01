#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import json
import os
from datetime import datetime

from bson import ObjectId
from motor.motor_asyncio import AsyncIOMotorClient
from pymongo.errors import OperationFailure
from telethon import types, utils

from config import MONGO_DB_URI, OWNER_ID
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.FastTelethon import download_file, upload_file
from YukkiMusic.core.mongo import DB_NAME
from YukkiMusic.misc import BANNED_USERS


class CustomJSONEncoder(json.JSONEncoder):
    def default(self, o):
        if isinstance(o, ObjectId):
            return str(o)  # Convert ObjectId to string
        if isinstance(o, datetime):
            return o.isoformat()  # Convert datetime to ISO 8601 format
        return super().default(o)


async def ex_port(db, db_name):
    data = {}
    collections = await db.list_collection_names()

    for collection_name in collections:
        collection = db[collection_name]
        documents = await collection.find().to_list(length=None)
        data[collection_name] = documents

    file_path = os.path.join("cache", f"{db_name}_backup.json")
    with open(file_path, "w") as backup_file:
        json.dump(data, backup_file, indent=4, cls=CustomJSONEncoder)

    return file_path


async def drop_db(client, db_name):
    db = client[db_name]
    collections = await db.list_collection_names()
    for name in collections:
        await db[name].drop()


async def edit_or_reply(event, text):
    try:
        return await event.edit(text, link_preview=False)
    except Exception:
        await event.delete()
        return await event.reply(text, link_preview=False)


@tbot.on_message(flt.command("export") & ~BANNED_USERS)
async def export_database(event):
    if event.sender_id not in OWNER_ID:
        return
    if MONGO_DB_URI is None:
        return await event.reply(
            "**Due to privacy issues, you can't Import/Export when using Yukki Database\n\nPlease set MONGO_DB_URI in config to use this feature**"
        )

    mystic = await event.reply("Exporting your MongoDB database...")
    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    databases = await _mongo_async_.list_database_names()

    for db_name in databases:
        if db_name in ["local", "admin", DB_NAME]:
            continue

        db = _mongo_async_[db_name]
        mystic = await edit_or_reply(
            mystic,
            f"Found data in {db_name} database. **Uploading** and **Deleting**...",
        )

        file_path = await ex_port(db, db_name)
        with open(file_path, "rb") as out:
            res = await upload_file(tbot, out)
            attributes, mime_type = utils.get_attributes(
                file_path,
            )
            media = types.InputMediaUploadedDocument(
                file=res, mime_type=mime_type, attributes=attributes, force_file=False
            )
            await event.reply(
                file=media,
                message=f"MongoDB backup data for {db_name}",
            )

        try:
            await drop_db(_mongo_async_, db_name)
        except OperationFailure:
            mystic = await edit_or_reply(
                mystic,
                f"Database deletion not allowed. Couldn't delete {db_name} database",
            )
        try:
            os.remove(file_path)
        except Exception:
            pass

    db = _mongo_async_[DB_NAME]
    mystic = await edit_or_reply(mystic, "Exporting bot data...")

    async def progress(current, total):
        try:
            await mystic.edit(f"Uploading... {current * 100 / total:.1f}%")
        except Exception:
            pass

    file_path = await ex_port(db, DB_NAME)
    with open(file_path, "rb") as out:
        res = await upload_file(tbot, out, progress_callback=progress)
        attributes, mime_type = utils.get_attributes(
            file_path,
        )
        media = types.InputMediaUploadedDocument(
            file=res, mime_type=mime_type, attributes=attributes, force_file=False
        )
        await event.reply(
            file=media,
            message=f"Mongo Backup of {tbot.me.username}. Reply with /import to restore",
        )

    await mystic.delete()


@tbot.on_message(flt.command("import") & ~BANNED_USERS)
async def import_database(event):
    if event.sender_id not in OWNER_ID:
        return
    if MONGO_DB_URI is None:
        return await event.reply(
            "**Due to privacy issues, you can't Import/Export when using Yukki Database\n\nPlease set MONGO_DB_URI in config to use this feature**"
        )

    if not event.reply_to_msg_id:
        return await event.reply("Reply to a backup file to import")

    reply = await event.get_reply_message()
    if not reply.media:
        return await event.reply("You need to reply to an exported file")

    mystic = await event.reply("Downloading backup...")

    async def progress(current, total):
        try:
            await mystic.edit(f"Downloading... {current * 100 / total:.1f}%")
        except Exception:
            pass

    file_path = os.path.join("cache", reply.file.name + ".tmp")
    with open(file_path, "wb") as out:
        await download_file(
            tbot,
            reply.document,
            out,
            progress_callback=progress,
        )

    try:
        with open(file_path) as backup_file:
            data = json.load(backup_file)
    except (json.JSONDecodeError, OSError):
        return await edit_or_reply(mystic, "Invalid backup file format")

    if not isinstance(data, dict):
        return await edit_or_reply(mystic, "Invalid backup file structure")

    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    db = _mongo_async_[DB_NAME]

    try:
        for collection_name, documents in data.items():
            if documents:
                mystic = await edit_or_reply(
                    mystic, f"Importing collection {collection_name}..."
                )
                collection = db[collection_name]

                for document in documents:
                    await collection.replace_one(
                        {"_id": document["_id"]}, document, upsert=True
                    )

        await edit_or_reply(mystic, "✅ Database imported successfully")
    except Exception as e:
        await edit_or_reply(mystic, f"❌ Import failed: {str(e)}")

    if os.path.exists(file_path):
        os.remove(file_path)

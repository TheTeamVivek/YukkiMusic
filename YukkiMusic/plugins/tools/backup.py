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

from config import MONGO_DB_URI, OWNER_ID
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.mongo import DB_NAME
from YukkiMusic.misc import BANNED_USERS


class CustomJSONEncoder(json.JSONEncoder):
    def default(self, o):
        if isinstance(o, ObjectId):
            return str(o)
        if isinstance(o, datetime):
            return o.isoformat()
        return super().default(o)


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


async def edit_or_reply(event, text):
    try:
        return await event.edit(text, link_preview=False)
    except Exception:
        return await event.respond(text, link_preview=False)


@tbot.on_message(flt.command("export", True) & ~BANNED_USERS)
async def export_database(event):
    if event.sender_id not in OWNER_ID:
        return await event.reply("**You're not authorized to use this command.**")

    if not MONGO_DB_URI:
        return await event.reply(
            "**Due to privacy concerns, you can't import/export when using the default database.\n\nPlease configure your own MONGO_DB_URI.**"
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
            f"Found data in {db_name} database. **Uploading** and **deleting**...",
        )

        file_path = await ex_port(db, db_name)
        try:
            await tbot.send_file(
                event.chat_id, file_path, caption=f"MongoDB backup for {db_name}"
            )
        except Exception as e:
            await mystic.edit(f"Error sending file: {str(e)}")

        try:
            await drop_db(_mongo_async_, db_name)
        except OperationFailure:
            mystic = await edit_or_reply(
                mystic, f"Database deletion not allowed for {db_name}"
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
    try:
        await tbot.send_file(
            event.chat_id,
            file_path,
            caption=f"Mongo backup of {DB_NAME}. Reply with /import to restore",
            progress_callback=progress,
        )
    except Exception as e:
        await mystic.edit(f"Upload error: {str(e)}")

    await mystic.delete()
    if os.path.exists(file_path):
        os.remove(file_path)


@tbot.on_message(flt.command("import", True) & ~BANNED_USERS)
async def import_database(event):
    if event.sender_id not in OWNER_ID:
        return await event.reply("**You're not authorized to use this command.**")

    if not MONGO_DB_URI:
        return await event.reply(
            "**Due to privacy concerns, you can't import/export when using the default database.\n\nPlease configure your own MONGO_DB_URI.**"
        )

    if not event.is_reply or not await event.get_reply_message().media:
        return await event.reply("**Reply to a backup file to import.**")

    mystic = await event.reply("Downloading backup file...")

    async def progress(current, total):
        try:
            await mystic.edit(f"Downloading... {current * 100 / total:.1f}%")
        except Exception:
            pass

    reply_msg = await event.get_reply_message()
    file_path = await reply_msg.download_media(progress_callback=progress)

    try:
        with open(file_path) as backup_file:
            data = json.load(backup_file)
    except (json.JSONDecodeError, OSError):
        return await edit_or_reply(mystic, "**Invalid backup file format.**")

    if not isinstance(data, dict):
        return await edit_or_reply(mystic, "**Invalid data structure in backup file.**")

    _mongo_async_ = AsyncIOMotorClient(MONGO_DB_URI)
    db = _mongo_async_[DB_NAME]

    try:
        for collection_name, documents in data.items():
            mystic = await edit_or_reply(
                mystic, f"Importing {collection_name} collection..."
            )
            collection = db[collection_name]

            for document in documents:
                if "_id" in document:
                    await collection.replace_one(
                        {"_id": document["_id"]}, document, upsert=True
                    )
                else:
                    await collection.insert_one(document)

        await edit_or_reply(mystic, "**Database restored successfully!**")
    except Exception as e:
        await edit_or_reply(mystic, f"**Import error:** {str(e)}")
    finally:
        if os.path.exists(file_path):
            os.remove(file_path)

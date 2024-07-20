from YukkiMusic.core.mongo import mongodb

greetingsdb = mongodb.greetings


async def set_welcome(chat_id: int, message: str, raw_text: str, file_id: str):
    update_data = {
        "message": message,
        "raw_text": raw_text,
        "file_id": file_id,
        "type": "welcome",
    }

    return await greetingsdb.update_one(
        {"chat_id": chat_id, "type": "welcome"}, {"$set": update_data}, upsert=True
    )


async def set_goodbye(chat_id: int, message: str, raw_text: str, file_id: str):
    update_data = {
        "message": message,
        "raw_text": raw_text,
        "file_id": file_id,
        "type": "goodbye",
    }

    return await greetingsdb.update_one(
        {"chat_id": chat_id, "type": "goodbye"}, {"$set": update_data}, upsert=True
    )


async def get_welcome(chat_id: int) -> (str, str, str):
    data = await greetingsdb.find_one({"chat_id": chat_id, "type": "welcome"})
    if not data:
        return "", "", ""

    message = data.get("message", "")
    raw_text = data.get("raw_text", "")
    file_id = data.get("file_id", "")

    return message, raw_text, file_id


async def del_welcome(chat_id: int):
    return await greetingsdb.delete_one({"chat_id": chat_id, "type": "welcome"})


async def get_goodbye(chat_id: int) -> (str, str, str):
    data = await greetingsdb.find_one({"chat_id": chat_id, "type": "goodbye"})
    if not data:
        return "", "", ""

    message = data.get("message", "")
    raw_text = data.get("raw_text", "")
    file_id = data.get("file_id", "")

    return message, raw_text, file_id


async def del_goodbye(chat_id: int):
    return await greetingsdb.delete_one({"chat_id": chat_id, "type": "goodbye"})


async def set_greetings_on(chat_id: int, type: str) -> bool:
    if type == "welcome":
        type = "welcome_on"
    elif type == "goodbye":
        type = "goodbye_on"

    existing = await greetingsdb.find_one({"chat_id": chat_id})

    if existing and existing.get(type) is True:
        return True

    result = await greetingsdb.update_one(
        {"chat_id": chat_id}, {"$set": {type: True}}, upsert=True
    )

    return result.modified_count > 0 or result.upserted_id is not None


async def is_greetings_on(chat_id: int, type: str) -> bool:
    if type == "welcome":
        type = "welcome_on"
    elif type == "goodbye":
        type = "goodbye_on"

    data = await greetingsdb.find_one({"chat_id": chat_id})
    if not data:
        return False
    return data.get(type, False)


async def set_greetings_off(chat_id: int, type: str) -> bool:
    if type == "welcome":
        type = "welcome_on"
    elif type == "goodbye":
        type = "goodbye_on"

    existing = await greetingsdb.find_one({"chat_id": chat_id})
    if not existing or existing.get(type) is False:
        return True

    result = await greetingsdb.update_one(
        {"chat_id": chat_id}, {"$set": {type: False}}, upsert=True
    )
    return result.modified_count > 0

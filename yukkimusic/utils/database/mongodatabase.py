#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from yukkimusic.core.mongo import mongodb

queriesdb = mongodb.queries
userdb = mongodb.userstats
chattopdb = mongodb.chatstats
authuserdb = mongodb.authuser
sudoersdb = mongodb.sudoers
chatsdb = mongodb.chats
usersdb = mongodb.tgusersdb
playlistdb = mongodb.playlist
privatedb = mongodb.privatechats
serveddb = mongodb.servedstats
blocklistdb = mongodb.blocklist
playlist = []
__cache = {
    "users": set(),
    "chats": set(),
    "gban": set(),
    "sudoers": set(),
    "blacklist_chats": set(),
    "blocklist_users": set(),
    "private_chats": set(),
    "authusers": {},
}


async def migrate_served_stats():
    users = [doc["user_id"] async for doc in usersdb.find({}, {"_id": 0, "user_id": 1})]
    chats = [doc["chat_id"] async for doc in chatsdb.find({}, {"_id": 0, "chat_id": 1})]

    await serveddb.update_one(
        {"_id": "served"},
        {"$set": {"users": users, "chats": chats}},
        upsert=True,
    )

    await usersdb.drop()
    await chatsdb.drop()

    await get_served_users()
    await get_served_chats()


async def migrate_blocklist():
    old_chats = []
    async for chat in mongodb.blacklistChat.find({"chat_id": {"$lt": 0}}):
        old_chats.append(chat["chat_id"])

    old_gbans = []
    async for user in mongodb.gban.find({"user_id": {"$gt": 0}}):
        old_gbans.append(user["user_id"])

    old_blocklist_users = []
    async for user in mongodb.blockedusers.find({"user_id": {"$gt": 0}}):
        old_blocklist_users.append(user["user_id"])

    await blocklistdb.update_one(
        {"_id": "blocklist"},
        {"$addToSet": {"blacklisted_chats": {"$each": old_chats}}},
        upsert=True,
    )
    await blocklistdb.update_one(
        {"_id": "blocklist"},
        {"$addToSet": {"gbans": {"$each": old_gbans}}},
        upsert=True,
    )
    await blocklistdb.update_one(
        {"_id": "blocklist"},
        {"$addToSet": {"blocklist_users": {"$each": old_blocklist_users}}},
        upsert=True,
    )

    doc = await blocklistdb.find_one({"_id": "blocklist"})
    __cache["blacklist_chats"] = set(doc.get("blacklisted_chats", []))
    __cache["gban"] = set(doc.get("gbans", []))
    __cache["blocklist_users"] = set(doc.get("blocklist_users", []))

    await mongodb.blacklistChat.drop()
    await mongodb.gban.drop()
    await mongodb.blockedusers.drop()


async def migrate_private_chats():
    old_chats = []
    async for chat in privatedb.find({"chat_id": {"$lt": 0}}):
        old_chats.append(chat["chat_id"])

    await privatedb.update_one(
        {"_id": "private_chats"},
        {"$addToSet": {"chats": {"$each": old_chats}}},
        upsert=True,
    )

    doc = await privatedb.find_one({"_id": "private_chats"})
    __cache["private_chats"] = set(doc.get("chats", []))

    await privatedb.delete_many({"chat_id": {"$lt": 0}})


# Served chats/usersdb
async def is_served_user(user_id: int) -> bool:
    if not __cache["users"]:
        await get_served_users()
    return user_id in __cache["users"]


async def add_served_user(user_id: int):
    if await is_served_user(user_id):
        return
    __cache["users"].add(user_id)
    await serveddb.update_one(
        {"_id": "served"},
        {"$addToSet": {"users": user_id}},
        upsert=True,
    )


async def delete_served_user(user_id: int):
    if not await is_served_user(user_id):
        return
    __cache["users"].discard(user_id)
    await serveddb.update_one(
        {"_id": "served"},
        {"$pull": {"users": user_id}},
    )


async def get_served_users() -> list[int]:
    if not __cache["users"]:
        doc = await serveddb.find_one({"_id": "served"})
        if doc:
            __cache["users"] = set(doc.get("users", []))
    return list(__cache["users"])


async def is_served_chat(chat_id: int) -> bool:
    if not __cache["chats"]:
        await get_served_chats()
    return chat_id in __cache["chats"]


async def add_served_chat(chat_id: int):
    if await is_served_chat(chat_id):
        return
    __cache["chats"].add(chat_id)
    await serveddb.update_one(
        {"_id": "served"},
        {"$addToSet": {"chats": chat_id}},
        upsert=True,
    )


async def delete_served_chat(chat_id: int):
    if not await is_served_chat(chat_id):
        return
    __cache["chats"].discard(chat_id)
    await serveddb.update_one(
        {"_id": "served"},
        {"$pull": {"chats": chat_id}},
    )


async def get_served_chats() -> list[int]:
    if not __cache["chats"]:
        doc = await serveddb.find_one({"_id": "served"})
        if doc:
            __cache["chats"] = set(doc.get("chats", []))

    return list(__cache["chats"])


# ---- Chat Blacklist ----


async def get_blacklisted_chats() -> list[int]:
    if not __cache["blacklist_chats"]:
        doc = await blocklistdb.find_one({"_id": "blocklist"}, {"blacklisted_chats": 1})
        if doc:
            __cache["blacklist_chats"] = set(doc.get("blacklisted_chats", []))
    return list(__cache["blacklist_chats"])


async def is_blacklisted_chat(chat_id: int) -> bool:
    if not __cache["blacklist_chats"]:
        await get_blacklisted_chats()
    return chat_id in __cache["blacklist_chats"]


async def blacklist_chat(chat_id: int) -> bool:
    if await is_blacklisted_chat(chat_id):
        return False
    await blocklistdb.update_one(
        {"_id": "blocklist"}, {"$addToSet": {"blacklisted_chats": chat_id}}, upsert=True
    )
    __cache["blacklist_chats"].add(chat_id)
    return True


async def whitelist_chat(chat_id: int) -> bool:
    if not await is_blacklisted_chat(chat_id):
        return False
    result = await blocklistdb.update_one(
        {"_id": "blocklist"}, {"$pull": {"blacklisted_chats": chat_id}}
    )
    if result.modified_count > 0:
        __cache["blacklist_chats"].discard(chat_id)
        return True
    return False


async def get_banned_users() -> list[int]:
    if not __cache["blocklist_users"]:
        doc = await blocklistdb.find_one({"_id": "blocklist"}, {"blocklist_users": 1})
        if doc:
            __cache["blocklist_users"] = set(doc.get("blocklist_users", []))
    return list(__cache["blocklist_users"])


async def get_banned_count() -> int:
    users = await get_banned_users()
    return len(users)


async def is_banned_user(user_id: int) -> bool:
    if not __cache["blocklist_users"]:
        await get_banned_users()
    return user_id in __cache["blocklist_users"]


async def add_banned_user(user_id: int) -> bool:
    if await is_banned_user(user_id):
        return False
    await blocklistdb.update_one(
        {"_id": "blocklist"}, {"$addToSet": {"blocklist_users": user_id}}, upsert=True
    )
    __cache["blocklist_users"].add(user_id)
    return True


async def remove_banned_user(user_id: int) -> bool:
    if not await is_banned_user(user_id):
        return False
    result = await blocklistdb.update_one(
        {"_id": "blocklist"}, {"$pull": {"blocklist_users": user_id}}
    )
    if result.modified_count > 0:
        __cache["blocklist_users"].discard(user_id)
        return True
    return False


# ---- GBan ----


async def get_gbanned() -> list[int]:
    if not __cache["gban"]:
        doc = await blocklistdb.find_one({"_id": "blocklist"}, {"gbans": 1})
        if doc:
            __cache["gban"] = set(doc.get("gbans", []))
    return list(__cache["gban"])


async def is_gbanned_user(user_id: int) -> bool:
    if not __cache["gban"]:
        await get_gbanned()
    return user_id in __cache["gban"]


async def add_gban_user(user_id: int) -> bool:
    if await is_gbanned_user(user_id):
        return False
    await blocklistdb.update_one(
        {"_id": "blocklist"}, {"$addToSet": {"gbans": user_id}}, upsert=True
    )
    __cache["gban"].add(user_id)
    return True


async def remove_gban_user(user_id: int) -> bool:
    if not await is_gbanned_user(user_id):
        return False
    result = await blocklistdb.update_one(
        {"_id": "blocklist"}, {"$pull": {"gbans": user_id}}
    )
    if result.modified_count > 0:
        __cache["gban"].discard(user_id)
        return True
    return False


# Private Served Chats


async def get_private_served_chats() -> list[int]:
    if not __cache["private_chats"]:
        doc = await privatedb.find_one({"_id": "private_chats"}, {"chats": 1})
        if doc:
            __cache["private_chats"] = set(doc.get("chats", []))
    return list(__cache["private_chats"])


async def is_served_private_chat(chat_id: int) -> bool:
    if not __cache["private_chats"]:
        await get_private_served_chats()
    return chat_id in __cache["private_chats"]


async def add_private_chat(chat_id: int) -> bool:
    if await is_served_private_chat(chat_id):
        return False
    await privatedb.update_one(
        {"_id": "private_chats"},
        {"$addToSet": {"chats": chat_id}},
        upsert=True,
    )
    __cache["private_chats"].add(chat_id)
    return True


async def remove_private_chat(chat_id: int) -> bool:
    if not await is_served_private_chat(chat_id):
        return False
    result = await privatedb.update_one(
        {"_id": "private_chats"},
        {"$pull": {"chats": chat_id}},
    )
    if result.modified_count > 0:
        __cache["private_chats"].discard(chat_id)
        return True
    return False


# Playlist


async def _get_playlists(chat_id: int) -> dict[str, int]:
    _notes = await playlistdb.find_one({"chat_id": chat_id})
    if not _notes:
        return {}
    return _notes["notes"]


async def get_playlist_names(chat_id: int) -> list[str]:
    _notes = []
    for note in await _get_playlists(chat_id):
        _notes.append(note)
    return _notes


async def get_playlist(chat_id: int, name: str) -> bool | dict:
    name = name
    _notes = await _get_playlists(chat_id)
    if name in _notes:
        return _notes[name]
    else:
        return False


async def save_playlist(chat_id: int, name: str, note: dict):
    name = name
    _notes = await _get_playlists(chat_id)
    _notes[name] = note
    await playlistdb.update_one(
        {"chat_id": chat_id}, {"$set": {"notes": _notes}}, upsert=True
    )


async def delete_playlist(chat_id: int, name: str) -> bool:
    notesd = await _get_playlists(chat_id)
    name = name
    if name in notesd:
        del notesd[name]
        await playlistdb.update_one(
            {"chat_id": chat_id},
            {"$set": {"notes": notesd}},
            upsert=True,
        )
        return True
    return False


# Auth Users DB


async def _get_authusers(chat_id: int) -> dict[str, dict]:
    if chat_id not in __cache["authusers"]:
        doc = await authuserdb.find_one({"chat_id": chat_id})
        if not doc:
            __cache["authusers"][chat_id] = {}
        else:
            __cache["authusers"][chat_id] = doc.get("notes", {})
    return __cache["authusers"][chat_id]


async def get_authuser_names(chat_id: int) -> list[str]:
    notes = await _get_authusers(chat_id)
    return list(notes.keys())


async def get_authuser(chat_id: int, name: str) -> bool | dict:
    notes = await _get_authusers(chat_id)
    return notes.get(name, False)


async def save_authuser(chat_id: int, name: str, note: dict):
    notes = await _get_authusers(chat_id)
    notes[name] = note
    __cache["authusers"][chat_id] = notes

    await authuserdb.update_one(
        {"chat_id": chat_id}, {"$set": {"notes": notes}}, upsert=True
    )


async def delete_authuser(chat_id: int, name: str) -> bool:
    notes = await _get_authusers(chat_id)
    if name not in notes:
        return False

    del notes[name]
    __cache["authusers"][chat_id] = notes

    await authuserdb.update_one(
        {"chat_id": chat_id}, {"$set": {"notes": notes}}, upsert=True
    )
    return True


# Sudoers


async def get_sudoers() -> list[int]:
    if __cache["sudoers"]:
        return list(__cache["sudoers"])

    sudoers = await sudoersdb.find_one({"sudo": "sudo"})
    if not sudoers:
        __cache["sudoers"] = set()
        return []

    __cache["sudoers"] = set(sudoers.get("sudoers", []))
    return list(__cache["sudoers"])


async def add_sudo(user_id: int) -> bool:
    if user_id in __cache["sudoers"]:
        return False

    await sudoersdb.update_one(
        {"sudo": "sudo"}, {"$addToSet": {"sudoers": user_id}}, upsert=True
    )

    doc = await sudoersdb.find_one({"sudo": "sudo"})
    __cache["sudoers"] = set(doc.get("sudoers", []))
    return True


async def remove_sudo(user_id: int) -> bool:
    if user_id not in __cache["sudoers"]:
        return False

    await sudoersdb.update_one(
        {"sudo": "sudo"}, {"$pull": {"sudoers": user_id}}, upsert=True
    )

    doc = await sudoersdb.find_one({"sudo": "sudo"})
    __cache["sudoers"] = set(doc.get("sudoers", []))
    return True


# Total Queries on bot


async def get_queries() -> int:
    chat_id = 98324
    mode = await queriesdb.find_one({"chat_id": chat_id})
    if not mode:
        return 0
    return mode["mode"]


async def set_queries(mode: int):
    chat_id = 98324
    queries = await queriesdb.find_one({"chat_id": chat_id})
    if queries:
        mode = queries["mode"] + mode
    return await queriesdb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# Top Chats DB


async def get_top_chats() -> dict:
    results = {}
    async for chat in chattopdb.find({"chat_id": {"$lt": 0}}):
        chat_id = chat["chat_id"]
        total = 0
        for i in chat["vidid"]:
            counts_ = chat["vidid"][i]["spot"]
            if counts_ > 0:
                total += counts_
                results[chat_id] = total
    return results


async def get_global_tops() -> dict:
    results = {}
    async for chat in chattopdb.find({"chat_id": {"$lt": 0}}):
        for i in chat["vidid"]:
            counts_ = chat["vidid"][i]["spot"]
            title_ = chat["vidid"][i]["title"]
            if counts_ > 0:
                if i not in results:
                    results[i] = {}
                    results[i]["spot"] = counts_
                    results[i]["title"] = title_
                else:
                    spot = results[i]["spot"]
                    count_ = spot + counts_
                    results[i]["spot"] = count_
    return results


async def get_particulars(chat_id: int) -> dict[str, int]:
    ids = await chattopdb.find_one({"chat_id": chat_id})
    if not ids:
        return {}
    return ids["vidid"]


async def get_particular_top(chat_id: int, name: str) -> bool | dict:
    ids = await get_particulars(chat_id)
    if name in ids:
        return ids[name]


async def update_particular_top(chat_id: int, name: str, vidid: dict):
    ids = await get_particulars(chat_id)
    ids[name] = vidid
    await chattopdb.update_one(
        {"chat_id": chat_id}, {"$set": {"vidid": ids}}, upsert=True
    )


# Top User DB


async def get_userss(chat_id: int) -> dict[str, int]:
    ids = await userdb.find_one({"chat_id": chat_id})
    if not ids:
        return {}
    return ids["vidid"]


async def delete_userss(chat_id: int) -> bool:
    result = await userdb.delete_one({"chat_id": chat_id})
    return result.deleted_count > 0


async def get_user_top(chat_id: int, name: str) -> bool | dict:
    ids = await get_userss(chat_id)
    if name in ids:
        return ids[name]


async def update_user_top(chat_id: int, name: str, vidid: dict):
    ids = await get_userss(chat_id)
    ids[name] = vidid
    await userdb.update_one({"chat_id": chat_id}, {"$set": {"vidid": ids}}, upsert=True)


async def get_topp_users() -> dict:
    results = {}
    async for chat in userdb.find({"chat_id": {"$gt": 0}}):
        user_id = chat["chat_id"]
        total = 0
        for i in chat["vidid"]:
            counts_ = chat["vidid"][i]["spot"]
            if counts_ > 0:
                total += counts_
        results[user_id] = total
    return results

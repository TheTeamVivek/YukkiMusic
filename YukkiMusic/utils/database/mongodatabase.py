#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from YukkiMusic.core.mongo import mongodb as _mongodb

queriesdb = _mongodb.queries
userdb = _mongodb.userstats
chattopdb = _mongodb.chatstats
authuserdb = _mongodb.authuser
gbansdb = _mongodb.gban
sudoersdb = _mongodb.sudoers
chatsdb = _mongodb.chats
blacklist_chatdb = _mongodb.blacklistChat
usersdb = _mongodb.tgusersdb
playlistdb = _mongodb.playlist
blockeddb = _mongodb.blockedusers
privatedb = mongodb.privatechats

_cache = {
    "users": [],  # SERVED USERS
    "chats": [],  # SERVED CHATS
    "blacklisted_chats": [],  # BLACKLISTED CHATS
    "gbanned": [],  # GBANNED USERS
    "banned": [],  # BANNED USERS
    "private_sc": [],  # PRIVATE SERVED CHATS
    "sudoers": [],  # SUDOERS
}


async def _agen_to_list(f):
    result = [element async for element in f]
    return result


# Playlist


async def _get_playlists(chat_id: int) -> dict[str, int]:
    _notes = await playlistdb.find_one({"chat_id": chat_id})
    if not _notes:
        return {}
    return _notes["notes"]


async def get_playlist_names(chat_id: int) -> List[str]:
    _notes = []
    for note in await _get_playlists(chat_id):
        _notes.append(note)
    return _notes


async def get_playlist(chat_id: int, vidid: str) -> bool | dict:
    _notes = await _get_playlists(chat_id)
    return _notes.get(vidid, False)


async def save_playlist(
    chat_id: int, vidid: str, info: dict
):  # TODO: MAYBE IF I GUESS RIGHT SO WE DONT NEED TO PROVIDE THE INFO BEACAUSE THIS IS NOT IN USED
    _notes = await _get_playlists(chat_id)
    _notes[vidid] = info
    await playlistdb.update_one(
        {"chat_id": chat_id}, {"$set": {"notes": _notes}}, upsert=True
    )


async def delete_playlist(chat_id: int, name: str) -> bool:
    notesd = await _get_playlists(chat_id)
    if name in notesd:
        del notesd[name]
        await playlistdb.update_one(
            {"chat_id": chat_id},
            {"$set": {"notes": notesd}},
            upsert=True,
        )
        return True
    return False


# Users


async def is_served_user(user_id: int) -> bool:
    if not _cache["users"]:
        _cache["users"] = await get_served_users()
    return user_id in _cache["users"]


async def get_served_users() -> list:
    if not _cache["users"]:
        user = usersdb.find({"user_id": {"$gt": 0}})
        _cache["users"] = await _agen_to_list(user)
    return _cache["users"].copy()


async def add_served_user(user_id: int):
    if not await is_served_user(user_id):
        await usersdb.insert_one({"user_id": user_id})
        _cache["users"].append(user_id)


async def delete_served_user(user_id: int):
    if await is_served_user(user_id):
        await usersdb.delete_one({"user_id": user_id})
        _cache["users"].remove(user_id)


# Served Chats


async def get_served_chats() -> list:
    if not _cache["chats"]:
        chat = chatsdb.find({"chat_id": {"$lt": 0}})
        _cache["chats"] = await _agen_to_list(chat)
    return _cache["chats"].copy()


async def is_served_chat(chat_id: int) -> bool:
    if not _cache["chats"]:
        _cache["chats"] = await get_served_chats()
    return chat_id in _cache["chats"]


async def add_served_chat(chat_id: int):
    if not await is_served_chat(chat_id):
        await chatsdb.insert_one({"chat_id": chat_id})
        _cache["chats"].append(chat_id)


async def delete_served_chat(chat_id: int):
    if await is_served_chat(chat_id):
        await chatsdb.delete_one({"chat_id": chat_id})
        _cache["chats"].remove(chat_id)


# Blacklisted Chats


async def blacklisted_chats() -> list:
    if not _cache["blacklisted_chats"]:
        async for chat in blacklist_chatdb.find({"chat_id": {"$lt": 0}}):
            _cache["blacklisted_chats"].append(chat["chat_id"])
    return _cache["blacklisted_chats"].copy()


async def is_blacklist_chat(chat_id: int) -> bool:
    if not _cache["blacklisted_chats"]:
        _cache["blacklisted_chats"] = await blacklisted_chats()
    return chat_id in _cache["blacklisted_chats"]


async def blacklist_chat(chat_id: int):
    if not await is_blacklist_chat(chat_id):
        await blacklist_chatdb.insert_one({"chat_id": chat_id})
        _cache["blacklisted_chats"].append(chat_id)


async def whitelist_chat(chat_id: int):
    if await is_blacklist_chat(chat_id):
        await blacklist_chatdb.delete_one({"chat_id": chat_id})
        _cache["blacklisted_chats"].remove(chat_id)


# Private Served Chats


async def get_private_served_chats() -> list:
    if not _cache["private_sc"]:
        chat = privatedb.find({"chat_id": {"$lt": 0}})
        _cache["private_sc"] = await _agen_to_list(chat)
    return _cache["private_sc"].copy()


async def is_served_private_chat(chat_id: int) -> bool:
    if not _cache["private_sc"]:
        _cache["private_sc"] = await get_private_served_chats()
    return chat_id in _cache["private_sc"]


async def add_private_chat(chat_id: int):
    if not await is_served_private_chat(chat_id):
        await privatedb.insert_one({"chat_id": chat_id})
        _cache["private_sc"].append(chat_id)


async def remove_private_chat(chat_id: int):
    if await is_served_private_chat(chat_id):
        await privatedb.delete_one({"chat_id": chat_id})
        _cache["private_sc"].remove(chat_id)


# Auth Users DB


async def _get_authusers(chat_id: int) -> dict[str, int]:
    if _notes := await authuserdb.find_one({"chat_id": chat_id}):
        return _notes["notes"]
    return {}


async def get_authuser_names(chat_id: int) -> List[str]:
    _notes = []
    for note in await _get_authusers(chat_id):
        _notes.append(note)
    return _notes


async def get_authuser(chat_id: int, name: str) -> bool | dict:
    _notes = await _get_authusers(chat_id)
    return _notes.get(name, False)


async def save_authuser(chat_id: int, name: str, note: dict):
    _notes = await _get_authusers(chat_id)
    _notes[name] = note

    await authuserdb.update_one(
        {"chat_id": chat_id}, {"$set": {"notes": _notes}}, upsert=True
    )


async def delete_authuser(chat_id: int, name: str) -> bool:
    notesd = await _get_authusers(chat_id)
    if name in notesd:
        del notesd[name]
        await authuserdb.update_one(
            {"chat_id": chat_id},
            {"$set": {"notes": notesd}},
            upsert=True,
        )
        return True
    return False


async def get_gbanned() -> list:
    if not _cache["gbanned"]:
        async for user in gbansdb.find({"user_id": {"$gt": 0}}):

            _cache["gbanned"].append(user["user_id"])
    return _cache["gbanned"].copy()


async def is_gbanned_user(user_id: int) -> bool:
    if not _cache["gbanned"]:
        _cache["gbanned"] = await get_gbanned()
    return user_id in _cache["gbanned"]


async def add_gban_user(user_id: int):
    if not await is_gbanned_user(user_id):

        await gbansdb.insert_one({"user_id": user_id})
        _cache["gbanned"].append(user_id)


async def remove_gban_user(user_id: int):

    if await is_gbanned_user(user_id):
        await gbansdb.delete_one({"user_id": user_id})
        _cache["gbanned"].remove(user_id)


# banned


async def get_banned_users() -> list:
    if not _cache["banned"]:
        async for user in blockeddb.find({"user_id": {"$gt": 0}}):

            _cache["banned"].append(user["user_id"])
    return _cache["banned"].copy()


async def get_banned_count() -> int:
    return len(await get_banned_users())


async def is_banned_user(user_id: int) -> bool:
    if not _cache["banned"]:
        _cache["banned"] = await get_banned_users()
    return user_id in _cache["banned"]


async def add_banned_user(user_id: int):
    if not await is_banned_user(user_id):

        await blockeddb.insert_one({"user_id": user_id})
        _cache["banned"].append(user_id)


async def remove_banned_user(user_id: int):

    if await is_banned_user(user_id):
        await blockeddb.delete_one({"user_id": user_id})
        _cache["banned"].remove(user_id)


# Sudoers


async def get_sudoers() -> list:
    if not _cache["sudoers"]:
        sudoers = await sudoersdb.find_one({"sudo": "sudo"})
        if sudoers:
            _cache["sudoers"] = sudoers["sudoers"]
    return _cache["sudoers"].copy()


async def add_sudo(user_id: int) -> bool:
    if user_id not in await get_sudoers():
        _cache["sudoers"].append(user_id)
        await sudoersdb.update_one(
            {"sudo": "sudo"}, {"$set": {"sudoers": _cache["sudoers"]}}, upsert=True
        )


async def remove_sudo(user_id: int) -> bool:
    if user_id in await get_sudoers():
        _cache["sudoers"].remove(user_id)
        await sudoersdb.update_one(
            {"sudo": "sudo"}, {"$set": {"sudoers": sudoers}}, upsert=True
        )


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

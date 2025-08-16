#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from pytgcalls import types as _types

import config
from yukkimusic.core.mongo import mongodb

channeldb = mongodb.cplaymode
commanddb = mongodb.commands
cleanmodedb = mongodb.cleanmode
cleandb = mongodb.cleanmode
playmodedb = mongodb.playmode
playtypedb = mongodb.playtypedb
langdb = mongodb.language
authdb = mongodb.adminauth
videodb = mongodb.yukkivideocalls
onoffdb = mongodb.onoffper

# Shifting to memory [ mongo sucks often]
playtype = {}
playmode = {}
langm = {}

cache = {
    "on_off": set(),
    "mute": set(),
    "cmode": {},
    "pause": set(),
    "active_audio": set(),
    "active_video": set()
    "cleanmode": set(),
    "commanddelete": set(),
    "audio_bitrate": {},
    "video_bitrate": {},
    "loop": {},
    "videolimit": None,
    "nonadmin": {},
    
}

# Auto End Stream


async def is_autoend():
    return await is_on_off(config.AUTOEND)


async def autoend_on():
    return await add_on(config.AUTOEND)


async def autoend_off():
    return await add_off(config.AUTOEND)


# Auto leave assistant


async def is_autoleave():
    return await is_on_off(config.AUTOLEAVE)


async def autoleave_on():
    return await add_on(config.AUTOLEAVE)


async def autoleave_off():
    return await add_off(config.AUTOLEAVE)


# LOOP PLAY
async def get_loop(chat_id: int) -> int:
    return cache["loop"].get(chat_id, 0)


async def set_loop(chat_id: int, mode: int):
    cache["loop"][chat_id] = mode


# Channel Play IDS
async def get_cmode(chat_id: int) -> int | None:
    mode = cache["cmode"].get(chat_id)
    if mode is None:
        doc = await channeldb.find_one({"_id": chat_id})
        if not doc:
            return None
        mode = int(doc["mode"])
        cache["cmode"][chat_id] = mode
    return mode


async def set_cmode(chat_id: int, mode: int):
    mode = int(mode)
    if cache["cmode"].get(chat_id, 0) != mode:
        await channeldb.update_one(
            {"_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
        )
        cache["cmode"][chat_id] = mode


# Muted
async def is_muted(chat_id: int) -> bool:
    return int(chat_id) in cache["mute"]


async def mute_on(chat_id: int):
    cache["mute"].add(int(chat_id))


async def mute_off(chat_id: int):
    cache["mute"].discard(int(chat_id))


# Pause-Skip
async def is_music_paused(chat_id: int) -> bool:
    return chat_id in cache["pause"]


async def set_music_paused(chat_id: int):
    cache["pause"].add(chat_id)


async def set_music_playing(chat_id: int):
    cache["pause"].discard(chat_id)

# Active Voice Chats
async def get_active_chats() -> list:
    return list(cache["active_audio"])


async def is_active_chat(chat_id: int) -> bool:
    return chat_id in cache["active_audio"]


async def add_active_chat(chat_id: int):
    cache["active_audio"].add(chat_id)


async def remove_active_chat(chat_id: int):
    cache["active_audio"].discard(chat_id)
    
# Active Video Chats
async def get_active_video_chats() -> list:
    return list(cache["active_video"])


async def is_active_video_chat(chat_id: int) -> bool:
    return chat_id in cache["active_video"]


async def add_active_video_chat(chat_id: int):
    cache["active_video"].add(chat_id)


async def remove_active_video_chat(chat_id: int):
    cache["active_video"].discard(chat_id)

# PLAY TYPE WHETHER ADMINS ONLY OR EVERYONE
async def get_playtype(chat_id: int) -> str:
    mode = playtype.get(chat_id)
    if not mode:
        mode = await playtypedb.find_one({"chat_id": chat_id})
        if not mode:
            playtype[chat_id] = "Everyone"
            return "Everyone"
        playtype[chat_id] = mode["mode"]
        return mode["mode"]
    return mode


async def set_playtype(chat_id: int, mode: str):
    playtype[chat_id] = mode
    await playtypedb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# play mode whether inline or direct query
async def get_playmode(chat_id: int) -> str:
    mode = playmode.get(chat_id)
    if not mode:
        mode = await playmodedb.find_one({"chat_id": chat_id})
        if not mode:
            playmode[chat_id] = "Direct"
            return "Direct"
        playmode[chat_id] = mode["mode"]
        return mode["mode"]
    return mode


async def set_playmode(chat_id: int, mode: str):
    playmode[chat_id] = mode
    await playmodedb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# language
async def get_lang(chat_id: int) -> str:
    mode = langm.get(chat_id)
    if not mode:
        lang = await langdb.find_one({"chat_id": chat_id})
        if not lang:
            langm[chat_id] = "en"
            return "en"
        langm[chat_id] = lang["lang"]
        return lang["lang"]
    return mode


async def set_lang(chat_id: int, lang: str):
    langm[chat_id] = lang
    await langdb.update_one({"chat_id": chat_id}, {"$set": {"lang": lang}}, upsert=True)

# ---------- CLEANMODE ----------

async def is_cleanmode_on(chat_id: int) -> bool:
    if not cache["cleanmode"]:
        doc = await cleanmodedb.find_one({"_id": "cleanmode"}) or {}
        cache["cleanmode"] = set(doc.get("cleanmode", []))
    return chat_id in cache["cleanmode"]


async def cleanmode_on(chat_id: int):  # ENABLE
    if not await is_cleanmode_on(chat_id):
        cache["cleanmode"].add(chat_id)
        await cleanmodedb.update_one(
            {"_id": "cleanmode"},
            {"$addToSet": {"cleanmode": chat_id}},
            upsert=True
        )


async def cleanmode_off(chat_id: int):  # DISABLE
    if await is_cleanmode_on(chat_id):
        cache["cleanmode"].remove(chat_id)
        await cleanmodedb.update_one(
            {"_id": "cleanmode"},
            {"$pull": {"cleanmode": chat_id}},
            upsert=True
        )


# ---------- COMMAND DELETE ----------

async def is_commanddelete_on(chat_id: int) -> bool:
    if not cache["commanddelete"]:
        doc = await cleanmodedb.find_one({"_id": "cleanmode"}) or {}
        cache["commanddelete"] = set(doc.get("commanddelete", []))
    return chat_id in cache["commanddelete"]


async def commanddelete_on(chat_id: int):  # ENABLE
    if not await is_commanddelete_on(chat_id):
        cache["commanddelete"].add(chat_id)
        await cleanmodedb.update_one(
            {"_id": "cleanmode"},
            {"$addToSet": {"commanddelete": chat_id}},
            upsert=True
        )


async def commanddelete_off(chat_id: int):  # DISABLE
    if await is_commanddelete_on(chat_id):
        cache["commanddelete"].remove(chat_id)
        await cleanmodedb.update_one(
            {"_id": "cleanmode"},
            {"$pull": {"commanddelete": chat_id}},
            upsert=True
        )

# Non Admin Chat

async def is_nonadmin_chat(chat_id: int) -> bool:
  if chat_id in cache["nonadmin"]:
        return cache["nonadmin"][chat_id]

    user = await authdb.find_one({"_id": chat_id})
    exists = bool(user)
    cache["nonadmin"][chat_id] = exists
    return exists

async def add_nonadmin_chat(chat_id: int):
    cache["nonadmin"][chat_id] = True
    if await is_nonadmin_chat(chat_id):
        return
    return await authdb.insert_one({"_id": chat_id})


async def remove_nonadmin_chat(chat_id: int):
    cache["nonadmin"].pop(chat_id, None)
    if not await is_nonadmin_chat(chat_id):
        return
    return await authdb.delete_one({"_id": chat_id})

# Video Limit

async def is_video_allowed(chat_id: int) -> bool:
    limit = await get_video_limit()
    if not limit or limit == 0:
        return False

    count = len(await get_active_video_chats())
    if count >= limit and not await is_active_video_chat(chat_id):
        return False
    return True


async def get_video_limit() -> int | None:
    limit = cache.get("videolimit")
    if limit is None:
        dblimit = await videodb.find_one({"_id": "videolimit"})
        limit = int(dblimit["limit"]) if dblimit else None
        cache["videolimit"] = limit
    return limit


async def set_video_limit(limt: int):
    cache["videolimit"] = limt
    return await videodb.update_one(
        {"_id": "videolimit"}, {"$set": {"limit": limt}}, upsert=True
    )


# On Off


async def preload_onoff_cache():
    cursor = onoffdb.find({})
    cache["on_off"] = {str(doc["on_off"]).lower() async for doc in cursor}


async def is_on_off(on_off: int | str) -> bool:
    return str(on_off).lower() in cache["on_off"]


async def add_on(on_off: int | str):
    value = str(on_off).lower()
    if value not in cache["on_off"]:
        await onoffdb.insert_one({"on_off": value})
        cache["on_off"].add(value)


async def add_off(on_off: int | str):
    value = str(on_off).lower()
    if value in cache["on_off"]:
        await onoffdb.delete_one({"on_off": value})
        cache["on_off"].remove(value)


# Maintenance
async def is_maintenance():
    return await is_on_off(config.MAINTENANCE)


async def maintenance_on():
    return await add_on(config.MAINTENANCE)


async def maintenance_off():
    return await add_off(config.MAINTENANCE)

# --- Save Bitrate ---
async def save_audio_bitrate(chat_id: int, bitrate: str):
    cache["audio_bitrate"][chat_id] = bitrate


async def save_video_bitrate(chat_id: int, bitrate: str):
    cache["video_bitrate"][chat_id] = bitrate

# --- Get Bitrate Names (raw strings) ---
async def get_aud_bit_name(chat_id: int) -> str:
    return cache["audio_bitrate"].get(chat_id, "STUDIO")

async def get_vid_bit_name(chat_id: int) -> str:
    return cache["video_bitrate"].get(chat_id, "UHD_4K")

# --- Get Bitrate Enum Values ---
async def get_audio_bitrate(chat_id: int):
    mode = cache["audio_bitrate"].get(chat_id, "STUDIO")
    return {
        "STUDIO": _types.AudioQuality.STUDIO,
        "HIGH": _types.AudioQuality.HIGH,
        "MEDIUM": _types.AudioQuality.MEDIUM,
        "LOW": _types.AudioQuality.LOW,
    }.get(mode, _types.AudioQuality.STUDIO)


async def get_video_bitrate(chat_id: int):
    mode = cache["video_bitrate"].get(chat_id, "UHD_4K")
    return {
        "UHD_4K": _types.VideoQuality.UHD_4K,
        "QHD_2K": _types.VideoQuality.QHD_2K,
        "FHD_1080p": _types.VideoQuality.FHD_1080p,
        "HD_720p": _types.VideoQuality.HD_720p,
        "SD_480p": _types.VideoQuality.SD_480p,
        "SD_360p": _types.VideoQuality.SD_360p,
    }.get(mode, _types.VideoQuality.UHD_4K)

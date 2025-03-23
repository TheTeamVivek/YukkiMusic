#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from pytgcalls.types import AudioQuality, VideoQuality

import config
from YukkiMusic.core.mongo import mongodb

channeldb = mongodb.cplaymode
commanddb = mongodb.commands
cleandb = mongodb.cleanmode
playmodedb = mongodb.playmode
playtypedb = mongodb.playtypedb
langdb = mongodb.language
authdb = mongodb.adminauth
videodb = mongodb.yukkivideocalls
onoffdb = mongodb.onoffper
autoenddb = mongodb.autoend

# Shifting to memory [ mongo sucks often]
audio = {}
video = {}
loop = {}
playtype = {}
playmode = {}
channelconnect = {}
langm = {}
pause = {}
mute = {}
active = []
activevideo = []
cleanmode = []
command = []
nonadmin = {}
vlimit = []
maintenance = []
autoend = {}

# Auto End Stream

async def is_autoend() -> bool:
    chat_id = 123
    if chat_id not in autoend:
        autoend[chat_id] = bool(await autoenddb.find_one({"chat_id": chat_id}))
    return autoend[chat_id]

async def autoend_on():
    autoend[123] = True
    if not await autoenddb.find_one({"chat_id": 123}):
        return await autoenddb.insert_one({"chat_id": 123})

async def autoend_off():
    autoend[123] = False
    if await autoenddb.find_one({"chat_id": 123}):
        return await autoenddb.delete_one({"chat_id": 123})

# LOOP PLAY
async def get_loop(chat_id: int) -> int:
    return loop.get(chat_id, 0)

async def set_loop(chat_id: int, mode: int):
    loop[chat_id] = mode


# Channel Play IDS
async def get_cmode(chat_id: int) -> int | None:
    if chat_id not in channelconnect:
        mode = await channeldb.find_one({"chat_id": chat_id})
        channelconnect[chat_id] = mode["mode"] if mode else None
    return channelconnect[chat_id]


async def set_cmode(chat_id: int, mode: int):
    channelconnect[chat_id] = mode
    await channeldb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# PLAY TYPE WHETHER ADMINS ONLY OR EVERYONE
async def get_playtype(chat_id: int) -> str:
    if chat_id not in playtype:
        playtype[chat_id] = (await playtypedb.find_one({"chat_id": chat_id}) or {"mode": "Everyone"})["mode"]
    return playtype[chat_id]


async def set_playtype(chat_id: int, mode: str):
    playtype[chat_id] = mode
    await playtypedb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# play mode whether inline or direct query
async def get_playmode(chat_id: int) -> str:
    if chat_id not in playmode:
        playmode[chat_id] = (await playmodedb.find_one({"chat_id": chat_id}) or {"mode": "Direct"})["mode"]
    return playmode[chat_id]


async def set_playmode(chat_id: int, mode: str):
    playmode[chat_id] = mode
    await playmodedb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# language
async def get_lang(chat_id: int) -> str:
    if chat_id not in langm:
        langm[chat_id] = (await langdb.find_one({"chat_id": chat_id}) or {"lang": "en"})["lang"]
    return langm[chat_id]


async def set_lang(chat_id: int, lang: str):
    langm[chat_id] = lang
    await langdb.update_one({"chat_id": chat_id}, {"$set": {"lang": lang}}, upsert=True)


# Muted
async def is_muted(chat_id: int) -> bool:
    return mute.get(chat_id, False)


async def mute_on(chat_id: int):
    mute[chat_id] = True


async def mute_off(chat_id: int):
    mute[chat_id] = False


# Pause-Skip
async def is_music_playing(chat_id: int) -> bool:
    return pause.get(chat_id, False)


async def music_on(chat_id: int):
    pause[chat_id] = True


async def music_off(chat_id: int):
    pause[chat_id] = False


# Active Voice Chats
async def get_active_chats() -> list:
    return active


async def is_active_chat(chat_id: int) -> bool:
    return chat_id in active


async def add_active_chat(chat_id: int):
    if chat_id not in active:
        active.append(chat_id)


async def remove_active_chat(chat_id: int):
    if chat_id in active:
        active.remove(chat_id)


# Active Video Chats
async def get_active_video_chats() -> list:
    return activevideo


async def is_active_video_chat(chat_id: int) -> bool:
    return chat_id in activevideo
    

async def add_active_video_chat(chat_id: int):
    if chat_id not in activevideo:
        activevideo.append(chat_id)


async def remove_active_video_chat(chat_id: int):
    if chat_id in activevideo:
        activevideo.remove(chat_id)


# Delete command mode


async def is_cleanmode_on(chat_id: int) -> bool:
    return chat_id not in cleanmode


async def cleanmode_off(chat_id: int):
    if chat_id in cleanmode:
        cleanmode.append(chat_id)


async def cleanmode_on(chat_id: int):
    if chat_id in cleanmode:
        cleanmode.remove(chat_id)


async def is_commanddelete_on(chat_id: int) -> bool:
    return chat_id not in command


async def commanddelete_off(chat_id: int):
    if chat_id in cleanmode:
        command.append(chat_id)


async def commanddelete_on(chat_id: int):
    if chat_id in command:
        command.remove(chat_id)


# Non Admin Chat
async def check_nonadmin_chat(chat_id: int) -> bool:
    user = await authdb.find_one({"chat_id": chat_id})
    if user:
        return True
    return False


async def is_nonadmin_chat(chat_id: int) -> bool:
    if chat_id in nonadmin:
        return nonadmin[chat_id]
    nonadmin[chat_id] = bool(await authdb.find_one({"chat_id": chat_id}))
    return nonadmin[chat_id]


async def add_nonadmin_chat(chat_id: int):
    nonadmin[chat_id] = True
    is_admin = await check_nonadmin_chat(chat_id)
    if not is_admin:
        return await authdb.insert_one({"chat_id": chat_id})


async def remove_nonadmin_chat(chat_id: int):
    nonadmin[chat_id] = False
    is_admin = await check_nonadmin_chat(chat_id)
    if is_admin:
        return await authdb.delete_one({"chat_id": chat_id})


# Video Limit
async def is_video_allowed(chat_id: int) -> bool:
    if not vlimit:
        dblimit = await videodb.find_one({"chat_id": 123456})
        limit = dblimit["limit"] if dblimit else config.VIDEO_STREAM_LIMIT
        vlimit[:] = [limit]
    else:
        limit = vlimit[0]

    return limit != 0 and (await is_active_video_chat(chat_id) or len(await get_active_video_chats()) < limit)
    
async def get_video_limit() -> str:
    if not vlimit:
        dblimit = await videodb.find_one({"chat_id": 123456})
        vlimit.append(dblimit["limit"] if dblimit else config.VIDEO_STREAM_LIMIT)
    return vlimit[0]


async def set_video_limit(limit: int):
    vlimit[:] = [limit]
    return await videodb.update_one(
        {"chat_id": 123456}, {"$set": {"limit": limit}}, upsert=True
    )


# On Off
async def is_on_off(on_off: int) -> bool:
    return bool(await onoffdb.find_one({"on_off": on_off}))
    

async def add_on(on_off: int):
    is_on = await is_on_off(on_off)
    if not is_on:
        return await onoffdb.insert_one({"on_off": on_off})


async def add_off(on_off: int):
    is_off = await is_on_off(on_off)
    if is_off:
        return await onoffdb.delete_one({"on_off": on_off})


# Maintenance


async def is_maintenance() -> bool:
    if not maintenance:
        maintenance[:] = [1 if await onoffdb.find_one({"on_off": 1}) else 2]
    return 1 not in maintenance

async def maintenance_off():
    maintenance[:] = [2]
    if await is_on_off(1):
        return await onoffdb.delete_one({"on_off": 1})


async def maintenance_on():
    maintenance[:] = [1]
    if not await is_on_off(1):
        return await onoffdb.insert_one({"on_off": 1})

async def save_audio_bitrate(chat_id: int, bitrate: str):
    audio[chat_id] = bitrate


async def save_video_bitrate(chat_id: int, bitrate: str):
    video[chat_id] = bitrate


async def get_aud_bit_name(chat_id: int) -> str:
    return audio.get(chat_id, "HIGH")


async def get_vid_bit_name(chat_id: int) -> str:
    return video.get(chat_id, "HD_720p")


async def get_audio_bitrate(chat_id: int) -> str:
    mode = audio.get(chat_id, "MEDIUM")
    return {
        "STUDIO": AudioQuality.STUDIO,
        "HIGH": AudioQuality.HIGH,
        "MEDIUM": AudioQuality.MEDIUM,
        "LOW": AudioQuality.LOW,
    }.get(mode, AudioQuality.MEDIUM)


async def get_video_bitrate(chat_id: int) -> str:
    mode = video.get(chat_id, "SD_480p")
    return {
        "UHD_4K": VideoQuality.UHD_4K,
        "QHD_2K": VideoQuality.QHD_2K,
        "FHD_1080p": VideoQuality.FHD_1080p,
        "HD_720p": VideoQuality.HD_720p,
        "SD_480p": VideoQuality.SD_480p,
        "SD_360p": VideoQuality.SD_360p,
    }.get(mode, VideoQuality.SD_480p)

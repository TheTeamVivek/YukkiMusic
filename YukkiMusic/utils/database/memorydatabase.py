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

import config as _config
from YukkiMusic.core.mongo import mongodb as _mongodb

_channeldb = _mongodb.cplaymode
_commanddb = _mongodb.commands  # TODO NOT IN JSED FIND IT
_cleandb = _mongodb.cleanmode  # TDOD NOT IN USED
_playmodedb = _mongodb.playmode
_playtypedb = _mongodb.playtypedb
_langdb = _mongodb.language
_authdb = _mongodb.adminauth
_videodb = _mongodb.videocalls
_onoffdb = _mongodb.onoffper
_autoenddb = _mongodb.autoend

# Shifting to memory [ mongo sucks often]
_audio = {}
_video = {}
_loop = {}
_playtype = {}
_playmode = {}
_channelconnect = {}
_langm = {}
_pause = {}
_mute = {}
_active = []
_activevideo = []
_cleanmode = []
_command = []
_nonadmin = {}
_vlimit = []
_maintenance = []
_autoend = {}

# Auto End Stream


async def is_autoend() -> bool:
    chat_id = 123
    if chat_id not in _autoend:
        _autoend[chat_id] = bool(await _autoenddb.find_one({"chat_id": chat_id}))
    return _autoend[chat_id]


async def autoend_on():
    _autoend[123] = True
    if not await _autoenddb.find_one({"chat_id": 123}):
        return await _autoenddb.insert_one({"chat_id": 123})


async def autoend_off():
    _autoend[123] = False
    if await _autoenddb.find_one({"chat_id": 123}):
        return await _autoenddb.delete_one({"chat_id": 123})


# LOOP PLAY
async def get_loop(chat_id: int) -> int:
    return _loop.get(chat_id, 0)


async def set_loop(chat_id: int, mode: int):
    _loop[chat_id] = mode


# Channel Play IDS
async def get_cmode(chat_id: int) -> int | None:
    if chat_id not in _channelconnect:
        mode = await _channeldb.find_one({"chat_id": chat_id})
        _channelconnect[chat_id] = mode["mode"] if mode else None
    return _channelconnect[chat_id]


async def set_cmode(chat_id: int, mode: int):
    _channelconnect[chat_id] = mode
    await _channeldb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# PLAY TYPE WHETHER ADMINS ONLY OR EVERYONE
async def get_playtype(chat_id: int) -> str:
    if chat_id not in _playtype:
        _playtype[chat_id] = (
            await _playtypedb.find_one({"chat_id": chat_id}) or {"mode": "EVERYONE"}
        )["mode"]
    return _playtype[chat_id]


async def set_playtype(chat_id: int, mode: str):
    _playtype[chat_id] = mode
    await _playtypedb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# play mode whether inline or direct query
async def get_playmode(chat_id: int) -> str:
    if chat_id not in _playmode:
        _playmode[chat_id] = (
            await _playmodedb.find_one({"chat_id": chat_id}) or {"mode": "DIRECT"}
        )["mode"]
    return _playmode[chat_id]


async def set_playmode(chat_id: int, mode: str):
    _playmode[chat_id] = mode
    await _playmodedb.update_one(
        {"chat_id": chat_id}, {"$set": {"mode": mode}}, upsert=True
    )


# language
async def get_lang(chat_id: int) -> str:
    if chat_id not in _langm:
        _langm[chat_id] = (
            await _langdb.find_one({"chat_id": chat_id}) or {"lang": "en"}
        )["lang"]
    return _langm[chat_id]


async def set_lang(chat_id: int, lang: str):
    _langm[chat_id] = lang
    await _langdb.update_one(
        {"chat_id": chat_id}, {"$set": {"lang": lang}}, upsert=True
    )


# Muted
async def is_muted(chat_id: int) -> bool:
    return _mute.get(chat_id, False)


async def mute_on(chat_id: int):
    _mute[chat_id] = True


async def mute_off(chat_id: int):
    _mute[chat_id] = False


# Pause-Skip
async def is_music_playing(chat_id: int) -> bool:
    return _pause.get(chat_id, False)


def is_music_playing_sync(chat_id: int) -> bool:
    return _pause.get(chat_id, False)


async def music_on(chat_id: int):
    _pause[chat_id] = True


async def music_off(chat_id: int):
    _pause[chat_id] = False


# Active Voice Chats
async def get_active_chats() -> list:
    return _active


async def is_active_chat(chat_id: int) -> bool:
    return chat_id in _active


async def add_active_chat(chat_id: int):
    if chat_id not in _active:
        _active.append(chat_id)


async def remove_active_chat(chat_id: int):
    if chat_id in _active:
        _active.remove(chat_id)


# Active Video Chats
async def get_active_video_chats() -> list:
    return _activevideo


async def is_active_video_chat(chat_id: int) -> bool:
    return chat_id in _activevideo


async def add_active_video_chat(chat_id: int):
    if chat_id not in _activevideo:
        _activevideo.append(chat_id)


async def remove_active_video_chat(chat_id: int):
    if chat_id in _activevideo:
        _activevideo.remove(chat_id)


# Delete command mode


async def is_cleanmode_on(chat_id: int) -> bool:
    return chat_id not in _cleanmode


async def cleanmode_off(chat_id: int):
    if chat_id not in _cleanmode:
        _cleanmode.append(chat_id)


async def cleanmode_on(chat_id: int):
    if chat_id in _cleanmode:
        _cleanmode.remove(chat_id)


async def is_commanddelete_on(chat_id: int) -> bool:
    return chat_id not in _command


async def commanddelete_off(chat_id: int):
    if chat_id in _cleanmode:
        _command.append(chat_id)


async def commanddelete_on(chat_id: int):
    if chat_id in _command:
        _command.remove(chat_id)


# Non Admin Chat
async def check_nonadmin_chat(chat_id: int) -> bool:
    user = await _authdb.find_one({"chat_id": chat_id})
    if user:
        return True
    return False


async def is_nonadmin_chat(chat_id: int) -> bool:
    if chat_id in _nonadmin:
        return _nonadmin[chat_id]
    _nonadmin[chat_id] = bool(await _authdb.find_one({"chat_id": chat_id}))
    return _nonadmin[chat_id]


async def add_nonadmin_chat(chat_id: int):
    _nonadmin[chat_id] = True
    is_admin = await check_nonadmin_chat(chat_id)
    if not is_admin:
        return await _authdb.insert_one({"chat_id": chat_id})


async def remove_nonadmin_chat(chat_id: int):
    _nonadmin[chat_id] = False
    is_admin = await check_nonadmin_chat(chat_id)
    if is_admin:
        return await _authdb.delete_one({"chat_id": chat_id})


# Video Limit
async def is_video_allowed(chat_id: int) -> bool:
    if not _vlimit:
        dblimit = await _videodb.find_one({"chat_id": 123456})
        limit = dblimit["limit"] if dblimit else _config.VIDEO_STREAM_LIMIT
        _vlimit[:] = [limit]
    else:
        limit = _vlimit[0]

    return limit != 0 and (
        await is_active_video_chat(chat_id)
        or len(await get_active_video_chats()) < limit
    )


async def get_video_limit() -> str:
    if not _vlimit:
        dblimit = await _videodb.find_one({"chat_id": 123456})
        _vlimit.append(dblimit["limit"] if dblimit else _config.VIDEO_STREAM_LIMIT)
    return _vlimit[0]


async def set_video_limit(limit: int):
    _vlimit[:] = [limit]
    return await _videodb.update_one(
        {"chat_id": 123456}, {"$set": {"limit": limit}}, upsert=True
    )


# On Off
async def is_on_off(on_off: int) -> bool:
    return bool(await _onoffdb.find_one({"on_off": on_off}))


async def add_on(on_off: int):
    is_on = await is_on_off(on_off)
    if not is_on:
        return await _onoffdb.insert_one({"on_off": on_off})


async def add_off(on_off: int):
    is_off = await is_on_off(on_off)
    if is_off:
        return await _onoffdb.delete_one({"on_off": on_off})


# Maintenance


async def is_maintenance() -> bool:
    if not _maintenance:
        _maintenance[:] = [1 if await _onoffdb.find_one({"on_off": 1}) else 2]
    return 1 not in _maintenance


async def maintenance_off():
    _maintenance[:] = [2]
    if await is_on_off(1):
        return await _onoffdb.delete_one({"on_off": 1})


async def maintenance_on():
    _maintenance[:] = [1]
    if not await is_on_off(1):
        return await _onoffdb.insert_one({"on_off": 1})


async def save_audio_bitrate(chat_id: int, bitrate: str):
    _audio[chat_id] = bitrate


async def save_video_bitrate(chat_id: int, bitrate: str):
    _video[chat_id] = bitrate


async def get_aud_bit_name(chat_id: int) -> str:
    return _audio.get(chat_id, "STUDIO")


async def get_vid_bit_name(chat_id: int) -> str:
    return _video.get(chat_id, "UHD_4K")


async def get_audio_bitrate(chat_id: int) -> str:
    mode = _audio.get(chat_id, "STUDIO")
    return {
        "STUDIO": _types.AudioQuality.STUDIO,
        "HIGH": _types.AudioQuality.HIGH,
        "MEDIUM": _types.AudioQuality.MEDIUM,
        "LOW": _types.AudioQuality.LOW,
    }.get(mode, _types.AudioQuality.STUDIO)


async def get_video_bitrate(chat_id: int) -> str:
    mode = _video.get(chat_id, "UHD_4K")
    return {
        "UHD_4K": _types.VideoQuality.UHD_4K,
        "QHD_2K": _types.VideoQuality.QHD_2K,
        "FHD_1080p": _types.VideoQuality.FHD_1080p,
        "HD_720p": _types.VideoQuality.HD_720p,
        "SD_480p": _types.VideoQuality.SD_480p,
        "SD_360p": _types.VideoQuality.SD_360p,
    }.get(mode, _types.VideoQuality.UHD_4K)

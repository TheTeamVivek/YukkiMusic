#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from typing import Union

from telethon.tl import types

from .channelplay import get_channeplay_cb, is_cplay
from .database import *
from .decorators import *
from .exceptions import AssistantErr
from .formatters import *
from .inline import *
from .logger import play_logs
from .pastebin import *
from .pastebin import paste
from .stream import *
from .sys import *
from .sys import bot_sys_stats
from .thumbnails import gen_qthumb, gen_thumb


async def get_value(chat_id, key) -> list[str]:
    from strings import get_string

    value = []

    language = await get_lang(chat_id)  # get_lang from .database
    if language != "en":
        value.append(get_string(language)[key])
        value.append(get_string("en")[key])
    else:
        value.append(get_string(language)[key])
    return value


async def get_message_link(msg):
    chat = await msg.get_chat()
    if username := chat.username:
        link = f"https://t.me/{username}/{msg.id}"
    else:
        link = f"https://t.me/c/{chat.id}/{msg.id}"
    return link


def get_chat_id(entity: types.User | types.Chat | types.Channel) -> int:
    chat_id = None
    if isinstance(entity, types.User):
        chat_id = entity.id
    elif isinstance(entity, types.Chat):
        chat_id = int(f"-{entity.id}")
    elif isinstance(entity, types.Channel):
        chat_id = int(f"-100{entity.id}")
    return chat_id


async def parse_flags(chat_id, text: str):
    vplay_flag = await get_value(chat_id, "VPLAY_FLAGS")
    cplay_flag = await get_value(chat_id, "CPLAY_FLAGS")
    playforce_flag = await get_value(chat_id, "FPLAY_FLAGS")

    is_vplay = is_forceplay = is_cplay = False
    args = text.lstrip("/").lower().split()
    comm = args[0]

    for arg in args:
        if arg in [f"v{comm}", *vplay_flag]:  # /play -v or /vplay
            is_vplay = True

        elif arg in [
            f"{comm}",
            f"{comm}force",
            *playforce_flag,
        ]:  # /fplay or /playforce or /play -f
            is_forceplay = True

        elif arg == f"v{comm}force":  # /vplayforce the oldstyle
            is_forceplay, is_vplay = True, True

        elif arg in [f"c{comm}", *cplay_flag]:  # /cplay
            is_cplay = True

        elif arg == f"cv{comm}":  # /cvplay
            is_vplay, is_cplay = True, True

        elif arg == f"cv{comm}force":  # /cvplayforce
            is_vplay, is_cplay, is_forceplay = True, True, True

        elif arg == f"c{comm}force":  # /cplayforce
            is_cplay, is_forceplay = True, True

    return is_vplay, is_forceplay, is_cplay

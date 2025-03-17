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


def get_chat_id(entity: types.User | types.Chat | types.Channel) -> int:
    chat_id = None
    if isinstance(entity, types.User):
        chat_id = entity.id
    elif isinstance(entity, types.Chat):
        chat_id = int(f"-{entity.id}")
    elif isinstance(entity, types.Channel):
        chat_id = int(f"-100{entity.id}")
    return chat_id

def parse_flags(text: str):
    is_vplay = is_forceplay = is_cplay = False
    args = text.lstrip("/").split()

    for arg in args:
        if arg in ["-v", "vplay"]:
            is_vplay = True
        elif arg in ["-f", "fplay", "forceplay"]:
            is_forceplay = True
        elif arg in ["-c", "cplay"]:
            is_cplay = True
        elif arg == "cvplay":
            is_vplay, is_cplay = True, True
        elif arg == "cvplayforce":
            is_vplay, is_cplay, is_forceplay = True, True, True

    return is_vplay, is_forceplay, is_cplay
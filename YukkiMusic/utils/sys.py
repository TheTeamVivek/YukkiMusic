#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

import time

import psutil

from YukkiMusic.misc import _boot_

from .formatters import get_readable_time

__all__ = ["bot_sys_stats"]


async def bot_sys_stats():
    bot_uptime = int(time.time() - _boot_)
    read_able_uptime = f"{get_readable_time(bot_uptime)}"
    cpu = f"{psutil.cpu_percent(interval=0.5)}%"
    ram = f"{psutil.virtual_memory().percent}%"
    disk = f"{psutil.disk_usage('/').percent}%"
    return read_able_uptime, cpu, ram, disk

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

from yukkimusic.misc import _boot_

from .decorators.asyncify import asyncify
from .formatters import get_readable_time


@asyncify
def bot_sys_stats():
    uptime = get_readable_time(int(time.time() - _boot_))
    cpu = f"{psutil.cpu_percent(interval=0.5)}%"
    ram = f"{psutil.virtual_memory().percent}%"
    disk = f"{psutil.disk_usage('/').percent}%"
    return uptime, cpu, ram, disk

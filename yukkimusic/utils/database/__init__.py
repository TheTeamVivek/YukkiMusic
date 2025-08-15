#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from .assistantdatabase import *
from .memorydatabase import *
from .memorydatabase import preload_onoff_cache
from .mongodatabase import *


async def init():
    await preload_onoff_cache()
    await mongodb.autoend.drop()

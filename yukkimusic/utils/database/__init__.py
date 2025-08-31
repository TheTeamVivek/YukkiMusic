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
from .mongodatabase import *


async def init():
    await migrate_served_stats()  # from mongodatabase
    await migrate_blocklist()  # from mongodatabase
    await migrate_private_chats()  # from mongodatabase
    await mongodb.autoend.drop()

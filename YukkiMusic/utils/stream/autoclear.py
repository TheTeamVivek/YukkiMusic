#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import os
from config import autoclean
from YukkiMusic.utils.decorators import asyncify


@asyncify
def auto_clean(popped):
    def _auto_clean(popped_item):
        try:
            rem = popped_item["file"]
            autoclean.remove(rem)
            count = autoclean.count(rem)
            if count == 0:
                if "vid_" not in rem and "live_" not in rem and "index_" not in rem:
                    try:
                        os.remove(rem)
                    except Exception:
                        pass
        except Exception:
            pass

    if isinstance(popped, dict):
        _auto_clean(popped)
    elif isinstance(popped, list):
        for pop in popped:
            _auto_clean(pop)
    else:
        raise ValueError("Expected popped to be a dict or list.")

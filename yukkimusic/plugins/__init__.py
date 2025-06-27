#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
#
# All rights reserved.
#

import importlib
import os

from yukkimusic import HELPABLE

base_dir = os.path.dirname(__file__)

for root, dirs, files in os.walk(base_dir):
    for file in files:
        if file.endswith(".py") and file != "__init__.py":
            full_path = os.path.join(root, file)
            rel_path = os.path.relpath(full_path, base_dir)
            mod_name = rel_path.replace(os.sep, ".")[:-3]

            mod = importlib.import_module(
                f"{__package__}.{mod_name}" if __package__ else mod_name
            )

            if (
                mod
                and hasattr(mod, "__MODULE__")
                and mod.__MODULE__
                and (hasattr(mod, "__HELP__") and mod.__HELP__)
            ):
                HELPABLE[mod.__MODULE__.lower()] = mod

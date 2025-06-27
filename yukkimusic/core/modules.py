#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
# pylint: disable=missing-module-docstring, missing-class-docstring, missing-function-docstring
from __future__ import annotations  # noqa

import asyncio
import importlib
import logging
from dataclasses import dataclass
from logging.handlers import RotatingFileHandler
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pymongo.asynchronous.database import AsyncDatabase

    from .bot import YukkiBot
    from .help import ModuleHelp
    from .userbot import Userbot

logger = logging.getLogger(__name__)


@dataclass
class LoaderContext:
    app: YukkiBot
    userbot: Userbot
    mongodb: AsyncDatabase
    help: ModuleHelp


def _setup_logger(name: str):
    plugin_logger = logging.getLogger(name)
    plugin_logger.setLevel(logging.INFO)

    if plugin_logger.handlers:
        return

    formatter = logging.Formatter(
        "{asctime} - {levelname} - {message}", style="{", datefmt="%d-%b-%y %H:%M:%S"
    )

    file_handler = RotatingFileHandler("logs.txt", maxBytes=5_000_000, backupCount=10)
    file_handler.setFormatter(formatter)

    stream_handler = logging.StreamHandler()
    stream_handler.setFormatter(formatter)

    plugin_logger.addHandler(file_handler)
    plugin_logger.addHandler(stream_handler)
    plugin_logger.propagate = False


async def load_mod(modules: list[str], ctx: LoaderContext):
    for mod_name in modules:
        try:
            mod = importlib.import_module(mod_name)
        except ImportError as e:
            logger.warning("[MOD] Failed to import '%s': %s", mod_name, e)
            continue

        _setup_logger(mod_name)

        if hasattr(mod, "setup"):
            try:
                if asyncio.iscoroutinefunction(mod.setup):
                    await mod.setup(ctx)
                else:
                    mod.setup(ctx)
                logger.info("[MOD] Sucessfully Loaded: %s", mod_name)
            except Exception as e:  # pylint: disable=broad-exception-caught
                logger.warning("[MOD] Setup failed for '%s': %s", mod_name, e)
        else:
            logger.warning("[MOD] '%s' has no 'setup' method, Skipping...", mod_name)

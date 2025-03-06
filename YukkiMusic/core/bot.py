#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio
import importlib.util
import os
import sys
import traceback
from datetime import datetime
from functools import wraps

import uvloop
from pyrogram import Client, StopPropagation, errors
from pyrogram.enums import ChatMemberStatus
from pyrogram.errors import (
    ChatSendMediaForbidden,
    ChatSendPhotosForbidden,
    ChatWriteForbidden,
    FloodWait,
    MessageIdInvalid,
    MessageNotModified,
)
from pyrogram.handlers import MessageHandler

import config
from YukkiMusic.utils.decorators.asyncify import asyncify

from ..logging import logger
from . import filters as telethon_filters

uvloop.install()


class YukkiBot(Client):
    def __init__(self, *args, **kwargs):
        logger(__name__).info("Starting Bot...")

        super().__init__(*args, **kwargs)
        self.loaded_plug_counts: int = 0
        self.name: str = None
        self.username: str = None
        self.mention: str = None
        self.id: int = None

    def on_message(self, filters=None, group=0):
        def decorator(func):
            @wraps(func)
            async def wrapper(client, message):
                try:
                    await func(client, message)
                except FloodWait as e:
                    logger(__name__).warning(
                        "FloodWait: Sleeping for %d seconds.", e.value
                    )
                    await asyncio.sleep(e.value)
                except (
                    ChatWriteForbidden,
                    ChatSendMediaForbidden,
                    ChatSendPhotosForbidden,
                    MessageNotModified,
                    MessageIdInvalid,
                ):
                    pass
                except StopPropagation:
                    raise
                except Exception as e:
                    date_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                    user_id = event.sender_id if message.from_user else "Unknown"
                    chat_id = event.chat_id if message.chat else "Unknown"
                    chat_username = (
                        f"@{message.chat.username}"
                        if message.chat.username
                        else "Private Group"
                    )
                    command = (
                        " ".join(message.command)
                        if hasattr(message, "command")
                        else message.text
                    )
                    error_trace = traceback.format_exc()
                    error_message = (
                        f"**Error:** {type(e).__name__}\n"
                        f"**Date:** {date_time}\n"
                        f"**Chat ID:** {chat_id}\n"
                        f"**Chat Username:** {chat_username}\n"
                        f"**User ID:** {user_id}\n"
                        f"**Command/Text:** {command}\n"
                        f"**Traceback:**\n{error_trace}"
                    )
                    await self.send_message(config.LOG_GROUP_ID, error_message)
                    try:
                        await self.send_message(config.OWNER_ID[0], error_message)
                    except Exception:
                        pass

            handler = MessageHandler(wrapper, filters)
            self.add_handler(handler, group)
            return func

        return decorator

    @asyncify
    def load_plugin(self, file_path: str, base_dir: str, attrs: dict):
        relative_path = os.path.relpath(file_path, base_dir).replace(os.sep, ".")
        module_path = f"{os.path.basename(base_dir)}.{relative_path[:-3]}"

        spec = importlib.util.spec_from_file_location(module_path, file_path)
        module = importlib.util.module_from_spec(spec)
        module.logger = logger(module_path)
        module.app = self
        module.Config = config
        mod.flt = telethon_filters
        for name, attr in attrs.items():
            setattr(module, name, attr)

        try:
            spec.loader.exec_module(module)
            self.loaded_plug_counts += 1
        except Exception as e:
            logger(__name__).error(
                "Failed to load %s: %s\n\n", module_path, e, exc_info=True
            )
            sys.exit()

        return module

    async def load_plugins_from(self, base_folder: str, attrs: dict):
        base_dir = os.path.abspath(base_folder)
        utils_path = os.path.join(base_dir, "utils.py")
        utils = None

        if os.path.exists(utils_path) and os.path.isfile(utils_path):
            try:
                spec = importlib.util.spec_from_file_location("utils", utils_path)
                utils = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(utils)
            except Exception as e:
                logger(__name__).error(
                    "Failed to load 'utils' module: %s", e, exc_info=True
                )

        if utils:
            attrs["utils"] = utils
        for root, _, files in os.walk(base_dir):
            for file in files:
                if (
                    file.endswith(".py")
                    and not file == "utils.py"
                    or not file.startswith("__")
                ):
                    file_path = os.path.join(root, file)
                    mod = await self.load_plugin(file_path, base_dir, attrs)
                    yield mod

    async def run_shell_command(self, command: list):
        process = await asyncio.create_subprocess_exec(
            *command,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )

        stdout, stderr = await process.communicate()

        return {
            "returncode": process.returncode,
            "stdout": stdout.decode().strip() if stdout else None,
            "stderr": stderr.decode().strip() if stderr else None,
        }

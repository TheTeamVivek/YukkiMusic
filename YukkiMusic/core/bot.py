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
import traceback
from datetime import datetime
from functools import wraps

from pyrogram import Client, StopPropagation, errors, types
from pyrogram.enums import ChatMemberStatus
from pyrogram.handlers import MessageHandler

import config

from ..logging import LOGGER


class YukkiBot(Client):
    def __init__(self, *args, **kwargs):
        LOGGER(__name__).info("Starting Bot...")

        super().__init__(
            "YukkiMusic",
            api_id=config.API_ID,
            api_hash=config.API_HASH,
            bot_token=config.BOT_TOKEN,
            workers=50,
            sleep_threshold=240,
            max_concurrent_transmissions=5,
            link_preview_options=types.LinkPreviewOptions(is_disabled=True),
        )
        self.loaded_plug_counts = 0

    def on_message(self, filters=None, group=0):
        def decorator(func):
            @wraps(func)
            async def wrapper(client, message):
                try:
                    if asyncio.iscoroutinefunction(func):
                        await func(client, message)
                    else:
                        func(client, message)
                except errors.FloodWait as e:
                    LOGGER(__name__).warning(
                        f"FloodWait: Sleeping for {e.value} seconds."
                    )
                    await asyncio.sleep(e.value)
                except (
                    errors.ChatWriteForbidden,
                    errors.ChatSendMediaForbidden,
                    errors.ChatSendPhotosForbidden,
                    errors.MessageNotModified,
                    errors.MessageIdInvalid,
                ):
                    pass
                except StopPropagation:
                    raise
                except Exception as e:
                    date_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                    user_id = message.from_user.id if message.from_user else "Unknown"
                    chat_id = message.chat.id if message.chat else "Unknown"
                    chat_username = (
                        f"@{message.chat.username}"
                        if message.chat.username
                        else "Private Group"
                    )
                    command = message.text
                    error_trace = traceback.format_exc()
                    error_message = (
                        f"<b>Error:</b> {type(e).__name__}\n"
                        f"<b>Date:</b> {date_time}\n"
                        f"<b>Chat ID:</b> {chat_id}\n"
                        f"<b>Chat Username:</b> {chat_username}\n"
                        f"<b>User ID:</b> {user_id}\n"
                        f"<b>Command/Text:</b>\n<pre language='python'><code>{command}</code></pre>\n\n"
                        f"<b>Traceback:</b>\n<pre language='python'><code>{error_trace}</code></pre>"
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

    async def start(self):
        await super().start()
        get_me = await self.get_me()
        self.username = get_me.username
        self.id = get_me.id
        self.name = get_me.full_name
        self.mention = get_me.mention

        try:
            await self.send_message(
                config.LOG_GROUP_ID,
                text=(
                    f"<u><b>{self.mention} Bot Started :</b></u>\n\n"
                    f"Id : <code>{self.id}</code>\n"
                    f"Name : {self.name}\n"
                    f"Username : @{self.username}"
                ),
            )
        except (errors.ChannelInvalid, errors.PeerIdInvalid):
            LOGGER(__name__).error(
                "Bot failed to access the log group. Ensure the bot is added and promoted as admin."
            )
            LOGGER(__name__).error("Error details:", exc_info=True)
            exit()
        if config.SET_CMDS:
            try:
                await self._set_default_commands()
            except Exception:
                LOGGER(__name__).warning("Failed to set commands:", exc_info=True)

        try:
            a = await self.get_chat_member(config.LOG_GROUP_ID, "me")
            if a.status != ChatMemberStatus.ADMINISTRATOR:
                LOGGER(__name__).error("Please promote bot as admin in logger group")
                exit()
        except Exception:
            pass
        LOGGER(__name__).info(f"MusicBot started as {self.name}")

    async def _set_default_commands(self):
        private_commands = [
            types.BotCommand("start", "Start the bot"),
            types.BotCommand("help", "Get the help menu"),
            types.BotCommand("ping", "Check if the bot is alive or dead"),
        ]
        group_commands = [types.BotCommand("play", "Start playing requested song")]
        admin_commands = [
            types.BotCommand("play", "Start playing requested song"),
            types.BotCommand("skip", "Move to next track in queue"),
            types.BotCommand("pause", "Pause the current playing song"),
            types.BotCommand("resume", "Resume the paused song"),
            types.BotCommand("end", "Clear the queue and leave voice chat"),
            types.BotCommand("shuffle", "Randomly shuffle the queued playlist"),
            types.BotCommand("reboot", "Reboot the bot for your chat"),
            types.BotCommand("playmode", "Change the default playmode for your chat"),
            types.BotCommand("settings", "Open bot settings for your chat"),
        ]
        owner_commands = [
            types.BotCommand("update", "Update the bot"),
            types.BotCommand("logs", "Get logs"),
            # types.BotCommand("export", "Export all data of mongodb"),
            # types.BotCommand("import", "Import all data in mongodb"),
            types.BotCommand("addsudo", "Add a user as a sudoer"),
            types.BotCommand("delsudo", "Remove a user from sudoers"),
            types.BotCommand("sudolist", "List all sudo users"),
            types.BotCommand("getvar", "Get a specific environment variable"),
            types.BotCommand("delvar", "Delete a specific environment variable"),
            types.BotCommand("setvar", "Set a specific environment variable"),
            types.BotCommand("usage", "Get dyno usage information"),
            types.BotCommand("maintenance", "Enable or disable maintenance mode"),
            types.BotCommand("logger", "Enable or disable logging"),
            types.BotCommand("block", "Block a user"),
            types.BotCommand("unblock", "Unblock a user"),
            types.BotCommand("blacklist", "Blacklist a chat"),
            types.BotCommand("whitelist", "Whitelist a chat"),
            types.BotCommand("blacklisted", "List all blacklisted chats"),
            types.BotCommand("autoend", "Enable or disable auto end for streams"),
            types.BotCommand("restart", "Restart the bot"),
        ]

        await self.set_bot_commands(
            private_commands, scope=types.BotCommandScopeAllPrivateChats()
        )
        await self.set_bot_commands(
            group_commands, scope=types.BotCommandScopeAllGroupChats()
        )
        await self.set_bot_commands(
            admin_commands, scope=types.BotCommandScopeAllChatAdministrators()
        )

        LOG_GROUP_ID = (
            f"@{config.LOG_GROUP_ID}"
            if isinstance(config.LOG_GROUP_ID, str)
            and not config.LOG_GROUP_ID.startswith("@")
            else config.LOG_GROUP_ID
        )

        for owner_id in config.OWNER_ID:
            try:
                await self.set_bot_commands(
                    owner_commands,
                    scope=types.BotCommandScopeChatMember(
                        chat_id=LOG_GROUP_ID, user_id=owner_id
                    ),
                )
                await self.set_bot_commands(
                    private_commands + owner_commands,
                    scope=types.BotCommandScopeChat(chat_id=owner_id),
                )
            except Exception:
                pass

    def load_plugin(self, file_path: str, base_dir: str, utils=None):
        file_name = os.path.basename(file_path)
        module_name, ext = os.path.splitext(file_name)
        if module_name.startswith("__") or ext != ".py":
            return None

        relative_path = os.path.relpath(file_path, base_dir).replace(os.sep, ".")
        module_path = f"{os.path.basename(base_dir)}.{relative_path[:-3]}"

        spec = importlib.util.spec_from_file_location(module_path, file_path)
        module = importlib.util.module_from_spec(spec)
        module.logger = LOGGER(module_path)
        module.app = self
        module.Config = config

        if utils:
            module.utils = utils

        try:
            spec.loader.exec_module(module)
            self.loaded_plug_counts += 1
        except Exception as e:
            LOGGER(__name__).error(
                f"Failed to load {module_path}: {e}\n\n", exc_info=True
            )
            exit()

        return module

    def load_plugins_from(self, base_folder: str):
        base_dir = os.path.abspath(base_folder)
        utils_path = os.path.join(base_dir, "utils.py")
        utils = None

        if os.path.exists(utils_path) and os.path.isfile(utils_path):
            try:
                spec = importlib.util.spec_from_file_location("utils", utils_path)
                utils = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(utils)
            except Exception as e:
                LOGGER(__name__).error(
                    f"Failed to load 'utils' module: {e}", exc_info=True
                )

        for root, _, files in os.walk(base_dir):
            for file in files:
                if file.endswith(".py") and not file == "utils.py":
                    file_path = os.path.join(root, file)
                    mod = self.load_plugin(file_path, base_dir, utils)
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

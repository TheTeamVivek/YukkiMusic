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
import inspect
import traceback
import logging
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime
from functools import wraps

from telethon import TelegramClient, errors, events
from telethon.tl import functions, types

import config

log = logging.getLogger(__name__)


@dataclass
class ShellRunResult:
    returncode: int
    stdout: str
    stderr: str
        
commands = {
    "en": {
        private : [
            types.BotCommand("start", "Start the bot"),
            types.BotCommand("help", "Get the help menu"),
            types.BotCommand("ping", "Check if the bot is alive or dead"),
        ],
        group : [types.BotCommand("play", "Start playing requested song")],

        admin : [
            types.BotCommand("play", "Start playing requested song"),
            types.BotCommand("skip", "Move to next track in queue"),
            types.BotCommand("pause", "Pause the current playing song"),
            types.BotCommand("resume", "Resume the paused song"),
            types.BotCommand("end", "Clear the queue and leave voice chat"),
            types.BotCommand("shuffle", "Randomly shuffle the queued playlist"),
            types.BotCommand("playmode", "Change the default playmode for your chat"),
            types.BotCommand("settings", "Open bot settings for your chat"),
            types.BotCommand("reboot", "Reboot  the bot for your chat"),
            
        ],
        owner: [
            types.BotCommand("autoend", "Enable or disable auto end for streams"),
            types.BotCommand("restart", "Restart the bot"),
            types.BotCommand("update", "Update the bot"),
            types.BotCommand("logs", "Get logs"),
            types.BotCommand("export", "Export all data of mongodb"),
            types.BotCommand("import", "Import all data in mongodb"),
            types.BotCommand("addsudo", "Add a user as a sudoer"),
            types.BotCommand("delsudo", "Remove a user from sudoers"),
            types.BotCommand("sudolist", "List all sudo users"),
            types.BotCommand("log", "Get the bot logs"),
            types.BotCommand("getvar", "Get a specific environment variable"),
            types.BotCommand("delvar", "Delete a specific environment variable"),
            types.BotCommand("setvar", "Set a specific environment variable"),
            #types.BotCommand("usage", "Get dyno usage information"),
            types.BotCommand("maintenance", "Enable or disable maintenance mode"),
            types.BotCommand("logger", "Enable or disable logging"),
            types.BotCommand("block", "Block a user"),
            types.BotCommand("unblock", "Unblock a user"),
            types.BotCommand("blacklist", "Blacklist a chat"),
            types.BotCommand("whitelist", "Whitelist a chat"),
            types.BotCommand("blacklisted", "List all blacklisted chats"),
        ],
    }            
}                    


class TelethonClient(TelegramClient):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def start(self, *args, **kwargs):
        await super().start(*args, **kwargs)
        me = await self.get_me()
        # pylint: disable=attribute-defined-outside-init
        self.me = me
        self.id = me.id
        self.name = f"{me.first_name} {me.last_name or ''}".strip()
        self.username = me.username
        self.mention = f"[{self.name}](tg://user?id={self.id})"
        try:
            await self.send_message(
                entity=config.LOG_GROUP_ID,
                message=(
                    f"<u><b>{self.mention} Bot Started :</b></u>\n\n"
                    f"Id : <code>{self.id}</code>\n"
                    f"Name : {self.name}\n"
                    f"Username : @{self.username}"
                ),
                parse_mode="HTML",
            )
        except (errors.ChatIdInvalidError, errors.ChatAdminRequiredError):
            log.error(
                "Bot failed to access the log group. Ensure the bot is added and promoted as admin."
            )
            sys.exit()

        try:
            _, status = await self.get_chat_member(config.LOG_GROUP_ID, "me")
            if status != "ADMIN":
                log.error("Please promote bot as admin in logger group")
                sys.exit()
        except Exception:
            pass
        log.info("MusicBot started as %s", self.name)
        if config.SET_CMDS:
            await self._set_default_commands()

    async def run_coro(self, func: Callable, err: bool = True, *args, **kwargs):
        try:
            if inspect.iscoroutinefunction(func):
                r = await func(*args, **kwargs)
            else:
                r = await asyncio.to_thread(func, *args, **kwargs)
            return r
        except Exception as e:
            if err:
                raise e

    async def create_mention(self, user: types.User | int, html: bool = False) -> str:
        if isinstance(user, int):
            user = await self.get_entity(user)
        user_name = f"{user.first_name} {user.last_name or ''}".strip()
        user_id = user.id
        if html:
            return f'<a href="tg://user?id={user_id}">{user_name}</a>'
        return f"[{user_name}](tg://user?id={user_id})"

    async def leave_chat(self, chat):
        await self.kick_participant(chat, "me")

    async def get_chat_member(
        self,
        chat_id: int | str,
        user_id: int | str,
    ):
        chat = await self.get_entity(chat_id)
        user = await self.get_entity(user_id)

        status_map = {
            "BANNED": types.ChannelParticipantBanned,
            "LEFTED": types.ChannelParticipantLeft,
            "OWNER": types.ChannelParticipantCreator,
            "ADMIN": types.ChannelParticipantAdmin,
            "SELF": types.ChannelParticipantSelf,
            "MEMBER": types.ChannelParticipant,
        }

        if isinstance(chat, types.Chat):
            r = await self(functions.messages.GetFullChatRequest(chat_id=chat.chat_id))

            members = getattr(r.full_chat.participants, "participants", [])

            for member in members:
                if member.user_id == user.user_id:
                    for status, cls in status_map.items():
                        if isinstance(member, cls):
                            return member, status
            raise errors.UserNotParticipantError

        elif isinstance(chat, types.Channel):
            r = await self(
                functions.channels.GetParticipantRequest(channel=chat, participant=user)
            )
            participant = r.participant
            for status, cls in status_map.items():
                if isinstance(participant, cls):
                    return participant, status
        else:
            raise ValueError(f'The chat_id "{chat_id}" belongs to a user')

    async def handle_error(self, exc: Exception, e = None):  # TODO Make it more brief
        date_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        error_trace = traceback.format_exc()

        error_message = (
            f"**Error:** {type(exc).__name__}\n"
            f"**Date:** {date_time}\n"
            f"**Traceback:**\n```python\n{error_trace}```\n"
        )

        await self.send_message(config.LOG_GROUP_ID, error_message)

        try:
            await self.send_message(config.OWNER_ID[0], error_message)
        except Exception:
            pass
        log.error(error_trace)

    def on_message(self, func=None, *args, **kwargs):
        def decorator(function):
            @wraps(function)
            async def wrapper(event):
                try:
                    return await function(event)
                except errors.FloodWaitError as e:
                    log.warning("FloodWait: Sleeping for %d seconds.", e.value)
                    await asyncio.sleep(e.value)
                except (
                    errors.ChatWriteForbiddenError,
                    errors.ChatSendMediaForbiddenError,
                    errors.ChatSendPhotosForbiddenError,
                    errors.MessageNotModifiedError,
                    errors.MessageIdInvalidError,
                ) as e:
                    pass
                #  if isinstance(e, errors.ChatWriteForbiddenError):
                #      await self.run_coro(
                #          event.chat_id, func=self.leave_chat, err=False
                #      )  # using for disable errors

                except events.StopPropagation as e:
                    raise events.StopPropagation from e

                except Exception as e:
                    await self.handle_error(e)

            if func is not None:
                kwargs["func"] = func
            kwargs["incoming"] = kwargs.get("incoming", True)
            self.add_event_handler(wrapper, events.NewMessage(*args, **kwargs))
            return wrapper

        return decorator

    async def run_shell_command(self, command: list):
        process = await asyncio.create_subprocess_exec(
            *command,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )

        stdout, stderr = await process.communicate()

        return ShellRunResult(
            returncode=process.returncode,
            stdout=stdout.decode().strip() if stdout else None,
            stderr=stderr.decode().strip() if stderr else None,
        )

    async def _set_default_commands(self):
        for lang_code, command in commands.items():
            try:
                await self.set_bot_commands(
                    commands=command["private"], 
                    lang_code=lang_code, 
                    scope=types.BotCommandScopeUsers()
                )
                await self.set_bot_commands(
                    commands=command["group"], 
                    lang_code=lang_code,
                    scope=types.BotCommandScopeChats()
                )
                await self.set_bot_commands(
                   commands=command["admin"], 
                   lang_code=lang_code, 
                   scope=types.BotCommandScopeChatAdmins()
                )
                
                logger_id = config.LOG_GROUP_ID
                for id in config.OWNER_ID:
                    await self.set_bot_commands(
                        commands=command["owner"],
                        lang_code=lang_code,
                        scope=types.BotCommandScopePeerUser(peer=logger_id, user_id=id),
                    )
                    await self.set_bot_commands(
                        commands=command["private"] + command["owner"],
                        lang_code=lang_code,
                        scope=types.BotCommandScopePeer(peer=id),
                    )
            except Exception:
                pass

    async def set_bot_commands(self, scope, commands: list[types.BotCommand], lang_code: str = "en"):
        return await self(
            functions.bots.SetBotCommandsRequest(
                scope=scope,
                lang_code=lang_code,
                commands=commands,
            )
        )

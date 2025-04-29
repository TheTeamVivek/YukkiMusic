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
import logging

from ntgcalls import TelegramServerError
from pyrogram.errors import (
    ChannelsTooMuch,
    FloodWait,
    InviteRequestSent,
    UserAlreadyParticipant,
)
from pytgcalls import PyTgCalls, filters
from pytgcalls.exceptions import AlreadyJoinedError
from pytgcalls.types import (
    ChatUpdate,
    GroupCallConfig,
    MediaStream,
    StreamAudioEnded,
)
from telethon.errors import ChatAdminRequiredError
from telethon.tl.functions.messages import (
    ExportChatInviteRequest,
    HideChatJoinRequestRequest,
)

import config
from strings import get_string
from YukkiMusic import tbot, userbot
from YukkiMusic.core.userbot import assistants
from YukkiMusic.misc import db
from YukkiMusic.utils.database import (
    add_active_chat,
    add_active_video_chat,
    get_assistant,
    get_audio_bitrate,
    get_lang,
    get_loop,
    get_video_bitrate,
    group_assistant,
    music_on,
    remove_active_chat,
    remove_active_video_chat,
    set_assistant,
    set_loop,
)
from YukkiMusic.utils.exceptions import AssistantErr
from YukkiMusic.utils.formatters import seconds_to_min
from YukkiMusic.utils.inline.play import play_markup
from YukkiMusic.utils.stream.autoclear import auto_clean
from YukkiMusic.utils.thumbnails import gen_thumb

links = {}
logger = logging.getLogger(__name__)


def _clean(data: list | dict):
    if isinstance(data, list):
        for element in data:
            if msg := element.get("mystic"):
                config.add_to_clean(msg.chat_id, msg.id)
    else:
        if msg := data.get("mystic"):
            config.add_to_clean(msg.chat_id, msg.id)


async def clear(chat_id):
    try:
        popped = db.pop(chat_id, None)
        if popped:
            await auto_clean(popped)
        _clean(popped)
        db[chat_id] = []
        await remove_active_video_chat(chat_id)
        await remove_active_chat(chat_id)
        await set_loop(chat_id, 0)
    except Exception:
        logger.error("", exc_info=True)


class Call:
    def __init__(self):
        self.clients = [
            PyTgCalls(
                client,
                cache_duration=100,
            )
            for client in userbot.clients
        ]

    async def pause_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.pause_stream(chat_id)

    async def resume_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.resume_stream(chat_id)

    async def mute_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.mute_stream(chat_id)

    async def unmute_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.unmute_stream(chat_id)

    async def stop_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        try:
            await clear(chat_id)
            await assistant.leave_call(chat_id)
        except Exception:
            pass

    async def force_stop_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        try:
            check = db.get(chat_id)
            check.pop(0)
        except Exception:
            pass
        await remove_active_video_chat(chat_id)
        await remove_active_chat(chat_id)
        try:
            await assistant.leave_call(chat_id)
        except Exception:
            pass

    async def skip_stream(
        self,
        chat_id: int,
        file_path: str,
        video: bool | str = None,
        image: bool | str = None,
    ):
        audio_stream_quality = await get_audio_bitrate(chat_id)
        video_stream_quality = await get_video_bitrate(chat_id)
        if video:
            stream = MediaStream(
                file_path,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        elif image and config.PRIVATE_BOT_MODE:
            stream = MediaStream(
                image,
                audio_path=file_path,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        else:
            stream = MediaStream(
                file_path,
                audio_parameters=audio_stream_quality,
                video_flags=MediaStream.Flags.IGNORE,
            )

        await self.play(chat_id, stream)

    async def seek_stream(self, chat_id, file_path, to_seek, duration, mode):
        audio_stream_quality = await get_audio_bitrate(chat_id)
        video_stream_quality = await get_video_bitrate(chat_id)
        stream = (
            MediaStream(
                file_path,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
                ffmpeg_parameters=f"-ss {to_seek} -to {duration}",
            )
            if mode == "video"
            else MediaStream(
                file_path,
                audio_parameters=audio_stream_quality,
                ffmpeg_parameters=f"-ss {to_seek} -to {duration}",
                video_flags=MediaStream.Flags.IGNORE,
            )
        )
        await self.play(chat_id, stream)

    async def stream_call(self, link):
        assistant = await group_assistant(self, config.LOG_GROUP_ID)
        await assistant.play(
            config.LOG_GROUP_ID,
            MediaStream(link),
            config=GroupCallConfig(auto_start=False),
        )
        await asyncio.sleep(0.5)
        await assistant.leave_call(config.LOG_GROUP_ID)

    async def join_chat(self, chat_id, attempts=1):
        max_attempts = len(assistants) - 1
        assistant = await get_assistant(chat_id)
        try:
            language = await get_lang(chat_id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")
        try:
            chat = await tbot.get_entity(chat_id)
        except ChatAdminRequiredError as e:
            raise AssistantErr(_["call_1"]) from e
        except Exception as e:
            raise AssistantErr(
                _["call_3"].format(tbot.mention, type(e).__name__)
            ) from e
        if chat_id in links:
            invitelink = links[chat_id]
        else:
            if hasattr(chat, "username") and chat.username:
                invitelink = chat.username
                try:
                    await assistant.resolve_peer(invitelink)
                except Exception:
                    pass
            else:
                try:
                    invitelink = await tbot(
                        ExportChatInviteRequest(chat_id, legacy_revoke_permanent=True)
                    )
                    invitelink = invitelink.link
                except ChatAdminRequiredError as e:
                    raise AssistantErr(_["call_1"]) from e
                except Exception as e:
                    raise AssistantErr(
                        _["call_3"].format(tbot.mention, type(e).__name__)
                    ) from e

            if invitelink.startswith("https://t.me/+"):
                invitelink = invitelink.replace(
                    "https://t.me/+", "https://t.me/joinchat/"
                )
            links[chat_id] = invitelink

        try:
            await asyncio.sleep(1)
            await assistant.join_chat(invitelink)
        except InviteRequestSent:
            try:
                await tbot(HideChatJoinRequestRequest(chat_id, userbot.id))
            except Exception as e:
                raise AssistantErr(_["call_3"].format(type(e).__name__)) from e
            await asyncio.sleep(1)
            raise AssistantErr(_["call_6"].format(tbot.mention))
        except UserAlreadyParticipant:
            pass
        except ChannelsTooMuch as e:
            if attempts <= max_attempts:
                attempts += 1
                await set_assistant(chat_id)
                return await self.join_chat(chat_id, attempts)
            raise AssistantErr(_["call_9"].format(config.SUPPORT_GROUP)) from e
        except FloodWait as e:
            time = e.value
            if time < 20:
                await asyncio.sleep(time)
                attempts += 1
                return await self.join_chat(chat_id, attempts)
            if attempts <= max_attempts:
                attempts += 1
                await set_assistant(chat_id)
                return await self.join_chat(chat_id, attempts)

            raise AssistantErr(_["call_10"].format(time)) from e
        except Exception as e:
            raise AssistantErr(_["call_3"].format(type(e).__name__)) from e

    async def play(self, chat_id, stream=None, config=None):
        assistant = await group_assistant(self, chat_id)
        if config is None:
            config = GroupCallConfig(auto_start=False)
        await assistant.play(chat_id, stream=stream, config=config)

    async def join_call(
        self,
        _,
        chat_id: int,
        file_path,
        video: bool = False,
        image: str | None = None,
    ):
        audio_stream_quality = await get_audio_bitrate(chat_id)
        video_stream_quality = await get_video_bitrate(chat_id)
        if video:
            stream = MediaStream(
                file_path,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        elif image and config.PRIVATE_BOT_MODE:
            stream = MediaStream(
                image,
                audio_path=file_path,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        else:
            stream = MediaStream(
                file_path,
                audio_parameters=audio_stream_quality,
                video_flags=MediaStream.Flags.IGNORE,
            )

        try:
            await self.play(
                chat_id=chat_id,
                stream=stream,
            )
        except Exception:
            await self.join_chat(chat_id)
            try:
                await self.play(
                    chat_id=chat_id,
                    stream=stream,
                )
            except Exception as e:
                raise AssistantErr(_["call_11"]) from e

        except AlreadyJoinedError as e:
            raise AssistantErr(_["call_12"]) from e
        except TelegramServerError as e:
            raise AssistantErr(_["call_13"]) from e
        await add_active_chat(chat_id)
        await music_on(chat_id)
        if video:
            await add_active_video_chat(chat_id)

    async def change_stream(self, chat_id):
        check = db.get(chat_id)
        popped = None
        loop = await get_loop(chat_id)
        try:
            if loop == 0:
                popped = check.pop(0)
                if popped:
                    await auto_clean(popped)
                    _clean(popped)

            else:
                loop = loop - 1
                await set_loop(chat_id, loop)

            if not check:
                return await self.stop_stream(chat_id)
        except Exception:
            return await self.stop_stream(chat_id)

        else:
            language = await get_lang(chat_id)
            _ = get_string(language)
            audio_stream_quality = await get_audio_bitrate(chat_id)
            video_stream_quality = await get_video_bitrate(chat_id)

            track = check[0]["track"]
            user = check[0]["by"]
            original_chat_id = check[0]["chat_id"]
            videoid = track.vidid
            thumb = track.thumb
            check[0]["played"] = 0
            url = (
                f"https://t.me/{tbot.username}?start=info_{videoid}"
                if track.is_youtube
                else track.link
            )

            img = await gen_thumb(videoid, thumb)
            mystic = await tbot.send_message(original_chat_id, _["call_8"])
            try:
                file_path = await track.download()
            except Exception as e:
                await tbot.handle_error(e)
                return await mystic.edit(_["call_7"], link_preview=False)

            if track.video:
                stream = MediaStream(
                    file_path,
                    audio_parameters=audio_stream_quality,
                    video_parameters=video_stream_quality,
                )
            elif thumb and config.PRIVATE_BOT_MODE:
                stream = MediaStream(
                    thumb,
                    audio_path=file_path,
                    audio_parameters=audio_stream_quality,
                    video_parameters=video_stream_quality,
                )
            else:
                stream = MediaStream(
                    file_path,
                    audio_parameters=audio_stream_quality,
                    video_flags=MediaStream.Flags.IGNORE,
                )
            try:
                await self.play(chat_id, stream)
            except Exception:
                return await mystic.edit(_["call_7"])
            what, button = play_markup(_, chat_id, track)
            await mystic.delete()
            run = await tbot.send_file(
                original_chat_id,
                file=img,
                caption=_["stream_1"].format(
                    track.title[:27],
                    url,
                    seconds_to_min(track.duration),
                    user,
                ),
                buttons=button,
            )
            db[chat_id][0]["mystic"] = run
            db[chat_id][0]["markup"] = what

    async def ping(self):
        pings = [client.ping for client in self.clients]
        if pings:
            return str(round(sum(pings) / len(pings), 3))

    async def start(self):
        """Starts all PyTgCalls instances for the existing userbot clients."""
        logger.info("Starting PyTgCall Clients")
        await asyncio.gather(*[client.start() for client in self.clients])
        await self._decorators()

    async def stop(self):
        await asyncio.gather(*[self.stop_stream(chat_id) for chat_id in self.calls])

    async def _decorators(self):
        for client in self.clients:

            @client.on_update(filters.chat_update(ChatUpdate.Status.LEFT_CALL))
            async def stream_services_handler(_, update):
                await self.stop_stream(update.chat_id)

            @client.on_update(filters.stream_end)
            async def stream_end_handler(_, update):
                if not isinstance(update, StreamAudioEnded):
                    return
                await self.change_stream(update.chat_id)


Yukki = Call()

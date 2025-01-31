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

from ntgcalls import TelegramServerError
from pyrogram.errors import (
    ChannelsTooMuch,
    FloodWait,
    InviteRequestSent,
    UserAlreadyParticipant,
)
from pyrogram.types import InlineKeyboardMarkup
from pytgcalls import PyTgCalls, filters
from pytgcalls.exceptions import AlreadyJoinedError
from pytgcalls.types import (
    ChatUpdate,
    GroupCallConfig,
    MediaStream,
    StreamAudioEnded,
    Update,
)
from telethon.errors import ChatAdminRequiredError
from telethon.tl.functions.messages import (
    ExportChatInviteRequest,
    HideChatJoinRequestRequest,
)

import config
from strings import get_string
from YukkiMusic import Platform, app, logger, tbot, userbot
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
from YukkiMusic.utils.inline.play import stream_markup, telegram_markup
from YukkiMusic.utils.stream.autoclear import auto_clean
from YukkiMusic.utils.thumbnails import gen_thumb

from .enum import PlayType

links = {}


async def _clear_(chat_id):
    popped = db.pop(chat_id, None)
    if popped:
        await auto_clean(popped)
    db[chat_id] = []
    await remove_active_video_chat(chat_id)
    await remove_active_chat(chat_id)
    await set_loop(chat_id, 0)


class Call:
    def __init__(self):
        self.calls: dict[int, PlayType] = {}
        self.clients = [
            PyTgCalls(
                client,
                cache_duration=100,
            )
            for client in userbot.clients
        ]

    async def is_active_chat(self, chat_id: int) -> bool:
        return bool(chat_id in self.calls)

    async def add_active_chat(self, chat_id: int):
        if chat_id not in self.calls:
            self.calls[chat_id] = PlayType.PLAYING

    async def remove_active_chat(self, chat_id: int):
        if chat_id in self.calls:
            del self.calls[chat_id]

    async def pause_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.pause_stream(chat_id)
        self.calls[chat_id] = PlayType.PAUSED

    async def resume_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.resume_stream(chat_id)
        self.calls[chat_id] = PlayType.PLAYING

    async def mute_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.mute_stream(chat_id)
        self.calls[chat_id] = PlayType.MUTED

    async def unmute_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        await assistant.unmute_stream(chat_id)
        self.calls[chat_id] = PlayType.PLAYING

    async def stop_stream(self, chat_id: int):
        assistant = await group_assistant(self, chat_id)
        try:
            await _clear_(chat_id)
            await assistant.leave_call(chat_id)
            del self.calls[chat_id]
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
        link: str,
        video: bool | str = None,
        image: bool | str = None,
    ):
        assistant = await group_assistant(self, chat_id)
        audio_stream_quality = await get_audio_bitrate(chat_id)
        video_stream_quality = await get_video_bitrate(chat_id)
        call_config = GroupCallConfig(auto_start=False)
        if video:
            stream = MediaStream(
                link,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        elif image and config.PRIVATE_BOT_MODE == str(True):
            stream = MediaStream(
                image,
                audio_path=link,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        else:
            stream = MediaStream(
                link,
                audio_parameters=audio_stream_quality,
                video_flags=MediaStream.Flags.IGNORE,
            )

        await assistant.play(chat_id, stream, config=call_config)

    async def seek_stream(self, chat_id, file_path, to_seek, duration, mode):
        assistant = await group_assistant(self, chat_id)
        audio_stream_quality = await get_audio_bitrate(chat_id)
        video_stream_quality = await get_video_bitrate(chat_id)
        call_config = GroupCallConfig(auto_start=False)
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
        await assistant.play(chat_id, stream, config=call_config)

    async def stream_call(self, link):
        assistant = await group_assistant(self, config.LOG_GROUP_ID)
        call_config = GroupCallConfig(auto_start=False)
        await assistant.play(
            config.LOG_GROUP_ID,
            MediaStream(link),
            config=call_config,
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
            raise AssistantErr(_["call_3"].format(app.mention, type(e).__name__)) from e
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
                        _["call_3"].format(app.mention, type(e).__name__)
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
            raise AssistantErr(_["call_6"].format(app.mention))
        except UserAlreadyParticipant:
            pass
        except ChannelsTooMuch as e:
            if attempts <= max_attempts:
                attempts += 1
                userbot = await set_assistant(chat_id)
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
                userbot = await set_assistant(chat_id)
                return await self.join_chat(chat_id, attempts)

            raise AssistantErr(_["call_10"].format(time)) from e
        except Exception as e:
            raise AssistantErr(_["call_3"].format(type(e).__name__)) from e

    async def join_call(
        self,
        chat_id: int,
        original_chat_id: int,  # TODO remove it and make compatible
        link,
        video: bool | str = None,
        image: bool | str = None,
    ):
        assistant = await group_assistant(self, chat_id)
        audio_stream_quality = await get_audio_bitrate(chat_id)
        video_stream_quality = await get_video_bitrate(chat_id)
        call_config = GroupCallConfig(auto_start=False)
        if video:
            stream = MediaStream(
                link,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        elif image and config.PRIVATE_BOT_MODE == str(True):
            stream = MediaStream(
                image,
                audio_path=link,
                audio_parameters=audio_stream_quality,
                video_parameters=video_stream_quality,
            )
        else:
            stream = MediaStream(
                link,
                audio_parameters=audio_stream_quality,
                video_flags=MediaStream.Flags.IGNORE,
            )

        try:
            await assistant.play(
                chat_id=chat_id,
                stream=stream,
                config=call_config,
            )
        except Exception:
            await self.join_chat(chat_id)
            try:
                await assistant.play(
                    chat_id=chat_id,
                    stream=stream,
                    config=call_config,
                )
            except Exception as e:
                raise AssistantErr(
                    "**No Active Voice Chat Found**\n\n"
                    "Please make sure group's voice chat is enabled. "
                    "If already enabled, please end it and start fresh voice chat again "
                    "and if the problem continues, try /restart"
                ) from e

        except AlreadyJoinedError as e:
            raise AssistantErr(
                "**ASSISTANT IS ALREADY IN VOICECHAT **\n\n"
                "Music bot system detected that assistant is already in the voicechat, "
                "if the problem continues restart the videochat and try again."
            ) from e
        except TelegramServerError as e:
            raise AssistantErr(
                "**TELEGRAM SERVER ERROR**\n\n" "Please restart Your voicechat."
            ) from e
        await add_active_chat(chat_id)
        await music_on(chat_id)
        if video:
            await add_active_video_chat(chat_id)

    async def change_stream(self, client, chat_id):
        check = db.get(chat_id)
        popped = None
        loop = await get_loop(chat_id)
        try:
            if loop == 0:
                popped = check.pop(0)
            else:
                loop = loop - 1
                await set_loop(chat_id, loop)
            if popped:
                await auto_clean(popped)
                if popped.get("mystic"):
                    try:
                        await popped.get("mystic").delete()
                    except Exception:
                        pass
            if not check:
                await _clear_(chat_id)
                return await client.leave_call(chat_id)
        except Exception:
            try:
                await _clear_(chat_id)
                return await client.leave_call(chat_id)
            except Exception:
                return
        else:
            queued = check[0]["file"]
            language = await get_lang(chat_id)
            _ = get_string(language)
            title = (check[0]["title"]).title()
            user = check[0]["by"]
            original_chat_id = check[0]["chat_id"]
            streamtype = check[0]["streamtype"]
            audio_stream_quality = await get_audio_bitrate(chat_id)
            video_stream_quality = await get_video_bitrate(chat_id)
            videoid = check[0]["vidid"]
            userid = check[0].get("user_id")
            check[0]["played"] = 0
            video = True if str(streamtype) == "video" else False
            call_config = GroupCallConfig(auto_start=False)
            if "live_" in queued:
                n, link = await Platform.youtube.video(videoid, True)
                if n == 0:
                    return await app.send_message(
                        original_chat_id,
                        text=_["call_7"],
                    )
                if video:
                    stream = MediaStream(
                        link,
                        audio_parameters=audio_stream_quality,
                        video_parameters=video_stream_quality,
                    )
                else:
                    try:
                        image = await Platform.youtube.thumbnail(videoid, True)
                    except Exception:
                        image = None
                    if image and config.PRIVATE_BOT_MODE == str(True):
                        stream = MediaStream(
                            image,
                            audio_path=link,
                            audio_parameters=audio_stream_quality,
                            video_parameters=video_stream_quality,
                        )
                    else:
                        stream = MediaStream(
                            link,
                            audio_parameters=audio_stream_quality,
                            video_flags=MediaStream.Flags.IGNORE,
                        )
                try:
                    await client.play(chat_id, stream, config=call_config)
                except Exception:
                    return await app.send_message(
                        original_chat_id,
                        text=_["call_7"],
                    )
                img = await gen_thumb(videoid)
                button = telegram_markup(_, chat_id)
                run = await app.send_photo(
                    original_chat_id,
                    photo=img,
                    caption=_["stream_1"].format(
                        title[:27],
                        f"https://t.me/{app.username}?start=info_{videoid}",
                        check[0]["dur"],
                        user,
                    ),
                    reply_markup=InlineKeyboardMarkup(button),
                )
                db[chat_id][0]["mystic"] = run
                db[chat_id][0]["markup"] = "tg"
            elif "vid_" in queued:
                mystic = await app.send_message(original_chat_id, _["call_8"])
                try:
                    file_path, direct = await Platform.youtube.download(
                        videoid,
                        mystic,
                        videoid=True,
                        video=True if str(streamtype) == "video" else False,
                    )
                except Exception:
                    return await mystic.edit_text(
                        _["call_7"], disable_web_page_preview=True
                    )
                if video:
                    stream = MediaStream(
                        file_path,
                        audio_parameters=audio_stream_quality,
                        video_parameters=video_stream_quality,
                    )
                else:
                    try:
                        image = await Platform.youtube.thumbnail(videoid, True)
                    except Exception:
                        image = None
                    if image and config.PRIVATE_BOT_MODE == str(True):
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
                    await client.play(chat_id, stream, config=call_config)
                except Exception:
                    return await app.send_message(
                        original_chat_id,
                        text=_["call_7"],
                    )
                img = await gen_thumb(videoid)
                button = stream_markup(_, videoid, chat_id)
                await mystic.delete()
                run = await app.send_photo(
                    original_chat_id,
                    photo=img,
                    caption=_["stream_1"].format(
                        title[:27],
                        f"https://t.me/{app.username}?start=info_{videoid}",
                        check[0]["dur"],
                        user,
                    ),
                    reply_markup=InlineKeyboardMarkup(button),
                )
                db[chat_id][0]["mystic"] = run
                db[chat_id][0]["markup"] = "stream"
            elif "index_" in queued:
                stream = (
                    MediaStream(
                        videoid,
                        audio_parameters=audio_stream_quality,
                        video_parameters=video_stream_quality,
                    )
                    if str(streamtype) == "video"
                    else MediaStream(
                        videoid,
                        audio_parameters=audio_stream_quality,
                        video_flags=MediaStream.Flags.IGNORE,
                    )
                )
                try:
                    await client.play(chat_id, stream, config=call_config)
                except Exception:
                    return await app.send_message(
                        original_chat_id,
                        text=_["call_7"],
                    )
                button = telegram_markup(_, chat_id)
                run = await app.send_photo(
                    original_chat_id,
                    photo=config.STREAM_IMG_URL,
                    caption=_["stream_2"].format(user),
                    reply_markup=InlineKeyboardMarkup(button),
                )
                db[chat_id][0]["mystic"] = run
                db[chat_id][0]["markup"] = "tg"
            else:
                url = check[0].get("url")
                if videoid == "telegram":
                    image = None
                elif videoid == "soundcloud":
                    image = None

                elif "saavn" in videoid:
                    url = check[0].get("url")
                    details = await Platform.saavn.info(url)
                    image = details["thumb"]
                else:
                    try:
                        image = await Platform.youtube.thumbnail(videoid, True)
                    except Exception:
                        image = None
                if video:
                    stream = MediaStream(
                        queued,
                        audio_parameters=audio_stream_quality,
                        video_parameters=video_stream_quality,
                    )
                else:
                    if image and config.PRIVATE_BOT_MODE == str(True):
                        stream = MediaStream(
                            image,
                            audio_path=queued,
                            audio_parameters=audio_stream_quality,
                            video_parameters=video_stream_quality,
                        )
                    else:
                        stream = MediaStream(
                            queued,
                            audio_parameters=audio_stream_quality,
                            video_flags=MediaStream.Flags.IGNORE,
                        )
                try:
                    await client.play(chat_id, stream, config=call_config)
                except Exception:
                    return await app.send_message(
                        original_chat_id,
                        text=_["call_7"],
                    )
                if videoid == "telegram":
                    button = telegram_markup(_, chat_id)
                    run = await app.send_photo(
                        original_chat_id,
                        photo=(
                            config.TELEGRAM_AUDIO_URL
                            if str(streamtype) == "audio"
                            else config.TELEGRAM_VIDEO_URL
                        ),
                        caption=_["stream_1"].format(
                            title, config.SUPPORT_GROUP, check[0]["dur"], user
                        ),
                        reply_markup=InlineKeyboardMarkup(button),
                    )
                    db[chat_id][0]["mystic"] = run
                    db[chat_id][0]["markup"] = "tg"
                elif videoid == "soundcloud":
                    button = telegram_markup(_, chat_id)
                    run = await app.send_photo(
                        original_chat_id,
                        photo=config.SOUNCLOUD_IMG_URL,
                        caption=_["stream_1"].format(
                            title, config.SUPPORT_GROUP, check[0]["dur"], user
                        ),
                        reply_markup=InlineKeyboardMarkup(button),
                    )
                    db[chat_id][0]["mystic"] = run
                    db[chat_id][0]["markup"] = "tg"
                elif "saavn" in videoid:
                    button = telegram_markup(_, chat_id)
                    run = await app.send_photo(
                        original_chat_id,
                        photo=image,
                        caption=_["stream_1"].format(title, url, check[0]["dur"], user),
                        reply_markup=InlineKeyboardMarkup(button),
                    )
                    db[chat_id][0]["mystic"] = run
                    db[chat_id][0]["markup"] = "tg"

                else:
                    img = await gen_thumb(videoid)
                    button = stream_markup(_, videoid, chat_id)
                    run = await app.send_photo(
                        original_chat_id,
                        photo=img,
                        caption=_["stream_1"].format(
                            title[:27],
                            f"https://t.me/{app.username}?start=info_{videoid}",
                            check[0]["dur"],
                            user,
                        ),
                        reply_markup=InlineKeyboardMarkup(button),
                    )
                    db[chat_id][0]["mystic"] = run
                    db[chat_id][0]["markup"] = "stream"

    async def ping(self):
        pings = [client.ping for client in self.clients]
        if pings:
            return str(round(sum(pings) / len(pings), 3))
        else:
            logger(__name__).error("No active clients for ping calculation.")
            return "No active clients"

    async def start(self):
        """Starts all PyTgCalls instances for the existing userbot clients."""
        logger(__name__).info("Starting PyTgCall Clients")
        await asyncio.gather(*[client.start() for client in self.clients])

    async def stop(self):
        await asyncio.gather(*[self.stop_stream(chat_id) for chat_id in self.calls])

    async def decorators(self):
        for client in self.clients:

            @client.on_update(filters.chat_update(ChatUpdate.Status.LEFT_CALL))
            async def stream_services_handler(client, update):
                await self.stop_stream(update.chat_id)

            @client.on_update(filters.stream_end)
            async def stream_end_handler(client, update: Update):
                if not isinstance(update, StreamAudioEnded):
                    return
                await self.change_stream(client, update.chat_id)


Yukki = Call()

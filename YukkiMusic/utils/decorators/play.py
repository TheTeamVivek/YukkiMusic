#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio

from pyrogram.enums import ChatMemberStatus
from pyrogram.errors import (
    ChannelsTooMuch,
    ChatAdminRequired,
    FloodWait,
    InviteRequestSent,
    UserAlreadyParticipant,
    UserNotParticipant,
)
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from config import PLAYLIST_IMG_URL, PRIVATE_BOT_MODE
from config import SUPPORT_GROUP as SUPPORT_CHAT
from config import adminlist
from strings import get_string
from YukkiMusic import YouTube, app
from YukkiMusic.core.call import Yukki
from YukkiMusic.core.userbot import assistants
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_assistant,
    get_cmode,
    get_lang,
    get_playmode,
    get_playtype,
    is_active_chat,
    is_commanddelete_on,
    is_maintenance,
    is_served_private_chat,
    set_assistant,
)
from YukkiMusic.utils.inline import botplaylist_markup

links = {}


async def join_chat(message, chat_id, _, myu, attempts=1):
    max_attempts = len(assistants) - 1  # Set the maximum number of attempts
    userbot = await get_assistant(chat_id)

    if chat_id in links:
        invitelink = links[chat_id]
    else:
        if message.chat.username:
            invitelink = message.chat.username
            try:
                await userbot.resolve_peer(invitelink)
            except:
                pass
        else:
            try:
                invitelink = await app.export_chat_invite_link(message.chat.id)
            except ChatAdminRequired:
                return await myu.edit(_["call_1"])
            except Exception as e:
                return await myu.edit(_["call_3"].format(app.mention, type(e).__name__))

        if invitelink.startswith("https://t.me/+"):
            invitelink = invitelink.replace("https://t.me/+", "https://t.me/joinchat/")
        links[chat_id] = invitelink

    try:
        await asyncio.sleep(1)
        await userbot.join_chat(invitelink)
    except InviteRequestSent:
        try:
            await app.approve_chat_join_request(chat_id, userbot.id)
        except Exception as e:
            return await myu.edit(_["call_3"].format(type(e).__name__))
        await asyncio.sleep(1)
        await myu.edit(_["call_6"].format(app.mention))
    except UserAlreadyParticipant:
        pass
    except ChannelsTooMuch:
        if attempts <= max_attempts:
            userbot = await set_assistant(chat_id)
            return await join_chat(message, chat_id, _, myu, attempts + 1)
        else:
            return await myu.edit(_["call_9"].format(SUPPORT_CHAT))
    except FloodWait as e:
        time = e.value
        if time < 20:
            await asyncio.sleep(time)
            return await join_chat(message, chat_id, _, myu, attempts + 1)
        else:
            if attempts <= max_attempts:
                userbot = await set_assistant(chat_id)
                return await join_chat(message, chat_id, _, myu, attempts + 1)

            return await myu.edit(_["call_10"].format(time))
    except Exception as e:
        return await myu.edit(_["call_3"].format(type(e).__name__))

    try:
        await myu.delete()
    except:
        pass


def PlayWrapper(command):
    async def wrapper(client, message):
        language = await get_lang(message.chat.id)
        _ = get_string(language)
        if message.sender_chat:
            upl = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            text="How to Fix ?",
                            callback_data="AnonymousAdmin",
                        ),
                    ]
                ]
            )
            return await message.reply_text(_["general_4"], reply_markup=upl)

        if await is_maintenance() is False:
            if message.from_user.id not in SUDOERS:
                return

        if PRIVATE_BOT_MODE == str(True):
            if not await is_served_private_chat(message.chat.id):
                await message.reply_text(
                    "**PRIVATE MUSIC BOT**\n\nOnly For Authorized chats from the owner ask my owner to allow your chat first."
                )
                return await app.leave_chat(message.chat.id)
        if await is_commanddelete_on(message.chat.id):
            try:
                await message.delete()
            except:
                pass

        audio_telegram = (
            (message.reply_to_message.audio or message.reply_to_message.voice)
            if message.reply_to_message
            else None
        )
        video_telegram = (
            (message.reply_to_message.video or message.reply_to_message.document)
            if message.reply_to_message
            else None
        )
        url = await YouTube.url(message)
        if audio_telegram is None and video_telegram is None and url is None:
            if len(message.command) < 2:
                if "stream" in message.command:
                    return await message.reply_text(_["str_1"])
                buttons = botplaylist_markup(_)
                return await message.reply_photo(
                    photo=PLAYLIST_IMG_URL,
                    caption=_["playlist_1"],
                    reply_markup=InlineKeyboardMarkup(buttons),
                )
        if message.command[0][0] == "c":
            chat_id = await get_cmode(message.chat.id)
            if chat_id is None:
                return await message.reply_text(_["setting_12"])
            try:
                chat = await app.get_chat(chat_id)
            except:
                return await message.reply_text(_["cplay_4"])
            channel = chat.title
        else:
            chat_id = message.chat.id
            channel = None
        try:
            is_call_active = (await app.get_chat(chat_id)).is_call_active
            if not is_call_active:
                return await message.reply_text(
                    "**No active video chat found **\n\nPlease make sure you started the voicechat."
                )
        except Exception:
            pass

        playmode = await get_playmode(message.chat.id)
        playty = await get_playtype(message.chat.id)
        if playty != "Everyone":
            if message.from_user.id not in SUDOERS:
                admins = adminlist.get(message.chat.id)
                if not admins:
                    return await message.reply_text(_["admin_18"])
                else:
                    if message.from_user.id not in admins:
                        return await message.reply_text(_["play_4"])
        if message.command[0][0] == "v":
            video = True
        else:
            if "-v" in message.text:
                video = True
            else:
                video = True if message.command[0][1] == "v" else None
        if message.command[0][-1] == "e":
            if not await is_active_chat(chat_id):
                return await message.reply_text(_["play_18"])
            fplay = True
        else:
            fplay = None

        if await is_active_chat(chat_id):
            userbot = await get_assistant(message.chat.id)
            # Getting all members id that in voicechat
            call_participants_id = [
                member.chat.id async for member in userbot.get_call_members(chat_id)
                if member.chat
            ]
            # Checking if assistant id not in list so clear queues and remove active voice chat and process

            if (not call_participants_id or userbot.id not in call_participants_id):
                await Yukki.stop_stream(chat_id)

        else:
            userbot = await get_assistant(message.chat.id)
            try:
                try:
                    get = await app.get_chat_member(chat_id, userbot.id)
                except ChatAdminRequired:
                    return await message.reply_text(_["call_1"])
                if (
                    get.status == ChatMemberStatus.BANNED
                    or get.status == ChatMemberStatus.RESTRICTED
                ):
                    try:
                        await app.unban_chat_member(chat_id, userbot.id)
                    except:
                        return await message.reply_text(
                            text=_["call_2"].format(userbot.username, userbot.id),
                        )
            except UserNotParticipant:
                myu = await message.reply_text(_["call_5"])
                await join_chat(message, chat_id, _, myu)

        return await command(
            client,
            message,
            _,
            chat_id,
            video,
            channel,
            playmode,
            url,
            fplay,
        )

    return wrapper

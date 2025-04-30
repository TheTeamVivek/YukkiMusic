#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from pyrogram.errors import ChannelPrivate
from telethon import Button
from telethon.tl import types

from config import PLAYLIST_IMG_URL, PRIVATE_BOT_MODE, adminlist
from strings import get_string
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import SUDOERS
from YukkiMusic.platforms import telegram
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
)
from YukkiMusic.utils.inline import botplaylist_markup

links = {}

__all__ = ["play_wrapper"]


def play_wrapper(func):
    async def wrapper(event):
        language = await get_lang(event.chat_id)
        _ = get_string(language)
        sender = await event.get_sender()
        if not isinstance(sender, types.User):
            upl = [
                [
                    Button.inline(
                        text=_["anon_admin"],
                        callback_data="AnonymousAdmin",
                    ),
                ]
            ]

            return await event.reply(_["general_4"], buttons=upl)

        if await is_maintenance() is False:
            if event.sender_id not in SUDOERS:
                return

        if PRIVATE_BOT_MODE:
            if not await is_served_private_chat(event.chat_id):
                await event.reply(
                    "**PRIVATE MUSIC BOT**\n\n"
                    "Only For Authorized chats from the owner"
                    "ask my owner to allow your chat first."
                )
                return await event.client.leave_chat(event.chat_id)
        if await is_commanddelete_on(event.chat_id):
            try:
                await event.delete()
            except Exception:
                pass

        command = event.text.split()
        url = await telegram.get_url_from_message(event)
        rmsg = await event.get_reply_message()
        if rmsg and rmsg.document is None and url is None:
            if len(command) < 2:
                if "/stream" in command:
                    return await event.reply(_["str_1"])
                buttons = botplaylist_markup(_)
                return await event.reply(
                    file=PLAYLIST_IMG_URL,
                    message=_["playlist_1"],
                    buttons=buttons,
                )
        if command[0][0] == "c":
            chat_id = await get_cmode(event.chat_id)
            if chat_id is None:
                return await event.reply(_["setting_12"])
            try:
                chat = await event.client.get_entity(chat_id)
            except Exception:
                return await event.reply(_["cplay_4"])
            channel = chat.title
        else:
            chat_id = event.chat_id
            channel = None
        try:
            is_call_active = (await event.client.get_entity(chat_id)).call_active
            if not is_call_active:
                return await event.reply(
                    "**No active video chat found **\n\nPlease make sure you started the voicechat."
                )
        except Exception:
            pass

        playmode = await get_playmode(event.chat_id)
        playty = await get_playtype(event.chat_id)
        if playty != "EVERYONE":
            if event.sender_id not in SUDOERS:
                admins = adminlist.get(event.chat_id)
                if not admins:
                    return await event.reply(_["admin_18"])
                else:
                    if event.sender_id not in admins:
                        return await event.reply(_["play_4"])
        if command[0][1] == "v":
            video = True
        else:
            if "-v" in event.text:
                video = True
            else:
                video = True if command[0][1] == "v" else None
        if command[0][-1] == "e":
            if not await is_active_chat(chat_id):
                return await event.reply(_["play_18"])
            fplay = True
        else:
            fplay = None
        if await is_active_chat(chat_id):
            userbot = await get_assistant(event.chat_id)
            # Getting all members id that in voicechat
            try:
                call_participants_id = [
                    member.chat.id
                    async for member in userbot.get_call_members(chat_id)
                    if member.chat
                ]
                # Checking if assistant id not in list
                # so clear queues and remove active voice chat and process

                if not call_participants_id or userbot.id not in call_participants_id:
                    await Yukki.stop_stream(chat_id)
            except ChannelPrivate:
                pass

        return await func(
            event,
            _,
            chat_id,
            video,
            channel,
            playmode,
            url,
            fplay,
        )

    return wrapper

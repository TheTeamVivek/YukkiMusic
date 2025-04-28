#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from telethon import Button
from telethon.tl.types import PeerUser

from config import adminlist
from strings import get_string
from YukkiMusic import tbot
from YukkiMusic.misc import SUDOERS

from ..formatters import int_to_alpha
from .database import (
    get_authuser_names,
    get_cmode,
    get_lang,
    is_active_chat,
    is_commanddelete_on,
    is_maintenance,
    is_nonadmin_chat,
)

__all__ = ["admin_rights_check", "admin_actual", "actual_admin_cb"]


def admin_rights_check(mystic):
    async def wrapper(event):
        if not await is_maintenance():
            if event.sender_id not in SUDOERS:
                return
        if await is_commanddelete_on(event.chat_id):
            try:
                await event.delete()
            except Exception:
                pass
        try:
            language = await get_lang(event.chat_id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")
        if not isinstance(event.message.from_id, PeerUser):
            upl = [
                [
                    Button.inline(
                        text=_["anon_admin"],
                        data="AnonymousAdmin",
                    ),
                ]
            ]
            return await event.reply(_["general_4"], buttons=upl)
        _, _, cplay = await parse_flags(event.chat_id, event.text)

        if cplay:
            chat_id = await get_cmode(event.chat_id)
            if chat_id is None:
                return await event.reply(_["setting_12"])
            try:
                await tbot.get_entity(chat_id)
            except Exception:
                return await event.reply(_["cplay_4"])
        else:
            chat_id = event.chat_id
        if not await is_active_chat(chat_id):
            return await event.reply(_["general_6"])
        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            if event.sender_id not in SUDOERS:
                admins = adminlist.get(event.chat_id)
                if not admins:
                    return await event.reply(_["admin_18"])
                else:
                    if event.sender_id not in admins:
                        return await event.reply(_["admin_19"])
        return await mystic(event, _, chat_id)

    return wrapper


def admin_actual(mystic):
    async def wrapper(event):
        if not await is_maintenance():
            if event.sender_id not in SUDOERS:
                return

        if await is_commanddelete_on(event.chat_id):
            try:
                await event.delete()
            except Exception:
                pass

        try:
            language = await get_lang(event.chat_id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")

        if isinstance(event.message.from_id, PeerUser):
            upl = [
                [
                    Button.inline(
                        text=_["anon_admin"],
                        data="AnonymousAdmin",
                    ),
                ]
            ]
            return await event.reply(_["general_4"], buttons=upl)

        if event.sender_id not in SUDOERS:
            try:
                member, status = await tbot.get_chat_member(
                    event.chat_id, event.sender_id
                )

                if status not in [
                    "ADMIN",
                    "OWNER",
                ] or (
                    member.admin_rights is None or not member.admin_rights.manage_call
                ):
                    return await event.reply(_["general_5"])

            except Exception as e:
                return await event.reply(f"Error: {str(e)}")

        return await mystic(event, _)

    return wrapper


def actual_admin_cb(mystic):
    async def wrapper(event):
        try:
            language = await get_lang(event.chat_id)
            _ = get_string(language)
        except Exception:
            _ = get_string("en")

        if not await is_maintenance():
            if event.sender_id not in SUDOERS:
                return await event.answer(
                    _["maint_4"],
                    alert=True,
                )

        if event.is_private:
            return await mystic(event, _)

        is_non_admin = await is_nonadmin_chat(event.chat_id)
        if not is_non_admin:
            try:
                member, status = await tbot.get_chat_member(
                    event.chat_id,
                    event.sender_id,
                )

                if status not in [
                    "ADMIN",
                    "OWNER",
                ] or (
                    member.admin_rights is None or not member.admin_rights.manage_call
                ):
                    if event.sender_id not in SUDOERS:
                        token = await int_to_alpha(event.sender_id)
                        _check = await get_authuser_names(event.sender_id)
                        if token not in _check:
                            return await query.answer(
                                _["general_5"],
                                alert=True,
                            )

            except Exception as e:
                await event.client.handle_error(e)
                return await query.answer(f"Error: {str(e)}")

        return await mystic(event, _)

    return wrapper

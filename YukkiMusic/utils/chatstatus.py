from telethon.tl.types import ChannelParticipantsAdmins

from YukkiMusic import telethn
from YukkiMusic.misc import SUDOERS



async def user_is_admin(user_id: int, message):
    status = False
    if message.is_private:
        return True

    async for user in telethn.iter_participants(
        message.chat_id, filter=ChannelParticipantsAdmins
    ):
        if user_id == user.id or user_id in SUDOERS:
            status = True
            break
    return status


async def can_delete_messages(message):
    if message.is_private:
        return True
    elif message.chat.admin_rights:
        status = message.chat.admin_rights.delete_messages
        return status
    else:
        return False

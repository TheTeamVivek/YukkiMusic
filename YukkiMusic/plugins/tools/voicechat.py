from pyrogram import filters
from pyrogram.raw import base
from pyrogram.raw.functions.channels import GetFullChannel
from pyrogram.raw.functions.phone import GetGroupParticipants
from pyrogram.raw.types import InputGroupCall, InputPeerChat

from config import LOG_GROUP_ID
from YukkiMusic import app
from YukkiMusic.utils.database import get_assistant
from YukkiMusic.utils import Yukkibin


@app.on_message(
    filters.command(["vcuser", "vcusers", "vcmember", "vcmembers"]) & filters.admin
)
async def vc_members(client, message):
    msg = await message.reply_text("**Rᴀdhᴇ rᴀdhᴇ**\n\nᴩlᴇᴀsᴇ wᴀiᴛ.......")
    userbot = await get_assistant(message.chat.id)
    try:
        full_chat: base.messages.ChatFull = await userbot.invoke(
            GetFullChannel(channel=(await userbot.resolve_peer(message.chat.id)))
        )

        if not full_chat.full_chat.call:
            return await msg.edit(
                "**Rᴀdhᴇ rᴀdhᴇ**\n\nOᴏᴩs, iᴛ lᴏᴏᴋs liᴋᴇ Vᴏiᴄᴇ ᴄhᴀᴛ is ᴏff"
            )

        access_hash = full_chat.full_chat.call.access_hash
        ids = full_chat.full_chat.call.id
        input_group_call = InputGroupCall(id=ids, access_hash=access_hash)
        input_peer_chat = InputPeerChat(chat_id=message.chat.id)

        result = await userbot.invoke(
            GetGroupParticipants(
                call=input_group_call,
                ids=[input_peer_chat],
                offset="",
                sources=[],
                limit=1000,
            )
        )

        users = result.participants

        if not users:
            return await msg.edit(
                "**Rᴀdhᴇ rᴀdhᴇ**\n\nThᴇrᴇ ᴀrᴇ nᴏ ʍᴇʍʙᴇrs in ᴛhᴇ vᴏiᴄᴇ ᴄhᴀᴛ ᴄurrᴇnᴛly."
            )

        mg = "**Rᴀdhᴇ rᴀdhᴇ**\n\n"
        for user in users:
            title, username = None, None
            if hasattr(user.peer, "channel_id") and user.peer.channel_id:
                user_id = int(f"-100{user.peer.channel_id}")
                try:
                    chat = await userbot.get_chat(user_id)
                    title = chat.title
                    username = chat.username
                except Exception:
                    chats = result.chats
                    for c in chats:
                        if c.id == user.peer.channel_id:
                            title = c.title
            else:
                user_id = user.peer.user_id
                try:
                    user_info = await userbot.get_users(user_id)
                    title = user_info.mention
                    username = user_info.username
                except Exception:
                    for user_obj in result.users:
                        if user_obj.id == user_id:
                            username = user_obj.username or "No Username"
                            title = f"[{user_obj.first_name}](tg://user?id={user_id})"

            is_left = user.left
            just_joined = user.just_joined
            is_muted = bool(user.muted and not user.can_self_unmute)
            is_silent = bool(user.muted and user.can_self_unmute)

            mg += f"""**{'Tiᴛlᴇ' if hasattr(user.peer, 'channel_id') and user.peer.channel_id else 'Nᴀʍᴇ'} = {title}**
    **Id** : {user_id}"""
            if username:
                mg += f"\n    **Usᴇrnᴀʍᴇ** : {username}"

            mg += f"""
    **Is Lᴇfᴛᴇd Frᴏʍ Grᴏuᴩ** : {is_left}
    **Is Jusᴛ Jᴏinᴇd** : {just_joined}
    **Is Silᴇnᴛ** : {is_silent}
    **Is Muᴛᴇd By Adʍin** : {is_muted}\n\n"""

        if mg != "**Rᴀdhᴇ rᴀdhᴇ**\n\n":
            if len(mg) < 4000:
                await msg.edit(mg)
            else:
                link = await Yukkibin(mg)
                await msg.edit(
                    f"[Yᴏu ᴄᴀn ᴄhᴇᴄᴋ ᴀll dᴇᴛᴀils ᴏf vᴏiᴄᴇᴄhᴀᴛ ʍᴇʍᴇʙᴇrs frᴏʍ hᴇrᴇ]({link})",
                    disable_web_page_preview=True,
                )
        else:
            await msg.edit("**Rᴀdhᴇ rᴀdhᴇ**\nNᴏ Mᴇʍʙᴇrs fᴏund in vᴏiᴄᴇᴄhᴀᴛ")

    except Exception as e:
        await app.send_message(LOG_GROUP_ID, f"An Error occured in vcmembers.py {e} ")
        await msg.edit(f"**Rᴀdhᴇ rᴀdhᴇ**\nAn error occurred: {e}")

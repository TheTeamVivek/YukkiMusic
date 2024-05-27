from asyncio import sleep

from telethon import events
from telethon.errors import ChatAdminRequiredError, UserAdminInvalidError
from telethon.tl.functions.channels import EditBannedRequest
from telethon.tl.types import ChannelParticipantsAdmins, ChatBannedRights

from YukkiMusic import telethn
from YukkiMusic.misc import SUDOERS

BANNED_RIGHTS = ChatBannedRights(
    until_date=None,
    view_messages=True,
    send_messages=True,
    send_media=True,
    send_stickers=True,
    send_gifs=True,
    send_games=True,
    send_inline=True,
    embed_links=True,
)


UNBAN_RIGHTS = ChatBannedRights(
    until_date=None,
    send_messages=None,
    send_media=None,
    send_stickers=None,
    send_gifs=None,
    send_games=None,
    send_inline=None,
    embed_links=None,
)


async def is_administrator(user_id: int, message):
    admin = False
    async for user in telethn.iter_participants(
        message.chat_id, filter=ChannelParticipantsAdmins
    ):
        if user_id == user.id or user_id in SUDOERS:
            admin = True
            break
    return admin


@telethn.on(events.NewMessage(pattern="^[!/]zombies ?(.*)"))
async def rm_deletedacc(show):
    con = show.pattern_match.group(1).lower()
    del_u = 0
    del_status = "**ɢʀᴏᴜᴘ ɪs ɴᴇᴀᴛ ᴀɴᴅ ᴄʟᴇᴀɴ , 0 ᴢᴏᴍʙɪᴇ ᴀᴄᴄᴏᴜɴᴛ ғᴏᴜɴᴅ**"
    if con != "clean":
        kontol = await show.reply("sᴇᴀʀᴄʜɪɴɢ ғᴏʀ ᴅᴇʟᴇᴛᴇᴅ ᴀᴄᴄᴏɪɴᴛs...")
        async for user in show.client.iter_participants(show.chat_id):
            if user.deleted:
                del_u += 1
                await sleep(1)
        if del_u > 0:
            del_status = (
                f"**sᴇᴀʀᴄʜɪɴɢ...** `{del_u}` **ᴅᴇʟᴇᴛᴇᴅ ᴀᴄᴄᴏᴜɴᴛ / ᴢᴏᴍʙɪᴇs ɪɴ ᴛʜɪs ɢʀᴏᴜᴘ"
                "\nʀᴇᴍᴏᴠᴇ ᴀʟʟ ᴅᴇʟᴇᴛᴇᴅ ᴀᴄᴄᴏᴜɴᴛ ʙʏ ** `/zombies clean`"
            )
        return await kontol.edit(del_status)
    chat = await show.get_chat()
    admin = chat.admin_rights
    creator = chat.creator
    if not admin and not creator:
        return await show.reply("**sᴏʀʀʏ sɪʀ! ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀɴ ᴀᴅᴍɪᴍ ᴏғ ᴛʜᴇ ᴄʜᴀᴛ.**")
    memek = await show.reply("ʀᴇᴍᴏᴠɪɴɢ... ᴀʟʟ ᴅᴇʟᴇᴛᴇᴅ ᴀᴄᴄᴏᴜɴᴛ ғʀᴏᴍ ᴛʜɪs ɢʀᴏᴜᴘ")
    del_u = 0
    del_a = 0
    async for user in telethn.iter_participants(show.chat_id):
        if user.deleted:
            try:
                await show.client(
                    EditBannedRequest(show.chat_id, user.id, BANNED_RIGHTS)
                )
            except ChatAdminRequiredError:
                return await show.edit(
                    "sᴏʀʀʏ sɪʀ! ɪ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ʙᴀɴ ʀɪɢʜᴛs ɪɴ ᴛʜɪs ɢʀᴏᴜᴘ ᴛᴏ ᴘᴇʀғᴏʀᴍ ᴛʜɪs ᴀᴄᴛɪᴏɴ"
                )
            except UserAdminInvalidError:
                del_u -= 1
                del_a += 1
            await telethn(EditBannedRequest(show.chat_id, user.id, UNBAN_RIGHTS))
            del_u += 1
    if del_u > 0:
        del_status = f"**ʀᴇᴍᴏᴠᴇᴅ ** `{del_u}` **ᴢᴏᴍʙɪᴇs**"
    if del_a > 0:
        del_status = (
            f"**ʀᴇᴍᴏᴠᴇᴅ ** `{del_u}` **ᴢᴏᴍʙɪᴇs** "
            f"\n`{del_a}` **ᴀᴅᴍɪɴ ᴢᴏᴍʙɪᴇs ᴄᴀɴ'ᴛ ʙᴇ ᴅᴇʟᴇᴛᴇᴅ. ʀᴇᴍᴏᴠᴇ ɪᴛ ʙʏ ᴍᴀɴᴜᴀʟʟʏ**"
        )
    await memek.edit(del_status)

import asyncio

from pyrogram import filters
from pyrogram.enums import ChatMembersFilter
from pyrogram.errors import FloodWait

from YukkiMusic import app
from YukkiMusic.utils.database import get_assistant
from YukkiMusic.utils.filter import admin_filter

SPAM_CHATS = []


@app.on_message(
    filters.command(["all", "allmention", "mentionall", "tagall"], prefixes=["/", "@"])
    & admin_filter
)
async def tag_all_users(_, message):
    if message.chat.id in SPAM_CHATS:
        return await message.reply_text(
            "ᴛᴀɢɢɪɴɢ ᴘʀᴏᴄᴇss ɪs ᴀʟʀᴇᴀᴅʏ ʀᴜɴɴɪɴɢ ɪғ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ sᴛᴏᴘ sᴏ ᴜsᴇ /cancel"
        )
    replied = message.reply_to_message
    if len(message.command) < 2 and not replied:
        await message.reply_text(
            "** ɢɪᴠᴇ sᴏᴍᴇ ᴛᴇxᴛ ᴛᴏ ᴛᴀɢ ᴀʟʟ, ʟɪᴋᴇ »** `@all Hi Friends`"
        )
        return
    if replied:
        try:
            SPAM_CHATS.append(message.chat.id)
            usernum = 0
            usertxt = ""
            async for m in app.get_chat_members(message.chat.id):
                if message.chat.id not in SPAM_CHATS:
                    break
                usernum += 1
                usertxt += f"[{m.user.first_name}](tg://user?id={m.user.id})"
                if usernum == 14:
                    await app.send_message(
                        message.chat.id,
                        f"{replied.text}\n\n{usertxt}",
                        disable_web_page_preview=True,
                    )
                    await asyncio.sleep(1)
                    usernum = 0
                    usertxt = ""

        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass
    else:
        try:
            text = message.text.split(None, 1)[1]
            SPAM_CHATS.append(message.chat.id)
            usernum = 0
            usertxt = ""
            async for m in app.get_chat_members(message.chat.id):
                if message.chat.id not in SPAM_CHATS:
                    break
                usernum += 1
                usertxt += f"[{m.user.first_name}](tg://user?id={m.user.id})"
                if usernum == 14:
                    await app.send_message(
                        message.chat.id,
                        f"{text}\n{usertxt}",
                        disable_web_page_preview=True,
                    )
                    await asyncio.sleep(2)
                    usernum = 0
                    usertxt = ""
        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass


async def tag_all_admins(_, message):
    if message.chat.id in SPAM_CHATS:
        return await message.reply_text(
            "ᴛᴀɢɢɪɴɢ ᴘʀᴏᴄᴇss ɪs ᴀʟʀᴇᴀᴅʏ ʀᴜɴɴɪɴɢ ɪғ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ sᴛᴏᴘ sᴏ ᴜsᴇ /cancel"
        )
    replied = message.reply_to_message
    if len(message.command) < 2 and not replied:
        await message.reply_text(
            "** ɢɪᴠᴇ sᴏᴍᴇ ᴛᴇxᴛ ᴛᴏ ᴛᴀɢ ᴀʟʟ, ʟɪᴋᴇ »** `@admins Hi Friends`"
        )
        return
    if replied:
        try:
            SPAM_CHATS.append(message.chat.id)
            usernum = 0
            usertxt = ""
            async for m in app.get_chat_members(
                message.chat.id, filter=ChatMembersFilter.ADMINISTRATORS
            ):
                if message.chat.id not in SPAM_CHATS:
                    break
                usernum += 1
                usertxt += f"[{m.user.first_name}](tg://user?id={m.user.id})"
                if usernum == 14:
                    await app.send_message(
                        message.chat.id,
                        f"{replied.text}\n\n usertxt",
                        disable_web_page_preview=True,
                    )
                    await asyncio.sleep(1)
                    usernum = 0
                    usertxt = ""

        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass
    else:
        try:
            text = message.text.split(None, 1)[1]
            SPAM_CHATS.append(message.chat.id)
            usernum = 0
            usertxt = ""
            async for m in app.get_chat_members(message.chat.id):
                if message.chat.id not in SPAM_CHATS:
                    break
                usernum += 1
                usertxt += f"[{m.user.first_name}](tg://user?id={m.user.id})"
                if usernum == 14:
                    await app.send_message(
                        message.chat.id,
                        f"{text}\n{usertxt}",
                        disable_web_page_preview=True,
                    )
                    await asyncio.sleep(2)
                    usernum = 0
                    usertxt = ""
        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass


@app.on_message(
    filters.command(
        ["atag", "aall", "amention", "amentionall"], prefixes=["/", "@", ".", "#"]
    )
    & admin_filter
)
async def atag_all_useres(_, message):
    userbot = await get_assistant(message.chat.id)
    if message.chat.id in SPAM_CHATS:
        return await message.reply_text(
            "ᴛᴀɢɢɪɴɢ ᴘʀᴏᴄᴇss ɪs ᴀʟʀᴇᴀᴅʏ ʀᴜɴɴɪɴɢ ɪғ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ sᴛᴏᴘ sᴏ ᴜsᴇ /acancel"
        )
    replied = message.reply_to_message
    if len(message.command) < 2 and not replied:
        await message.reply_text(
            "** ɢɪᴠᴇ sᴏᴍᴇ ᴛᴇxᴛ ᴛᴏ ᴛᴀɢ ᴀʟʟ, ʟɪᴋᴇ »** `@aall Hi Friends`"
        )
        return
    if replied:
        SPAM_CHATS.append(message.chat.id)
        usernum = 0
        usertxt = ""
        async for m in app.get_chat_members(message.chat.id):
            if message.chat.id not in SPAM_CHATS:
                break
            usernum += 1
            usertxt += f"[{m.user.first_name}](tg://openmessage?user_id={m.user.id})"
            if usernum == 14:
                await userbot.send_message(
                    message.chat.id,
                    f"{replied.text}\n\n{usertxt}",
                    disable_web_page_preview=True,
                )
                await asyncio.sleep(2)
                usernum = 0
                usertxt = ""
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass
    else:
        text = message.text.split(None, 1)[1]

        SPAM_CHATS.append(message.chat.id)
        usernum = 0
        usertxt = ""
        async for m in app.get_chat_members(message.chat.id):
            if message.chat.id not in SPAM_CHATS:
                break
            usernum += 1
            usertxt += f'<a href="tg://openmessage?user_id={m.user.id}">{m.user.first_name}</a>'

            if usernum == 14:
                await userbot.send_message(
                    message.chat.id, f"{text}\n{usertxt}", disable_web_page_preview=True
                )
                await asyncio.sleep(2)
                usernum = 0
                usertxt = ""
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass


@app.on_message(
    filters.command(["admin", "admins"], prefixes=["/", "@"]) & filters.group
)
async def admintag_with_reporting(_, message):
    if message.from_user is None:
        return
    admins = []
    async for i in app.get_chat_members(
        chat_id=message.chat.id, filter=ChatMembersFilter.ADMINISTRATORS
    ):
        admins.append(i.user.id)

    if message.from_user.id in admins:
        return await tag_all_admins(_, message)
    if message.from_user.id not in admins:
        if not message.reply_to_message:
            return await message.reply_text("Reply to a message to report that user.")
        reply_id = reply.from_user.id
        reply = message.reply_to_message if message.reply_to_message else message
        linked_chat = (await app.get_chat(message.chat.id)).linked_chat
        if (
            reply_id in admins
            or reply_id == message.chat.id
            or reply_id == linked_chat.id
        ):
            return await message.reply_text(
                "Do you know that the user you are replying is an admin ?"
            )
        else:
            if reply_id in admins or reply_id == message.chat.id:
                return await message.reply_text(
                    "Do you know that the user you are replying is an admin ?"
                )
        user_mention = reply.from_user.mention
        text = f"Reported {user_mention} to admins!."
        admin_data = [
            i
            async for i in app.get_chat_members(
                chat_id=message.chat.id, filter=ChatMembersFilter.ADMINISTRATORS
            )
        ]
        for admin in admin_data:
            if admin.user.is_bot or admin.user.is_deleted:
                continue
            text += f"[\u2063](tg://user?id={admin.user.id})"

        await reply.reply_text(text)


@app.on_message(
    filters.command(
        [
            "stopmention",
            "cancel",
            "cancelmention",
            "offmention",
            "mentionoff",
            "cancelall",
        ],
        prefixes=["/", "@"],
    )
    & admin_filter
)
async def cancelcmd(_, message):
    chat_id = message.chat.id
    if chat_id in SPAM_CHATS:
        try:
            SPAM_CHATS.remove(chat_id)
        except Exception:
            pass
        return await message.reply_text("**ᴛᴀɢɢɪɴɢ ᴘʀᴏᴄᴇss sᴜᴄᴄᴇssғᴜʟʟʏ sᴛᴏᴘᴘᴇᴅ!**")

    else:
        await message.reply_text("**ɴᴏ ᴘʀᴏᴄᴇss ᴏɴɢᴏɪɴɢ!**")
        return


__MODULE__ = "Tᴀɢᴀʟʟ"
__HELP__ = """

@all ᴏʀ /all | /tagall ᴏʀ  @tagall | /mentionall ᴏʀ  @mentionall [ᴛᴇxᴛ] ᴏʀ [ʀᴇᴘʟʏ ᴛᴏ ᴀɴʏ ᴍᴇssᴀɢᴇ] ᴛᴏ ᴛᴀɢ ᴀʟʟ ᴜsᴇʀ's ɪɴ ʏᴏᴜʀ ɢʀᴏᴜᴘ ʙᴛ ʙᴏᴛ

@aall ᴏʀ /aall | /atagall ᴏʀ  @atagall | /amentionall ᴏʀ  @amentionall  [ᴛᴇxᴛ] ᴏʀ [ʀᴇᴘʟʏ ᴛᴏ ᴀɴʏ ᴍᴇssᴀɢᴇ] ᴛᴏ ᴛᴀɢ ᴀʟʟ ᴜsᴇʀ's ɪɴ ʏᴏᴜʀ ɢʀᴏᴜᴘ ʙʏ ʙᴏᴛ's ᴀssɪsᴛᴀɴᴛ


/admins Oʀ @admins [ᴛᴇxᴛ] ᴏʀ [ʀᴇᴘʟʏ ᴛᴏ ᴀɴʏ ᴍᴇssᴀɢᴇ] ᴛᴏ ᴛᴀɢ ᴀʟʟ ᴀᴅᴍɪɴ's ɪɴ ʏᴏᴜʀ ɢʀᴏᴜᴘ

/cancel Oʀ @cancel |  /offmention Oʀ @offmention | /mentionoff Oʀ @mentionoff | /cancelall Oʀ @cancelall - ᴛᴏ sᴛᴏᴘ ʀᴜɴɴɪɴɢ ᴀɴʏ ᴛᴀɢ ᴘʀᴏᴄᴇss

**__Nᴏᴛᴇ__** Tʜɪs ᴄᴏᴍᴍᴀɴᴅ ᴄᴀɴ ᴏɴʟʏ ᴜsᴇ ᴛʜᴇ Aᴅᴍɪɴs ᴏғ Cʜᴀᴛ ᴀɴᴅ ᴍᴀᴋᴇ Sᴜʀᴇ Bᴏᴛ ᴀɴᴅ ᴀssɪsᴛᴀɴᴛ ɪs ᴀɴ ᴀᴅᴍɪɴ ɪɴ ʏᴏᴜʀ ɢʀᴏᴜᴘ's
"""

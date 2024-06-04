from pyrogram import filters
from pyrogram.enums import ChatMembersFilter
from pyrogram.types import ChatJoinRequest
from pyrogram.errors.exceptions.bad_request_400 import UserAlreadyParticipant
from YukkiMusic import app
from YukkiMusic.core.mongo import mongodb
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.keyboard import ikb
from YukkiMusic.utils.permissions import adminsOnly, member_permissions

approvaldb = mongodb.autoapprove


@app.on_message(filters.command("autoapprove") & filters.group)
@adminsOnly("can_change_info")
async def approval_command(client, message):
    chat_id = message.chat.id
    chat = await approvaldb.find_one({"chat_id": chat_id})
    if chat:
        mode = chat.get("mode", "")
        if not mode:
            mode = "automatic"
            await approvaldb.update_one(
                {"chat_id": chat_id},
                {"$set": {"mode": mode}},
                upsert=True,
            )
        if mode == "automatic":
            switch = "manual"
        else:
            switch = "automatic"
        buttons = {
            "Turn OFF": "approval_off",
            f"{(mode.upper())}": f"approval_{switch}",
        }
        keyboard = ikb(buttons, 1)
        await message.reply(
            "**Autoapproval for this chat: Enabled.**", reply_markup=keyboard
        )
    else:
        buttons = {"Turn ON": "approval_on"}
        keyboard = ikb(buttons, 1)
        await message.reply(
            "**Autoapproval for this chat: Disabled.**", reply_markup=keyboard
        )


@app.on_callback_query(filters.regex("approval(.*)"))
async def approval_cb(client, cb):
    chat_id = cb.message.chat.id
    from_user = cb.from_user
    permissions = await member_permissions(chat_id, from_user.id)
    permission = "can_restrict_members"
    if permission not in permissions:
        if from_user.id not in SUDOERS:
            return await cb.answer(
                f"You don't have the required permission.\n Permission: {permission}",
                show_alert=True,
            )
    command_parts = cb.data.split("_", 1)
    option = command_parts[1]
    if option == "off":
        if await approvaldb.count_documents({"chat_id": chat_id}) > 0:
            approvaldb.delete_one({"chat_id": chat_id})
            buttons = {"Turn ON": "approval_on"}
            keyboard = ikb(buttons, 1)
            return await cb.edit_message_text(
                "**Autoapproval for this chat: Disabled.**",
                reply_markup=keyboard,
            )
    if option == "on":
        switch = "manual"
        mode = "automatic"
    if option == "automatic":
        switch = "manual"
        mode = option
    if option == "manual":
        switch = "automatic"
        mode = option
    await approvaldb.update_one(
        {"chat_id": chat_id},
        {"$set": {"mode": mode}},
        upsert=True,
    )
    chat = await approvaldb.find_one({"chat_id": chat_id})
    mode = chat["mode"].upper()
    buttons = {"Turn OFF": "approval_off", f"{mode}": f"approval_{switch}"}
    keyboard = ikb(buttons, 1)
    await cb.edit_message_text(
        "**Autoapproval for this chat: Enabled.**", reply_markup=keyboard
    )


@app.on_message(filters.command("clear_pending") & filters.group)
@adminsOnly("can_restrict_members")
async def clear_pending_command(client, message):
    chat_id = message.chat.id
    result = await approvaldb.update_one(
        {"chat_id": chat_id},
        {"$set": {"pending_users": []}},
    )
    if result.modified_count > 0:
        await message.reply_text("Cleared pending users.")
    else:
        await message.reply_text("No pending users to clear.")


@app.on_chat_join_request(filters.group)
async def accept(client, message: ChatJoinRequest):
    chat = message.chat
    user = message.from_user
    chat_id = await approvaldb.find_one({"chat_id": chat.id})
    if chat_id:
        mode = chat_id["mode"]
        if mode == "automatic":
            await app.approve_chat_join_request(chat_id=chat.id, user_id=user.id)
            return
        if mode == "manual":
            is_user_in_pending = await approvaldb.count_documents(
                {"chat_id": chat.id, "pending_users": int(user.id)}
            )
            if is_user_in_pending == 0:
                await approvaldb.update_one(
                    {"chat_id": chat.id},
                    {"$addToSet": {"pending_users": int(user.id)}},
                    upsert=True,
                )
                buttons = {
                    "accept": f"manual_approve_{user.id}",
                    "Decline": f"manual_decline_{user.id}",
                }
                keyboard = ikb(buttons, int(2))
                text = f"**User: {user.mention} has send a request to join our  group. Any admins can accept or decline it.**"
                admin_data = [
                    i
                    async for i in app.get_chat_members(
                        chat_id=message.chat.id,
                        filter=ChatMembersFilter.ADMINISTRATORS,
                    )
                ]
                for admin in admin_data:
                    if admin.user.is_bot or admin.user.is_deleted:
                        continue
                    text += f"[\u2063](tg://user?id={admin.user.id})"
                return await app.send_message(chat.id, text, reply_markup=keyboard)


@app.on_callback_query(filters.regex("manual_(.*)"))
async def manual(app, cb):
    chat = cb.message.chat
    from_user = cb.from_user
    permissions = await member_permissions(chat.id, from_user.id)
    permission = "can_restrict_members"
    if permission not in permissions:
        if from_user.id not in SUDOERS:
            return await cb.answer(
                f"You don't have the required permission.\n Permission: {permission}",
                show_alert=True,
            )
    datas = cb.data.split("_", 2)
    dis = datas[1]
    id = datas[2]
    if dis == "approve":
        try:
            await app.approve_chat_join_request(chat_id=chat.id, user_id=id)
        except UserAlreadyParticipant:
            await cb.answer(
                f"User Is Approved in Your Group By AnyOne",
                show_alert=True,
            )
            return await cb.message.delete()
    
    if dis == "decline":
        await app.decline_chat_join_request(chat_id=chat.id, user_id=id)
    await approvaldb.update_one(
        {"chat_id": chat.id},
        {"$pull": {"pending_users": int(id)}},
    )
    return await cb.message.delete()


__MODULE__ = "Aᴘᴘʀᴏᴠᴇ"
__HELP__ = """
command: /autoapprove

Tʜɪs ᴍᴏᴅᴜʟᴇ ʜᴇʟᴘs ᴛᴏ ᴀᴜᴛᴏᴍᴀᴛɪᴄᴀʟʟʏ ᴀᴄᴄᴇᴘᴛ ᴄʜᴀᴛ ɪᴏɪɴ ʀᴇǫᴜᴇsᴛ sᴇɴᴅ ʙʏ ᴀ ᴜsᴇʀ ᴛʜʀᴏᴜɢʜ ɪɴᴠɪᴛᴀᴛɪᴏɴ ʟɪɴᴋ ᴏғ ʏᴏᴜʀ ɢʀᴏᴜᴘ

**Mᴏᴅᴇs:**
ᴡʜᴇɴ ʏᴏᴜ sᴇɴᴅ /autoapprove ɪɴ ʏᴏᴜʀ ɢʀᴏᴜᴘ ʏᴏᴜ sᴇᴇ ᴛᴜʀɴ ᴏɴ ʙᴜᴛᴛᴏɴ ɪғ ᴀᴜᴛᴛᴏᴘʀᴏᴠᴇ ɴᴏᴛ ᴇɴᴀʙʟᴇᴅ ғᴏʀ ʏᴏᴜʀ ᴄʜᴀᴛ ɪғ ᴀʟʀᴇᴅʏ ᴛᴜʀɴᴇᴅ ᴏɴ ʏᴏᴜ ᴡɪʟʟ sᴇ ᴛᴡᴏ ᴍᴏᴅᴇs ᴛʜᴀᴛ's ᴀʀᴇ ʙᴇʟᴏᴡ ᴀɴᴅ ʜɪs ᴜsᴀsɢᴇ


¤ Automatic - ᴀᴜᴛᴏᴍᴀᴛɪᴄᴀʟʟʏ ᴀᴄᴄᴇᴘᴛs ᴄʜᴀᴛ ᴊᴏɪɴ ʀᴇǫᴜᴇsᴛ.

¤ Manual - ᴀ ᴍᴇssᴀɢᴇ ᴡɪʟʟ ʙᴇ sᴇɴᴅ ᴛᴏ ᴛʜᴇ ᴄʜᴀᴛ ʙʏ ᴛᴀɢɢɪɴɢ ᴛʜᴇ ᴀᴅᴍɪɴs. ᴛʜᴇ ᴀᴅᴍɪɴs ᴄᴀɴ ᴀᴄᴄᴇᴘᴛ ᴏʀ ᴅᴇᴄʟɪɴᴇ ᴛʜᴇ ʀᴇǫᴜᴇsᴛs.

Usᴇ: /clearpending ᴄᴏᴍᴍᴀɴᴅ ᴛᴏ ʀᴇᴍᴏᴠᴇ ᴀʟʟ ᴘᴇɴᴅɪɴɢ ᴜsᴇʀ ɪᴅ ғʀᴏᴍ ᴅʙ. ᴛʜɪs ᴡɪʟʟ ᴀʟʟᴏᴡ ᴛʜᴇ ᴜsᴇʀ ᴛᴏ sᴇɴᴅ ʀᴇǫᴜᴇsᴛ ᴀɢᴀɪɴ.
"""

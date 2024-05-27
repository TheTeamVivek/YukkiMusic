from pyrogram import filters
from pyrogram.errors import ChatAdminRequired, InviteRequestSent, UserAlreadyParticipant

from YukkiMusic import app
from YukkiMusic.utils.database import get_assistant


@app.on_message(filters.command("join") & filters.group)
async def invite_assistant(client, message):
    try:
        # Get the music bot assistant
        userbot = await get_assistant(message.chat.id)

        # Check if the bot has admin rights in the group
        try:
            await client.get_chat_member(message.chat.id, "me")
        except ChatAdminRequired:
            return await message.reply_text(
                "I don't have permission to invite please give me rights"
            )

        # Unban the assistant if it's banned in the group
        try:
            await client.unban_chat_member(message.chat.id, userbot.id)
        except:
            pass

        # Get the invite link for the group
        invitelink = await client.export_chat_invite_link(message.chat.id)

        # Invite the assistant to the group
        await userbot.join_chat(invitelink)

        await message.reply_text("Assistant successfully invited to the group !")

    except InviteRequestSent:
        await message.reply_text("Invite request already sent.")

    except UserAlreadyParticipant:
        await message.reply_text("Assistant is already a participant in the group.")

    except Exception as e:
        await message.reply_text(f"An error occurred: {e}")

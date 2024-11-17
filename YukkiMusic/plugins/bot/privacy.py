from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message
import json
import os
from datetime import datetime
import config
from strings import command
from YukkiMusic import app
from config import BANNED_USERS
from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import (
    get_playlist_names,
    get_playlist,
    delete_playlist,
    get_userss,
    is_banned_user,
    authuserdb,
    delete_served_user,
    remove_sudo,
)


TEXT = f"""
🔒 **Privacy Policy for {app.mention} !**

Your privacy is important to us. To learn more about how we collect, use, and protect your data, please review our Privacy Policy here: [Privacy Policy](https://github.com/TheTeamVivek/YukkiMusic/blob/master/PRIVACY.md).

If you have any questions or concerns, feel free to reach out to our [Support Team]({config.SUPPORT_GROUP}).
"""


PRIVACY_SECTIONS = {
    "collect": """
**What Information We Collect**

• Basic Telegram user data (ID, username)
• Chat/Group IDs where the bot is used
• Command usage and interactions
• Playlists and music preferences
• Voice chat participation data
• User settings and configurations
""",
    "why": """
**Why We Collect It**

• To provide music streaming services
• To maintain user playlists
• To process voice chat requests
• To manage user permissions
• To improve bot features
• To prevent abuse and spam
""",
    "do": """
**What We Do**

• Store data securely in encrypted databases
• Process music requests and streams
• Maintain user preferences
• Monitor for proper functionality
• Delete temporary files after use
• Implement security measures
""",
    "donot": """
**What We Don't Do**

• Share your data with third parties
• Store unnecessary personal information
• Keep data longer than needed
• Use data for marketing
• Track users across platforms
• Sell any user information
""",
    "rights": """
**Your Rights**

• Access your stored data
• Request data deletion
• Modify your settings
• Opt-out of data collection
• Report privacy concerns
• Contact support for help
""",
}


async def find_chat_ids_by_auth_user_id(auth_user_id):
    chat_ids = []
    async for document in authuserdb.find():
        for note_key, note_data in document.get("notes", {}).items():
            if note_data.get("auth_user_id") == auth_user_id:
                chat_ids.append(document["chat_id"])
    return chat_ids


async def delete_auth_user_data(auth_user_id):
    async for document in authuserdb.find():
        chat_id = document["chat_id"]
        notes = document.get("notes", {})
        keys_to_remove = [
            key
            for key, value in notes.items()
            if value.get("auth_user_id") == auth_user_id
        ]
        for key in keys_to_remove:
            notes.pop(key)
        if keys_to_remove:
            await authuserdb.update_one(
                {"chat_id": chat_id}, {"$set": {"notes": notes}}
            )


@app.on_message(command("PRIVACY_COMMAND") & ~BANNED_USERS)
async def privacy_menu(client, message: Message):
    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    "Privacy Policy", callback_data="show_privacy_sections"
                )
            ],
            [
                InlineKeyboardButton("Retrieve Data", callback_data="retrieve_data"),
                InlineKeyboardButton("Delete Data", callback_data="delete_data"),
            ],
            [InlineKeyboardButton("Close", callback_data="close")],
        ]
    )
    await message.reply_text(TEXT, reply_markup=keyboard, disable_web_page_preview=True)


@app.on_callback_query(filters.regex("show_privacy_sections") & ~BANNED_USERS)
async def show_privacy_sections(client, callback_query):
    """Show detailed privacy policy sections"""
    keyboard = InlineKeyboardMarkup(
        [
            [InlineKeyboardButton("What We Collect", callback_data="privacy_collect")],
            [InlineKeyboardButton("Why We Collect", callback_data="privacy_why")],
            [InlineKeyboardButton("What We Do", callback_data="privacy_do")],
            [InlineKeyboardButton("What We Don't Do", callback_data="privacy_donot")],
            [InlineKeyboardButton("Your Rights", callback_data="privacy_rights")],
            [
                InlineKeyboardButton("Back", callback_data="privacy_back"),
                InlineKeyboardButton("Close", callback_data="close"),
            ],
        ]
    )
    await callback_query.edit_message_text(
        f"{TEXT}\n\nSelect a section to learn more:",
        reply_markup=keyboard,
        disable_web_page_preview=True,
    )


@app.on_callback_query(filters.regex("privacy_") & ~BANNED_USERS)
async def privacy_section_callback(client, callback_query):
    """Handle privacy section callbacks"""
    section = callback_query.data.split("_")[1]

    if section == "back":
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Privacy Policy", callback_data="show_privacy_sections"
                    )
                ],
                [
                    InlineKeyboardButton(
                        "Retrieve Data", callback_data="retrieve_data"
                    ),
                    InlineKeyboardButton("Delete Data", callback_data="delete_data"),
                ],
                [InlineKeyboardButton("Close", callback_data="close")],
            ]
        )
        return await callback_query.edit_message_text(
            TEXT, reply_markup=keyboard, disable_web_page_preview=True
        )

    if section in PRIVACY_SECTIONS:
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("Back", callback_data="show_privacy_sections"),
                    InlineKeyboardButton("Close", callback_data="close"),
                ],
            ]
        )
        await callback_query.edit_message_text(
            PRIVACY_SECTIONS[section], reply_markup=keyboard
        )


@app.on_callback_query(filters.regex("retrieve_data"))
async def export_user_data(_, cq):
    m = await cq.message.edit("Please wait..")
    user_id = cq.from_user.id
    user_data = {
        "user_id": user_id,
        "export_date": datetime.now().isoformat(),
        "basic_info": {
            "user_id": user_id,
            "username": cq.from_user.username,
            "first_name": cq.from_user.first_name,
            "last_name": cq.from_user.last_name,
        },
        "playlists": {},
        "authed_in": await find_chat_ids_by_auth_user_id(user_id),
        "ban_status": await is_banned_user(user_id),
        "sudo_status": user_id in SUDOERS,
        "user_stats": await get_userss(user_id),
    }
    try:
        playlist_names = await get_playlist_names(user_id)
        for name in playlist_names:
            playlist = await get_playlist(user_id, name)
            if playlist:
                user_data["playlists"][name] = playlist
    except Exception as e:
        pass
    user_data = {
        k: (
            {sk: sv for sk, sv in v.items() if sv is not None}
            if isinstance(v, dict)
            else v
        )
        for k, v in user_data.items()
        if v is not None
    }

    file_path = os.path.join("cache", f"user_data_{user_id}.json")

    try:
        with open(file_path, "w", encoding="utf-8") as f:
            json.dump(user_data, f, indent=4, ensure_ascii=False)
    except Exception as e:
        return await m.edit(
            f"Error occurred while creating data file: {str(e)}", show_alert=True
        )

    try:
        await cq.message.reply_document(
            document=file_path,
            caption=(
                "🔒 Here is your user data export from YukkiMusic.\n\n"
                "⚠️ This file contains your personal information. "
                "Please handle it carefully and do not share it with others.\n\n"
                "📊 Includes:\n"
                "- Personal Information\n"
                "- Playlists\n"
                "- Usage Statistics\n"
                "- Authorization Status\n"
                "- Ban Status\n"
                "- Sudo Privileges\n"
            ),
            file_name=f"data_{user_id}_.json",
        )
    except Exception as e:
        await m.edit(
            f"Error occurred while creating data file: {str(e)}", show_alert=True
        )
    finally:
        try:
            await cq.message.delete()
            os.remove(file_path)
        except Exception:
            pass


@app.on_callback_query(filters.regex("delete_data"))
async def retrieve_data(_, cq):
    await cq.message.edit(
        "Are you sure you want to delete your data?",
        reply_markup=InlineKeyboardMarkup(
            [
                [InlineKeyboardButton("Yes", callback_data="confirm_delete_data")],
                [InlineKeyboardButton("No", callback_data="privacy_back")],
            ]
        ),
    )


@app.on_callback_query(filters.regex("confirm_delete_data"))
async def delete_user_data(_, cq):
    await cq.answer("Please wait...", show_alert=True)

    user_id = cq.from_user.id

    try:
        _playlist = await get_playlist_names(user_id)
        for x in _playlist:
            await delete_playlist(user_id, x)
    except Exception:
        pass

    await delete_auth_user_data(user_id)
    await delete_served_user(user_id)

    if user_id in SUDOERS:
        SUDOERS.remove(user_id)
    try:
        await remove_sudo(user_id)
    except Exception:
        pass

    await delete_userss(user_id)
    await cq.edit_message_text("Your data has been deleted from the bot.")

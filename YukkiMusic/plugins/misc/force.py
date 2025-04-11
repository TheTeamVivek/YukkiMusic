# YukkiMusic/plugins/misc/force.py

import os
import re # لاستخدام التعبيرات النمطية للمطابقة
import yaml
import traceback
from pyrogram import Client, filters
from pyrogram.types import InlineKeyboardMarkup, InlineKeyboardButton, Message
from pyrogram.errors import ChatAdminRequired, UserNotParticipant, ChatWriteForbidden
from pyrogram.enums import ChatType

from YukkiMusic import app
from config import MUST_JOIN, START_IMG_URL, BANNED_USERS, OWNER_ID

from strings import get_command, get_string, languages_present
from YukkiMusic.misc import SUDOERS

from YukkiMusic.utils.database import get_lang



PROTECTED_COMMAND_KEYS = [

    "PING_COMMAND",

    "SETTINGS_COMMAND",
    "RELOAD_COMMAND", # إداري
    "GSTATS_COMMAND",
    "STATS_COMMAND",
    "LANGUAGE_COMMAND",

    # Play Commands
    "PLAY_COMMAND",
    "PLAYMODE_COMMAND",
    "CHANNELPLAY_COMMAND",
    "STREAM_COMMAND",

    #Playlists Command
    "PLAYLIST_COMMAND",
    "DELETE_PLAYLIST_COMMAND",
    "PLAY_PLAYLIST_COMMAND",
    "ADD_PLAYLIST_COMMAND", # قد يكون أمر خاص

    # Tools
    "QUEUE_COMMAND",
    "SONG_COMMAND",
    "LYRICS_COMMAND",

    # Admin Commands
    "AUTH_COMMAND",
    "UNAUTH_COMMAND",
    "AUTHUSERS_COMMAND",
    "PAUSE_COMMAND",
    "RESUME_COMMAND",
    "MUTE_COMMAND",
    "UNMUTE_COMMAND",
    "STOP_COMMAND",
    "SKIP_COMMAND",
    "SHUFFLE_COMMAND",
    "LOOP_COMMAND",
    "SEEK_COMMAND",
    "SEEK_BACK_COMMAND",
    "REBOOT_COMMAND", # إداري

    # Sudo Commands (عادة محمية بالفعل بصلاحيات السودو)
    "ADDSUDO_COMMAND",
    "DELSUDO_COMMAND",
    "SUDOUSERS_COMMAND",
    "BROADCAST_COMMAND",
    "BLACKLISTCHAT_COMMAND",
    "WHITELISTCHAT_COMMAND",
    "BLACKLISTEDCHAT_COMMAND",
    "VIDEOLIMIT_COMMAND",
    "VIDEOMODE_COMMAND",
    "MAINTENANCE_COMMAND",
    "LOGGER_COMMAND",
    "GETLOG_COMMAND",
    "GETVAR_COMMAND",
    "DELVAR_COMMAND",
    "SETVAR_COMMAND",
    "USAGE_COMMAND",
    "UPDATE_COMMAND",
    "RESTART_COMMAND",
    "AUTOEND_COMMAND",
    "AUTHORIZE_COMMAND",
    "UNAUTHORIZE_COMMAND",
    "AUTHORIZED_COMMAND",
    "BLOCK_COMMAND",
    "UNBLOCK_COMMAND",
    "BLOCKED_COMMAND",
    "SPEEDTEST_COMMAND",
    "ACTIVEVC_COMMAND",
    "ACTIVEVIDEO_COMMAND",
    "AC_COMMAND",
    "GBAN_COMMAND",
    "UNGBAN_COMMAND",
    "GBANNED_COMMAND",
]




async def is_protected_command(client: Client, message: Message) -> bool:
    text = message.text or message.caption
    if not text:
        return False


    lang_code = await get_lang(message.chat.id)
    if not lang_code or lang_code not in languages_present:
        lang_code = "en" 


    try:
        current_lang_cmds = get_command(lang_code)
        en_cmds = get_command("en")
    except Exception as e:
         print(f"[ERROR] force.py: Failed to get commands for lang '{lang_code}' or 'en'. Error: {e}")
         return False 

    commands_to_check = set() 


    for key in PROTECTED_COMMAND_KEYS:

        cmd_values = current_lang_cmds.get(key)
        if cmd_values:
            if isinstance(cmd_values, str):
                commands_to_check.add(cmd_values.lower())
            elif isinstance(cmd_values, list):
                 for val in cmd_values:
                      if isinstance(val, str):
                            commands_to_check.add(val.lower())

        
        cmd_values_en = en_cmds.get(key)
        if cmd_values_en:
            if isinstance(cmd_values_en, str):
                commands_to_check.add(cmd_values_en.lower())
            elif isinstance(cmd_values_en, list):
                for val in cmd_values_en:
                     if isinstance(val, str):
                            commands_to_check.add(val.lower())

    if not commands_to_check:
         print("[WARNING] force.py: No protected commands found to check against.")
         return False

    
    
    prefixes = ["/", "!", "%", ",", ".", "@", "#"]
    escaped_prefixes = [re.escape(p) for p in prefixes]
    bot_username = client.me.username or ""
    
    
    command_pattern = "|".join(re.escape(cmd) for cmd in commands_to_check) 
    full_pattern = re.compile(
        rf"^(?:{'|'.join(escaped_prefixes)})?({command_pattern})(?:@?{re.escape(bot_username)})?(?:\s|$)",
        re.IGNORECASE # تجاهل حالة الأحرف
    )

    match = full_pattern.match(text)
    if match:
        
        return True 

    return False 





@app.on_message(
    (filters.group | filters.private) 
    & ~BANNED_USERS 
    & filters.text 
    & ~filters.via_bot 
    & ~filters.forwarded, 
    group=-2 
)
async def check_subscription_before_command(client: Client, message: Message):
    
    is_protected = await is_protected_command(client, message)
    if not is_protected:
        
        return

    

    if not MUST_JOIN:
        
        return

    if not message.from_user:
        
        return

    user_id = message.from_user.id

    
    if user_id in OWNER_ID or user_id in SUDOERS:
        
        return

    try:
        try:
            
            await client.get_chat_member(MUST_JOIN, user_id)
            
            return
        except UserNotParticipant:
            
            link = None
            channel_name = MUST_JOIN 

            if MUST_JOIN.startswith("@"):
                 link = "https://t.me/" + MUST_JOIN.replace("@", "")
                 channel_name = MUST_JOIN
            elif MUST_JOIN.lstrip("-").isdigit():
                try:
                    chat_info = await client.get_chat(int(MUST_JOIN))
                    channel_name = chat_info.title
                    link = chat_info.invite_link
                    if not link:
                        print(f"[CRITICAL ERROR] force.py: Failed to get invite link for channel ID: {MUST_JOIN}. Check bot permissions (Invite Users via Link).")
                        await message.stop_propagation()
                        return
                except Exception as e:
                    print(f"[CRITICAL ERROR] force.py: Failed to get chat info for channel ID: {MUST_JOIN}. Error: {e}")
                    await message.stop_propagation()
                    return
            else:
                 print(f"[CRITICAL ERROR] force.py: Invalid MUST_JOIN value: {MUST_JOIN}")
                 await message.stop_propagation()
                 return

            
            join_text = (
                f"عذراً {message.from_user.mention} 👋🏻\n\n"
                f"لاستخدام أوامر البوت، يجب عليك الانضمام إلى [قناتنا]({link}) أولاً.\n"
                f"القناة: **{channel_name}**\n\n"
                "اضغط على الزر أدناه للانضمام ثم حاول مجدداً."
            )
            join_button = InlineKeyboardMarkup([
                [InlineKeyboardButton("✨ انضم إلى القناة ✨", url=link)]
            ])

            try:
                
                if message.chat.type == ChatType.PRIVATE and START_IMG_URL:
                     await message.reply_photo(
                        photo=START_IMG_URL,
                        caption=join_text,
                        reply_markup=join_button,
                        quote=False
                    )
                else:
                    await message.reply_text(
                        text=join_text,
                        reply_markup=join_button,
                        disable_web_page_preview=True,
                        quote=message.chat.type != ChatType.PRIVATE
                    )

                
                try:
                    if message.chat.type != ChatType.PRIVATE:
                        await message.delete()
                except Exception:
                    pass

            except ChatWriteForbidden:
                print(f"[WARNING] force.py: Cannot write in chat: {message.chat.id}")
                pass
            except Exception as e:
                print(f"[ERROR] force.py: Failed to send force-subscribe message to user {user_id} in chat {message.chat.id}: {e}")

            
            await message.stop_propagation()

    except ChatAdminRequired:
        print(f"⚠️ [CRITICAL ERROR] force.py: Bot must be an admin in the MUST_JOIN channel ({MUST_JOIN}) to check membership.")
        await message.stop_propagation() 
    except Exception as e:
        print(f"[ERROR] force.py: Unexpected error during subscription check for user {user_id} in chat {message.chat.id}:")
        traceback.print_exc()
        await message.stop_propagation() 

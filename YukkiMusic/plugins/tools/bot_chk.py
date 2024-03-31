import os
import re
import pytz
import asyncio
import datetime

from pyrogram import Client, filters
from pyrogram.errors import FloodWait
from YukkiMusic.utils.database import get_client
TIME_ZONE = "Asia/Kolkata"
BOT_LIST = ["YukkiMusic_vkBot","TprinceMusicBot"]  # 
CHANNEL_ID = -1002113072448
MESSAGE_ID = 10 
BOT_ADMIN_IDS = ["6815918609"]
GRP_ID = -1002080548793

async def main_devchecker():
    while True:
        print("Checking...")
        xxx_teletips = f"<u>**🏷 ᴡᴇʟᴄᴏᴍᴇ ᴛᴏ {(await app.get_chat(CHANNEL_ID)).title} ɪɴғᴏʀᴍᴀᴛɪᴏɴ ᴄʜᴀɴɴᴇʟ**</u>\n\n 📈 | <u>**ʀᴇᴀʟ ᴛɪᴍᴇ ʙᴏᴛ's sᴛᴀᴛᴜs 🍂**</u>"
        for bot in BOT_LIST:
            await asyncio.sleep(7)
            try:
                app = get_client(1)
                bot_info = await app.get_users(bot)
            except Exception:
                bot_info = bot

            try:
                yyy_teletips = await app.send_message(bot, "/start")
                aaa = yyy_teletips.id
                await asyncio.sleep(7)
                zzz_teletips = app.get_chat_history(bot, limit=1)
                async for ccc in zzz_teletips:
                    bbb = ccc.id
                if aaa == bbb:
                    xxx_teletips += f"\n\n╭⎋ **[{bot_info.first_name}](tg://user?id={bot_info.id})**\n╰⊚ **sᴛᴀᴛᴜs: ᴏғғʟɪɴᴇ ❌**"
                    for bot_admin_id in BOT_ADMIN_IDS:
                        try:
                            await app.send_message(int(GRP_ID), f"@admins\n **ᴋʏᴀ ᴋᴀʀ ʀᴀʜᴀ ʜᴀɪ 😡\n[{bot_info.first_name}](tg://user?id={bot_info.id}) ᴏғғ ʜᴀɪ. ᴀᴄᴄʜᴀ ʜᴜᴀ ᴅᴇᴋʜ ʟɪʏᴀ ᴍᴀɪɴᴇ.**")
                        except Exception:
                            pass
                    await app.read_chat_history(bot)
                else:
                    xxx_teletips += f"\n\n╭⎋ **[{bot_info.first_name}](tg://user?id={bot_info.id})**\n╰⊚ **sᴛᴀᴛᴜs: ᴏɴʟɪɴᴇ ✅**"
                    await app.read_chat_history(bot)
            except FloodWait as e:
                ttm = re.findall("\d{0,5}", str(e))
                await asyncio.sleep(int(ttm))
        time = datetime.datetime.now(pytz.timezone(f"{TIME_ZONE}"))
        last_update = time.strftime(f"%d %b %Y at %I:%M %p")
        chnk = await app.get_chat(CHANNEL_ID).title
        xxx_teletips += f"\n\n✔️ <u>ʟᴀsᴛ ᴄʜᴇᴄᴋᴇᴅ ᴏɴ:</u>\n**ᴅᴀᴛᴇ & ᴛɪᴍᴇ: {last_update}**\n**ᴛɪᴍᴇ ᴢᴏɴᴇ: ({TIME_ZONE})**\n\n<i><u>♻️ ʀᴇғʀᴇsʜᴇs ᴀᴜᴛᴏᴍᴀᴛɪᴄᴀʟʟʏ ᴡɪᴛʜɪɴ 30 ᴍɪɴᴜᴛᴇs.</u></i>\n\n<b>**๏ ᴘᴏᴡᴇʀᴇᴅ ʙʏ @{chnk} ๏**</b>"
        await app.edit_message_text(int(CHANNEL_ID), MESSAGE_ID, xxx_teletips)
        print(f"Last checked on: {last_update}")                
        await asyncio.sleep(1800)

asyncio.create_task(main_devchecker())
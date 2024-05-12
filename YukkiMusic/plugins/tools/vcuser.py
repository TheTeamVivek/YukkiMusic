from pytgcalls.exceptions import GroupCallNotFound
from pyrogram import filters
from YukkiMusic.core.call import Yukki
from YukkiMusic app

@app.on_message(filters.command("vcuser"))
async def get_vc_users(client, message):
    try:
        AB = await Yukki.get_participant( message.chat.id)
    except GroupCallNotFound:
        return await message.reply_text("Assisitant iss not in vc")
    users_info = ""
    for participant in AB:
        user_id = participant.user_id
        user = await app.get_users(user_id)
        users_info += f"[{user.first_name}](tg://user?id={user_id})\n"
    
    await message.reply(users_info)
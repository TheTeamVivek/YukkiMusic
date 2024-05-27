from pyrogram import filters

from YukkiMusic import app


@app.on_message(filters.command("nak"))
async def clean(_, message):
    a = await app.ask(
        chat_id=message.chat.id,
        text="your name??",
        user_id=message.from_user.id,
        timeout=10,
    )
    await message.reply_text(f"your name is {a}")

from pyrogram import filters
from pyrogram.enums import ChatAction
from googlesearch import search
from YukkiMusic import app


@app.on_message(
    filters.command(["shv"], prefixes=["+", ".", "/", "-", "?", "$", "#", "&"])
)
async def google(bot, message):
    if len(message.command) < 2 and not message.reply_to_message:
        await message.reply_text("Example:\n\n`/google lord ram`")
        return

    if message.reply_to_message and message.reply_to_message.text:
        user_input = message.reply_to_message.text
    else:
        user_input = " ".join(message.command[1:])

    a = search(user_input, advanced=True)
    txt = f"Search Query: {user_input}\n\nresults"
    for result in a:
        txt += f"\n\n[❍ {result.title}]({result.url})\n<b>{result.description}</b>"
    await message.reply_text(
        txt,
        disable_web_page_preview=True,
    )

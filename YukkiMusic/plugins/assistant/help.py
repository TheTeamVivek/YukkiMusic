from pyrogram import filters

from YukkiMusic import app


@app.on_message(filters.command(["help"], prefixes=["."]))
async def inline_help_menu(client, message):
    bot_results = await app.get_inline_bot_results(f"@{app.username}", "help_menu")
    await app.send_inline_bot_result(
        chat_id=message.chat.id,
        query_id=bot_results.query_id,
        result_id=bot_results.results[0].id,
    )
    try:
        await message.delete()
    except:
        pass

from pyrogram import filters
from SafoneAPI import SafoneAPI

from YukkiMusic import app

api = SafoneAPI()


@app.on_message(filters.command(["gen", "ccgen"], [".", "!", "/"]))
async def gen_cc(client, message):
    if len(message.command) < 2:
        return await message.reply_text("ɢɪᴠᴇ ᴍᴇ ᴀ ʙɪɴ ᴛᴏ ɢᴇɴᴇᴇʀᴀᴛᴇ ᴄᴄ")

    try:
        await message.delete()
    except:
        pass

    aux = await message.reply_text("ɢᴇɴᴇʀᴀᴛɪɴɢ....")
    bin = message.text.split(None, 1)[1]

    if len(bin) < 6:
        return await aux.edit("ɢɪᴠᴇ ᴍᴇ ᴀ ʙɪɴ ᴏғ 6 ᴅɪɢɪᴛ")

    try:
        resp = await api.ccgen(bin, 10)
        cards = resp.liveCC

        await aux.edit(
            f"""
⦿ Bɪɴ: `{resp.results[0].bin}`
⦿ Tɪᴍᴇ Tᴏᴏᴋ: {resp.took}\n\n
➤ Sᴏᴍᴇ Lɪᴠᴇ Gᴇɴᴇʀᴀᴛᴇᴅ Cᴄ ➻

╭✠╼━━━━━━❖━━━━━━━✠╮ 

{cards[0]}\n{cards[1]}\n{cards[2]}
{cards[3]}\n{cards[4]}\n{cards[5]}
{cards[6]}\n{cards[7]}\n{cards[8]}
{cards[9]}\n
╰✠╼━━━━━━❖━━━━━━━✠╯
""",
        )

    except Exception as e:
        return await aux.edit(f"Eʀʀᴏʀ: {e}.")

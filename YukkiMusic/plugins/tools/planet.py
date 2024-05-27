import requests
from pyrogram import filters

from YukkiMusic import app

API_KEY = "Zg2rDIcq415hV5BMnHJTaQ==WO9Xuk9P89Qdx9X1"


@app.on_message(filters.command("planet"))
async def planet_info(client, message):
    if len(message.command) < 2:
        await message.reply_text("ɢɪᴠᴇ ᴍᴇ ᴘʟᴀɴᴇᴛ ɴᴀᴍᴇ ᴀғᴛᴇʀ ᴄᴏᴍᴍᴀɴᴅ")
        return

    planet_name = message.command[1]
    api_url = f"https://api.api-ninjas.com/v1/planets?name={planet_name}"
    response = requests.get(api_url, headers={"X-Api-Key": "API_KEY"})

    if response.status_code == requests.codes.ok:
        planet_data = response.json()[0]
        formatted_data = "\n".join(
            [
                f"{key.capitalize().replace('_', ' ')}: {value}"
                for key, value in planet_data.items()
            ]
        )
        await message.reply_text(formatted_data)
    else:
        await message.reply_text(f"Error: {response.status_code} {response.text}")

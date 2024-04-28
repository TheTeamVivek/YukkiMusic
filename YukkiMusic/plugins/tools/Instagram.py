from pyrogram import Client
from pyrogram import filters

import requests
from YukkiMusic import app


@app.on_message(filters.command("im", prefixes="/"))
async def download_instagram_video(client, message):
    if len(message.command) < 3:
        await message.reply_text(
            "Please provide the Instagram video URL after the command."
        )
        return

    url = message.command[1]
    api_url = (
        f"https://nodejs-1xn1lcfy3-jobians.vercel.app/v2/downloader/instagram?url={url}"
    )

    response = requests.get(api_url)
    data = response.json()

    if data["status"]:
        video_url = data["data"][0]["url"]
        await client.send_video(message.chat.id, video_url)
    else:
        await message.reply_text("Failed to download video.")

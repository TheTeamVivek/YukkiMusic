#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import speedtest

from strings import command
from YukkiMusic import app
from YukkiMusic.misc import SUDOERS


async def testspeed(m):
    try:
        test = speedtest.Speedtest()
        test.get_best_server()
        m = await m.edit("⇆ Running Download Speedtest ...")
        test.download()
        m = await m.edit("⇆ Running Upload SpeedTest...")
        test.upload()
        test.results.share()
        result = test.results.dict()
        m = await m.edit("↻ Sharing SpeedTest results")
    except Exception as e:
        return await m.edit(e)
    return result


@app.on_message(command("SPEEDTEST_COMMAND") & SUDOERS)
async def speedtest_function(client, message):
    m = await message.reply_text("ʀᴜɴɴɪɴɢ sᴘᴇᴇᴅᴛᴇsᴛ")
    result = await testspeed(m)
    output = f"""**Speedtest Results**
    
<u>**Client:**</u>
**ISP :** {result["client"]["isp"]}
**Country :** {result["client"]["country"]}
  
<u>**Server:**</u>
**Name :** {result["server"]["name"]}
**Country:** {result["server"]["country"]}, {result["server"]["cc"]}
**Sponsor:** {result["server"]["sponsor"]}
**Latency:** {result["server"]["latency"]}  
**Ping :** {result["ping"]}"""
    msg = await app.send_photo(
        chat_id=message.chat.id, photo=result["share"], caption=output
    )
    await m.delete()

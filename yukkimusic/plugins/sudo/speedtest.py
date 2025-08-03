#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import asyncio

import speedtest

from strings import command, pick_commands
from yukkimusic import app
from yukkimusic.misc import SUDOERS

from . import mhelp


async def testspeed(m):
    try:
        test = await asyncio.to_thread(speedtest.Speedtest)
        await asyncio.to_thread(test.get_best_server)

        m = await m.edit("⇆ Running Download Speedtest ...")
        await asyncio.to_thread(test.download)

        m = await m.edit("⇆ Running Upload SpeedTest...")
        await asyncio.to_thread(test.upload)

        await asyncio.to_thread(test.results.share)
        result = await asyncio.to_thread(test.results.dict)

        m = await m.edit("↻ Sharing SpeedTest results")
    except Exception as e:
        return await m.edit(str(e))

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
    await app.send_photo(chat_id=message.chat.id, photo=result["share"], caption=output)
    await m.delete()


(
    mhelp.add(
        "en", f"<b>{pick_commands('SPEEDTEST_COMMAND')}</b> - Check server speeds"
    )
    .add("ar", f"<b>{pick_commands('SPEEDTEST_COMMAND')}</b> - فحص سرعة الخادم")
    .add("as", f"<b>{pick_commands('SPEEDTEST_COMMAND')}</b> - ছাৰ্ভাৰৰ গতি পৰীক্ষা কৰক")
    .add("hi", f"<b>{pick_commands('SPEEDTEST_COMMAND')}</b> - सर्वर की गति जांचें")
    .add("ku", f"<b>{pick_commands('SPEEDTEST_COMMAND')}</b> - خێرایی ڕاژەکار بپشکنە")
    .add(
        "tr", f"<b>{pick_commands('SPEEDTEST_COMMAND')}</b> - Sunucu hızını kontrol et"
    )
)

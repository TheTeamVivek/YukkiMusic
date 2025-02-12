#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#


from speedtest import Speedtest

from YukkiMusic import tbot
from YukkiMusic.misc import SUDOERS


async def run_speedtest(m):
    try:
        test = Speedtest()
        test.get_best_server()
        m = await m.edit("⇆ Running Download Speedtest ...")
        test.download()
        m = await m.edit("⇆ Running Upload SpeedTest...")
        test.upload()
        test.results.share()
        result = test.results.dict()
        m = await m.edit("↻ Sharing SpeedTest results")
    except Exception as e:
        await m.edit(f"{type(e).__name__}: {str(e)}")
        return None
    return result


@tbot.on_message(command="SPEEDTEST_COMMAND", from_users=list(SUDOERS))
async def speedtest_function(event):
    m = await event.reply("Running Speedtest...")
    result = await run_speedtest(m)
    if result is not None:
        output = f"""**Speedtest Results**
    
<u>**Client:**</u>
**ISP :** {result['client']['isp']}
**Country :** {result['client']['country']}
  
<u>**Server:**</u>
**Name :** {result['server']['name']}
**Country:** {result['server']['country']}, {result['server']['cc']}
**Sponsor:** {result['server']['sponsor']}
**Latency:** {result['server']['latency']}  
**Ping :** {result['ping']}"""
        await event.respond(file=result["share"], message=output)
        await m.delete()

import asyncio

from YukkiMusic import tbot

from .autoleave import auto_end, auto_leave
from .seeker import timer

tasks = [
    auto_leave,
    auto_end,
    timer,
    leave_if_muted,
]


async def run_all_tasks():
    while True:
        await asyncio.gather(*[tbot.add_task(t) for t in tasks])


asyncio.create_task(run_all_tasks())

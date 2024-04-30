from YukkiMusic.misc import clonedb

active = []
stream = {}


async def add_active_chat(chat_id: int, bot_id: int):
    if chat_id not in active:
        active.append((chat_id, bot_id))


async def remove_active_chat(chat_id: int, bot_id: int):
    if (chat_id, bot_id) in active:
        active.remove((chat_id, bot_id))


async def is_active_chat(chat_id: int, bot_id: int) -> bool:
    if (chat_id, bot_id) not in active:
        return False
    else:
        return True


async def is_streaming(chat_id: int, bot_id: int) -> bool:
    run = stream.get((chat_id, bot_id))
    if not run:
        return False
    return run


async def stream_on(chat_id: int, bot_id: int):
    stream[(chat_id, bot_id)] = True


async def stream_off(chat_id: int, bot_id: int):
    stream[(chat_id, bot_id)] = False


async def _clear_(chat_id, bot_id):
    try:
        clonedb[(chat_id, bot_id)] = []
        await remove_active_chat(chat_id, bot_id)
    except:
        return

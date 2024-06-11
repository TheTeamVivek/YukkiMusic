import glob
from os.path import dirname, isfile
from functools import wraps
from pyrogram.handlers import MessageHandler
from pyrogram import filters

def userbot_command(command, prefixes=["."]):
    def decorator(func):
        @wraps(func)
        async def wrapper(client, message):
            return await func(client, message)

        for userbot_client in [userbot.one, userbot.two, userbot.three, userbot.four, userbot.five]:
            if userbot_client:
                userbot_client.add_handler(
                    MessageHandler(wrapper, filters.command(command, prefixes=prefixes))
                )
        return wrapper
    return decorator

def __list_all_modules():
    work_dir = dirname(__file__)
    mod_paths = glob.glob(work_dir + "/*/*.py")

    all_modules = [
        (((f.replace(work_dir, "")).replace("/", "."))[:-3])
        for f in mod_paths
        if isfile(f)
        and f.endswith(".py")
        and not f.endswith("__init__.py")
    ]

    return all_modules


USERBOT_MODULES = sorted(__list_all_modules())
__all__ = USERBOT_MODULES + ["USERBOT_MODULES"]
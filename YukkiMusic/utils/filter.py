import pyrogram.handlers
from pyrogram import filters

from YukkiMusic.core.userbot import clients
from YukkiMusic.utils.admin_check import admin_check


async def admin_filter_f(filt, client, message):
    return not message.edit_date and await admin_check(message)


admin_filter = filters.create(func=admin_filter_f, name="AdminFilter")


def register_all_clients(command, prefix=["/"], *additional_filters):
    def decorator(func):
        combined_filter = filters.command(command, prefixes=prefix)
        for additional_filter in additional_filters:
            combined_filter &= additional_filter
        for client in clients:
            client.add_handler(pyrogram.handlers.MessageHandler(func, combined_filter))
        return func

    return decorator

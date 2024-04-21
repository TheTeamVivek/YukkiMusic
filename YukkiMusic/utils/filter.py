from pyrogram import filters

from YukkiMusic.utils.admin_check import admin_check


async def admin_filter_f(filt, client, message):
    return not message.edit_date and await admin_check(message)


admin_filter = filters.create(func=admin_filter_f, name="AdminFilter")

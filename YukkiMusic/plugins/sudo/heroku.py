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
import math
import os
import shutil
import sys
from datetime import datetime

import dotenv
import heroku3
import requests
import urllib3
from git import Repo
from git.exc import GitCommandError, InvalidGitRepositoryError

import config
from YukkiMusic import tbot
from YukkiMusic.core import filters
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import BANNED_USERS, HAPP, SUDOERS, db, is_heroku
from YukkiMusic.utils import (
    admin_actual,
    get_active_chats,
    get_cmode,
    language,
    pastebin,
    remove_active_chat,
    remove_active_video_chat,
)

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


@tbot.on_message(filters.command("GETLOG_COMMAND", True) & filters.user(SUDOERS))
@language
async def log_(event, _):
    async def _get_log():
        try:
            with open(config.LOG_FILE_NAME) as log:
                lines = log.readlines()
            num = (
                int(event.text.split(None, 1)[1])
                if len(event.text.split()) > 1
                else 100
            )
            return await pastebin.paste("".join(lines[-num:]))
        except Exception:
            return None

    try:
        if await is_heroku():
            if HAPP:
                log_data = HAPP.get_log()
                link = await pastebin.paste(log_data)
            else:
                link = await _get_log()
        else:
            link = await _get_log()

        if link:
            return await event.reply(link)
        return await event.reply(_["heroku_2"])

    except Exception:
        await event.reply(_["heroku_2"])


@tbot.on_message(filters.command("GETVAR_COMMAND", True) & filters.user(SUDOERS))
@language
async def varget_(event, _):
    usage = _["heroku_3"]
    if len(event.text.split()) != 2:
        return await event.reply(usage)
    check_var = event.text.split(None, 2)[1]
    if await is_heroku():
        if HAPP is None:
            return await event.reply(_["heroku_1"])
        heroku_config = HAPP.config()
        if check_var in heroku_config:
            return await event.reply(f"**{check_var}:** `{heroku_config[check_var]}`")
        else:
            return await event.reply(_["heroku_4"])
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await event.reply(_["heroku_5"])
        output = dotenv.get_key(path, check_var)
        if not output:
            await event.reply(_["heroku_4"])
        else:
            return await event.reply(f"**{check_var}:** `{str(output)}`")


@tbot.on_message(filters.command("DELVAR_COMMAND", True) & filters.user(SUDOERS))
@language
async def vardel_(event, _):
    usage = _["heroku_6"]
    if len(event.text.split()) != 2:
        return await event.reply(usage)
    check_var = event.text.split(None, 2)[1]
    if await is_heroku():
        if HAPP is None:
            return await event.reply(_["heroku_1"])
        heroku_config = HAPP.config()
        if check_var in heroku_config:
            await event.reply(_["heroku_7"].format(check_var))
            del heroku_config[check_var]
        else:
            return await event.reply(_["heroku_4"])
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await event.reply(_["heroku_5"])
        output = dotenv.unset_key(path, check_var)
        if not output[0]:
            return await event.reply(_["heroku_4"])
        else:
            await event.reply(_["heroku_7"].format(check_var))
            os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")


@tbot.on_message(filters.command("SETVAR_COMMAND", True) & filters.user(SUDOERS))
@language
async def set_var(event, _):
    usage = _["heroku_8"]
    if len(event.text.split()) < 3:
        return await event.reply(usage)
    to_set = event.text.split(None, 2)[1].strip()
    value = event.text.split(None, 2)[2].strip()
    if await is_heroku():
        if HAPP is None:
            return await event.reply(_["heroku_1"])
        heroku_config = HAPP.config()
        if to_set in heroku_config:
            await event.reply(_["heroku_9"].format(to_set))
        else:
            await event.reply(_["heroku_10"].format(to_set))
        heroku_config[to_set] = value
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await event.reply(_["heroku_5"])
        dotenv.set_key(path, to_set, value)
        if dotenv.get_key(path, to_set):
            await event.reply(_["heroku_9"].format(to_set))
        else:
            await event.reply(_["heroku_10"].format(to_set))
        os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")


@tbot.on_message(filters.command("USAGE_COMMAND", True) & filters.user(SUDOERS))
@language
async def usage_dynos(event, _):
    ### Credits CatUserbot
    if await is_heroku():
        if HAPP is None:
            return await event.reply(_["heroku_1"])
    else:
        return await event.reply(_["heroku_11"])
    dyno = await event.reply(_["heroku_12"])
    heroku = heroku3.from_key(config.HEROKU_API_KEY)
    account_id = heroku.account().id
    user_agent = (
        "Mozilla/5.0 (Linux; Android 10; SM-G975F) "
        "AppleWebKit/537.36 (KHTML, like Gecko) "
        "Chrome/80.0.3987.149 Mobile Safari/537.36"
    )
    headers = {
        "User-Agent": user_agent,
        "Authorization": f"Bearer {config.HEROKU_API_KEY}",
        "Accept": "application/vnd.heroku+json; version=3.account-quotas",
    }
    path = "/accounts/" + account_id + "/actions/get-quota"
    r = requests.get("https://api.heroku.com" + path, headers=headers)
    if r.status_code != 200:
        return await dyno.edit("Unable to fetch.")
    result = r.json()
    quota = result["account_quota"]
    quota_used = result["quota_used"]
    remaining_quota = quota - quota_used
    percentage = math.floor(remaining_quota / quota * 100)
    minutes_remaining = remaining_quota / 60
    hours = math.floor(minutes_remaining / 60)
    minutes = math.floor(minutes_remaining % 60)
    app = result["apps"]
    try:
        app[0]["quota_used"]
    except IndexError:
        app_quota_used = 0
        app_percentage = 0
    else:
        app_quota_used = app[0]["quota_used"] / 60
        app_percentage = math.floor(app[0]["quota_used"] * 100 / quota)
    app_hours = math.floor(app_quota_used / 60)
    app_minutes = math.floor(app_quota_used % 60)
    await asyncio.sleep(1.5)
    text = f"""
<b>Dyno usage</b>

<u>Usage:</u><br>
Total used: <code>{app_hours}</code><b>h</b> <code>{app_minutes}</code><b>m</b> 
[<b>{app_percentage}%</b>]

<u>Remaining Quota</u><br>
Total Left: <code>{hours}</code><b>h</b> <code>{minutes}</code><b>m</b> 
[<b>{percentage}%</b>]"""
    return await dyno.edit(text, parse_mode="HTML")


@tbot.on_message(filters.command("UPDATE_COMMAND", True) & filters.user(SUDOERS))
@language
async def update_(event, _):
    if await is_heroku():
        if HAPP is None:
            return await event.reply(_["heroku_1"])
    response = await event.reply(_["heroku_13"])
    try:
        repo = Repo()
    except GitCommandError:
        return await response.edit(_["heroku_14"])
    except InvalidGitRepositoryError:
        return await response.edit(_["heroku_15"])
    to_exc = f"git fetch origin {config.UPSTREAM_BRANCH} &> /dev/null"
    os.system(to_exc)
    await asyncio.sleep(7)
    verification = ""
    REPO_ = repo.remotes.origin.url.split(".git")[0]
    for checks in repo.iter_commits(f"HEAD..origin/{config.UPSTREAM_BRANCH}"):
        verification = str(checks.count())
    if verification == "":
        return await response.edit("Bot is up to date")

    def ordinal(num):
        suffix = "tsnrhtdd"[(num // 10 % 10 != 1) * (num % 10 < 4) * num % 10 :: 4]
        return f"{num}{suffix}"

    updates = "".join(
        f"<b>➣ #{info.count()}: "
        f"<a href={REPO_}/commit/{info}>{info.summary}</a> By -> {info.author}</b>\n"
        f"\t\t\t\t<b>➥ Commited On:</b> {ordinal(int(datetime.fromtimestamp(info.committed_date).strftime('%d')))} "
        f"{datetime.fromtimestamp(info.committed_date).strftime('%b')}, "
        f"{datetime.fromtimestamp(info.committed_date).strftime('%Y')}\n\n"
        for info in repo.iter_commits(f"HEAD..origin/{config.UPSTREAM_BRANCH}")
    )

    _update_response_ = (
        "<b>A new upadte is available for the Bot!</b>\n\n"
        "➣ Pushing upadtes Now\n\n__<b>Updates:</b>__\n"
    )
    _final_updates_ = f"{_update_response_} {updates}"

    if len(_final_updates_) > 4096:
        url = await pastebin.paste(updates)
        nrs = await response.edit(
            f"<b>A new upadte is available for the Bot!</b>\n\n"
            f"➣ Pushing upadtes Now\n\n__**Updates:**__\n\n[Check Upadtes]({url})",
            link_preview=False,
            parse_mode="HTML",
        )
    else:
        nrs = await response.edit(_final_updates_, link_preview=False)
    os.system("git stash &> /dev/null && git pull")

    try:
        served_chats = await get_active_chats()
        for x in served_chats:
            try:
                await tbot.send_message(
                    entity=int(x),
                    message=f"{tbot.mention} Is upadted herself\n\n"
                    "You can start playing after 15-20 Seconds",
                )
                await remove_active_chat(x)
                await remove_active_video_chat(x)
            except Exception:
                pass
        await response.edit(
            _final_updates_
            + f"» Bot Upadted Sucessfully Now wait until the bot starts",
            link_preview=False,
        )
    except Exception:
        pass

    if await is_heroku():
        try:
            os.system(
                f"git push https://heroku:{config.HEROKU_API_KEY}@git.heroku.com/{config.HEROKU_APP_NAME}.git HEAD:main"
            )
            return
        except Exception as err:
            await response.edit(
                f"{nrs.text}\n\nSomething went wrong, Please check logs"
            )
            return await tbot.send_message(
                entity=config.LOGGER_ID,
                message="An exception occurred #updater due to : <code>{}</code>".format(
                    err
                ),
            )
    else:
        os.system("pip3 install --no-cache-dir -U -r requirements.txt")
        os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")
        sys.exit()


@tbot.on_message(
    filters.command("REBOOT_COMMAND", True)
    & filters.group
    & ~filters.user(BANNED_USERS)
)
@admin_actual
async def reboot(event, _):
    mystic = await event.reply(
        f"Please Wait... \nRebooting{tbot.mention} For Your Chat."
    )
    await asyncio.sleep(1)
    try:
        db[event.chat_id] = []
        await Yukki.stop_stream(event.chat_id)
    except Exception:
        pass
    chat_id = await get_cmode(event.chat_id)
    if chat_id:
        try:
            db[chat_id] = []
            await Yukki.stop_stream(chat_id)
        except Exception:
            pass
    return await mystic.edit("Sucessfully Restarted \nTry playing Now..")


@tbot.on_message(filters.command("RESTART_COMMAND", True) & ~filters.user(BANNED_USERS))
async def restart_(event):
    if event.sender_id not in SUDOERS:
        if event.is_private:
            return
        return await reboot(event)
    response = await event.reply("Restarting...")
    ac_chats = await get_active_chats()
    for x in ac_chats:
        try:
            await tbot.send_message(
                int(x),
                message=f"{tbot.mention} Is restarting...\n\nYou can start playing after 15-20 seconds",
            )
            await remove_active_chat(x)
            await remove_active_video_chat(x)
        except Exception:
            pass

    try:
        shutil.rmtree("downloads")
        shutil.rmtree("raw_files")
        shutil.rmtree("cache")
    except Exception:
        pass
    await response.edit(
        "Restart process started, please wait for few seconds until the bot starts..."
    )
    os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")

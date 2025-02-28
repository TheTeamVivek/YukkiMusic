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
import socket
from datetime import datetime

import dotenv
import heroku3
import requests
import urllib3
from git import Repo
from git.exc import GitCommandError, InvalidGitRepositoryError
from pyrogram import filters
from pyrogram.enums import ChatType
from pyrogram.types import Message

import config
from config import BANNED_USERS
from strings import command
from YukkiMusic import app
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import HAPP, SUDOERS, XCB, db
from YukkiMusic.utils.database import (
    get_active_chats,
    get_cmode,
    remove_active_chat,
    remove_active_video_chat,
)
from YukkiMusic.utils.decorators import admin_actual, language
from YukkiMusic.utils.decorators.language import language
from YukkiMusic.utils.pastebin import paste

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


async def is_heroku():
    return "heroku" in socket.getfqdn()


async def paste_neko(code: str):
    return await paste(code)


@app.on_message(command("GETLOG_COMMAND") & SUDOERS)
@language
async def log_(client, message, _):
    async def _get_log():
        log = open(config.LOG_FILE_NAME)
        lines = log.readlines()
        log.close()
        data = ""
        try:
            num = int(message.text.split(None, 1)[1])
        except Exception:
            num = 100
        for x in lines[-num:]:
            data += x
        link = await paste(data)
        return link

    try:
        if await is_heroku():
            if HAPP is None:
                if os.path.exists(config.LOG_FILE_NAME):
                    return await message.reply(await _get_log())
                return await message.reply(_["heroku_1"])
            data = HAPP.get_log()
            link = await paste(data)
            return await message.reply(link)
        else:
            if os.path.exists(config.LOG_FILE_NAME):
                link = await _get_log()
                return await message.reply(link)
            else:
                return await message.reply(_["heroku_2"])
    except Exception:
        await message.reply(_["heroku_2"])


@app.on_message(command("GETVAR_COMMAND") & SUDOERS)
@language
async def varget_(client, message, _):
    usage = _["heroku_3"]
    if len(message.command) != 2:
        return await message.reply(usage)
    check_var = message.text.split(None, 2)[1]
    if await is_heroku():
        if HAPP is None:
            return await message.reply(_["heroku_1"])
        heroku_config = HAPP.config()
        if check_var in heroku_config:
            return await message.reply(
                f"**{check_var}:** `{heroku_config[check_var]}`"
            )
        else:
            return await message.reply(_["heroku_4"])
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await message.reply(_["heroku_5"])
        output = dotenv.get_key(path, check_var)
        if not output:
            await message.reply(_["heroku_4"])
        else:
            return await message.reply(f"**{check_var}:** `{str(output)}`")


@app.on_message(command("DELVAR_COMMAND") & SUDOERS)
@language
async def vardel_(client, message, _):
    usage = _["heroku_6"]
    if len(message.command) != 2:
        return await message.reply(usage)
    check_var = message.text.split(None, 2)[1]
    if await is_heroku():
        if HAPP is None:
            return await message.reply(_["heroku_1"])
        heroku_config = HAPP.config()
        if check_var in heroku_config:
            await message.reply(_["heroku_7"].format(check_var))
            del heroku_config[check_var]
        else:
            return await message.reply(_["heroku_4"])
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await message.reply(_["heroku_5"])
        output = dotenv.unset_key(path, check_var)
        if not output[0]:
            return await message.reply(_["heroku_4"])
        else:
            await message.reply(_["heroku_7"].format(check_var))
            os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")


@app.on_message(command("SETVAR_COMMAND") & SUDOERS)
@language
async def set_var(client, message, _):
    usage = _["heroku_8"]
    if len(message.command) < 3:
        return await message.reply(usage)
    to_set = message.text.split(None, 2)[1].strip()
    value = message.text.split(None, 2)[2].strip()
    if await is_heroku():
        if HAPP is None:
            return await message.reply(_["heroku_1"])
        heroku_config = HAPP.config()
        if to_set in heroku_config:
            await message.reply(_["heroku_9"].format(to_set))
        else:
            await message.reply(_["heroku_10"].format(to_set))
        heroku_config[to_set] = value
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await message.reply(_["heroku_5"])
        dotenv.set_key(path, to_set, value)
        if dotenv.get_key(path, to_set):
            await message.reply(_["heroku_9"].format(to_set))
        else:
            await message.reply(_["heroku_10"].format(to_set))
        os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")


@app.on_message(command("USAGE_COMMAND") & SUDOERS)
@language
async def usage_dynos(client, message, _):
    ### Credits CatUserbot
    if await is_heroku():
        if HAPP is None:
            return await message.reply(_["heroku_1"])
    else:
        return await message.reply(_["heroku_11"])
    dyno = await message.reply(_["heroku_12"])
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
**Dyno usage**

<u>Usage:</u>
Total used: `{app_hours}`**h**  `{app_minutes}`**m**  [`{app_percentage}`**%**]

<u>Remaining Quota</u>
Total Left: `{hours}`**h**  `{minutes}`**m**  [`{percentage}`**%**]"""
    return await dyno.edit(text)


@app.on_message(command("UPDATE_COMMAND") & SUDOERS)
@language
async def update_(client, message, _):
    if await is_heroku():
        if HAPP is None:
            return await message.reply(_["heroku_1"])
    response = await message.reply(_["heroku_13"])
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
        "**A new upadte is available for the Bot!**\n\n"
        "➣ Pushing upadtes Now\n\n__**Updates:**__\n"
    )
    _final_updates_ = f"{_update_response_} {updates}"

    if len(_final_updates_) > 4096:
        url = await paste(updates)
        nrs = await response.edit(
            f"**A new upadte is available for the Bot!**\n\n"
            f"➣ Pushing upadtes Now\n\n__**Updates:**__\n\n[Check Upadtes]({url})",
            link_preview=False,
        )
    else:
        nrs = await response.edit(_final_updates_, link_preview=False)
    os.system("git stash &> /dev/null && git pull")

    try:
        served_chats = await get_active_chats()
        for x in served_chats:
            try:
                await app.send_message(
                    chat_id=int(x),
                    text=f"{app.mention} Is upadted herself\n\n"
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
                f"{XCB[5]} {XCB[7]} {XCB[9]}{XCB[4]}{XCB[0]*2}{XCB[6]}{XCB[4]}{XCB[8]}{XCB[1]}{XCB[5]}{XCB[2]}{XCB[6]}{XCB[2]}{XCB[3]}{XCB[0]}{XCB[10]}{XCB[2]}{XCB[5]} {XCB[11]}{XCB[4]}{XCB[12]}"
            )
            return
        except Exception as err:
            await response.edit(
                f"{nrs.text}\n\nSomething went wrong, Please check logs"
            )
            return await app.send_message(
                chat_id=config.LOGGER_ID,
                text="An exception occurred #updater due to : <code>{}</code>".format(
                    err
                ),
            )
    else:
        os.system("pip3 install --no-cache-dir -U -r requirements.txt")
        os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")
        exit()


@app.on_message(command("REBOOT_COMMAND") & filters.group & ~BANNED_USERS)
@admin_actual
async def reboot(client, message: Message, _):
    mystic = await message.reply(
        f"Please Wait... \nRebooting{app.mention} For Your Chat."
    )
    await asyncio.sleep(1)
    try:
        db[message.chat.id] = []
        await Yukki.stop_stream(message.chat.id)
    except Exception:
        pass
    chat_id = await get_cmode(message.chat.id)
    if chat_id:
        try:
            await app.get_chat(chat_id)
        except Exception:
            pass
        try:
            db[chat_id] = []
            await Yukki.stop_stream(chat_id)
        except Exception:
            pass
    return await mystic.edit_text("Sucessfully Restarted \nTry playing Now..")


@app.on_message(command("RESTART_COMMAND") & ~BANNED_USERS)
async def restart_(client, message):
    if message.from_user and not message.from_user.id in SUDOERS:
        if message.chat.type not in [ChatType.GROUP, ChatType.SUPERGROUP]:
            return
        return await reboot(client, message)
    response = await message.reply("Restarting...")
    ac_chats = await get_active_chats()
    for x in ac_chats:
        try:
            await app.send_message(
                chat_id=int(x),
                text=f"{app.mention} Is restarting...\n\nYou can start playing after 15-20 seconds",
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
    await response.edit_text(
        "Restart process started, please wait for few seconds until the bot starts..."
    )
    os.system(f"kill -9 {os.getpid()} && python3 -m YukkiMusic")

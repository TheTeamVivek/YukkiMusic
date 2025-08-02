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
import os
import socket
from datetime import datetime

import aiofiles
import dotenv
from git import Repo
from git.exc import GitCommandError, InvalidGitRepositoryError

import config
from strings import command
from yukkimusic import app
from yukkimusic.misc import HAPP, SUDOERS
from yukkimusic.utils import pastebin
from yukkimusic.utils.database import (
    get_active_chats,
    remove_active_chat,
    remove_active_video_chat,
)
from yukkimusic.utils.decorators import asyncify, language


@asyncify
def is_heroku():
    return "heroku" in socket.getfqdn()


@app.on_message(command("GETLOG_COMMAND") & SUDOERS)
@language
async def log_(_, message, lang):
    async def _get_log():
        async with aiofiles.open("logs.txt") as f:
            lines = await f.readlines()

        data = ""
        try:
            num = int(message.text.split(None, 1)[1])
        except Exception:
            num = 100
        for x in lines[-num:]:
            data += x
        link = await pastebin.paste(data)
        return link

    try:
        if await is_heroku():
            if HAPP is None:
                if os.path.exists("logs.txt"):
                    return await message.reply_text(await _get_log())
                return await message.reply_text(lang["heroku_1"])
            data = HAPP.get_log()
            link = await pastebin.paste(data)
            return await message.reply_text(link)
        if os.path.exists("logs.txt"):
            link = await _get_log()
            return await message.reply_text(link)
        return await message.reply_text(lang["heroku_2"])
    except Exception:
        await message.reply_text(lang["heroku_2"])


@app.on_message(command("GETVAR_COMMAND") & SUDOERS)
@language
async def varget_(_, message, lang):
    usage = lang["heroku_3"]
    if len(message.command) != 2:
        return await message.reply_text(usage)
    check_var = message.text.split(None, 2)[1]
    if await is_heroku():
        if HAPP is None:
            return await message.reply_text(lang["heroku_1"])
        heroku_config = HAPP.config()
        if check_var in heroku_config:
            return await message.reply_text(
                f"**{check_var}:** `{heroku_config[check_var]}`"
            )
        return await message.reply_text(lang["heroku_4"])
    path = dotenv.find_dotenv()
    if not path:
        return await message.reply_text(lang["heroku_5"])
    output = dotenv.get_key(path, check_var)
    if not output:
        return await message.reply_text(lang["heroku_4"])
    return await message.reply_text(f"**{check_var}:** `{str(output)}`")


@app.on_message(command("DELVAR_COMMAND") & SUDOERS)
@language
async def vardel_(client, message, _):
    usage = _["heroku_6"]
    if len(message.command) != 2:
        return await message.reply_text(usage)
    check_var = message.text.split(None, 2)[1]
    if await is_heroku():
        if HAPP is None:
            return await message.reply_text(_["heroku_1"])
        heroku_config = HAPP.config()
        if check_var in heroku_config:
            await message.reply_text(_["heroku_7"].format(check_var))
            del heroku_config[check_var]
        else:
            return await message.reply_text(_["heroku_4"])
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await message.reply_text(_["heroku_5"])
        output = dotenv.unset_key(path, check_var)
        if not output[0]:
            return await message.reply_text(_["heroku_4"])
        await message.reply_text(_["heroku_7"].format(check_var))
        os.system(f"kill -9 {os.getpid()} && python3 -m yukkimusic")


@app.on_message(command("SETVAR_COMMAND") & SUDOERS)
@language
async def set_var(client, message, _):
    usage = _["heroku_8"]
    if len(message.command) < 3:
        return await message.reply_text(usage)
    to_set = message.text.split(None, 2)[1].strip()
    value = message.text.split(None, 2)[2].strip()
    if await is_heroku():
        if HAPP is None:
            return await message.reply_text(_["heroku_1"])
        heroku_config = HAPP.config()
        if to_set in heroku_config:
            await message.reply_text(_["heroku_9"].format(to_set))
        else:
            await message.reply_text(_["heroku_10"].format(to_set))
        heroku_config[to_set] = value
    else:
        path = dotenv.find_dotenv()
        if not path:
            return await message.reply_text(_["heroku_5"])
        dotenv.set_key(path, to_set, value)
        if dotenv.get_key(path, to_set):
            await message.reply_text(_["heroku_9"].format(to_set))
        else:
            await message.reply_text(_["heroku_10"].format(to_set))
        os.system(f"kill -9 {os.getpid()} && python3 -m yukkimusic")


@app.on_message(command("UPDATE_COMMAND") & SUDOERS)
@language
async def update_(client, message, _):
    if await is_heroku():
        if HAPP is None:
            return await message.reply_text(_["heroku_1"])
    response = await message.reply_text(_["heroku_13"])
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

    def ordinal(fmt):
        return "%d%s" % (
            fmt,
            "tsnrhtdd"[(fmt // 10 % 10 != 1) * (fmt % 10 < 4) * fmt % 10 :: 4],
        )

    updates = "".join(
        f"<b>➣ #{info.count()}: <a href={REPO_}/commit/{info}>{info.summary}</a> By -> {info.author}</b>\n\t\t\t\t<b>➥ Committed On:</b> {ordinal(int(datetime.fromtimestamp(info.committed_date).strftime('%d')))} {datetime.fromtimestamp(info.committed_date).strftime('%b')}, {datetime.fromtimestamp(info.committed_date).strftime('%Y')}\n\n"
        for info in repo.iter_commits(f"HEAD..origin/{config.UPSTREAM_BRANCH}")
    )
    _update_response_ = "**A new update is available for the Bot! **\n\n➣ Pushing updates Now\n\n__**Updates:**__\n"
    _final_updates_ = f"{_update_response_} {updates}"

    if len(_final_updates_) > 4096:
        url = await pastebin.paste(updates)
        nrs = await response.edit(
            f"**A new update is available for the Bot!**\n\n➣ Pushing updates Now\n\n__**Updates:**__\n\n[Check Updates]({url})",
        )
    else:
        nrs = await response.edit(_final_updates_)
    os.system("git stash &> /dev/null && git pull")

    try:
        served_chats = await get_active_chats()
        for x in served_chats:
            try:
                await app.send_message(
                    chat_id=int(x),
                    text="{} Is updated herself\n\nYou can start playing after 15-20 Seconds".format(
                        app.mention
                    ),
                )
                await remove_active_chat(x)
                await remove_active_video_chat(x)
            except Exception:
                pass
        await response.edit(
            f"{_final_updates_}» Bot updated Successfully Now wait until the bot starts"
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
            return await app.send_message(
                chat_id=config.LOGGER_ID,
                text="An exception occurred #updater due to : <code>{}</code>".format(
                    err
                ),
            )
    else:
        os.system(f"kill -9 {os.getpid()} && python3 -m yukkimusic")
        exit()

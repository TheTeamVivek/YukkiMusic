#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import contextlib

# This aeval and sh module is taken from < https://github.com/TheHamkerCat/WilliamButcherBot >
# Credit goes to TheHamkerCat.
#
import os
import traceback
from io import StringIO
from time import time

from telethon import Button, events

from YukkiMusic import tbot
from YukkiMusic.misc import SUDOERS


async def aexec(code, event):
    local_vars = {}
    exec(
        "async def __aexec(event): " + "".join(f"\n {a}" for a in code.split("\n")),
        globals(),
        local_vars,
    )
    __aexec_func = local_vars["__aexec"]
    return await __aexec_func(event)


@tbot.on(
    events.MessageEdited(
        func=flt.command(["ev", "eval"]) & flt.user(SUDOERS) & ~flt.forwarded
    )
)
@tbot.on_message(flt.command(["ev", "eval"]) & flt.user(SUDOERS) & ~flt.forwarded)
async def executor(event):
    if len(event.text.split()) < 2:
        return await event.reply("**Give me something to execute**")
    try:
        cmd = event.text.split(" ", maxsplit=1)[1]
    except IndexError:
        return await event.delete()
    t1 = time()

    redirected_output = StringIO()
    redirected_error = StringIO()
    stdout, stderr, exc = None, None, None

    with contextlib.redirect_stdout(redirected_output), contextlib.redirect_stderr(
        redirected_error
    ):
        try:
            await aexec(cmd, event)
        except Exception:
            exc = traceback.format_exc()

    stdout = redirected_output.getvalue()
    stderr = redirected_error.getvalue()
    evaluation = "\n"
    if exc:
        evaluation += exc
    elif stderr:
        evaluation += stderr
    elif stdout:
        evaluation += stdout
    else:
        evaluation += "Success"

    final_output = f"<b>RESULTS:</b>\n<pre language='python'>{evaluation}</pre>"
    if len(final_output) > 4096:
        filename = "output.txt"
        with open(filename, "w+", encoding="utf8") as out_file:
            out_file.write(str(evaluation))
        t2 = time()
        keyboard = [[Button.inline(text="⏳", data=f"runtime {t2-t1} Seconds")]]
        await event.reply(
            file=filename,
            message=f"<b>EVAL :</b>\n<code>{cmd[0:980]}</code>\n\n<b>Results:</b>\nAttached Document",
            buttons=keyboard,
        )
        await event.delete()
        os.remove(filename)
    else:
        t2 = time()
        keyboard = [
            [
                Button.inline(text="⏳", data=f"runtime {round(t2-t1, 3)} Seconds"),
                Button.inline(text="🗑", data=f"forceclose abc|{event.sender_id}"),
            ]
        ]
        await event.reply(message=final_output, buttons=keyboard, parse_mode="HTML")
        raise events.StopPropagation


@tbot.on(events.CallbackQuery(pattern="runtime"))
async def runtime_func_cq(event):
    data = event.data.decode("utf-8")
    runtime = data.split(None, 1)[1]
    await event.answer(runtime, alert=True)


@tbot.on(events.CallbackQuery(pattern="forceclose"))
async def forceclose_command(event):
    callback_data = event.data.decode("utf-8").strip()
    callback_request = callback_data.split(None, 1)[1]
    query, user_id = callback_request.split("|")
    if event.sender_id != int(user_id):
        try:
            return await event.answer(
                "This is not for you stay away from here", alert=True
            )
        except Exception:
            return
    await event.delete()
    try:
        await event.answer()
    except Exception:
        return


@tbot.on(
    events.MessageEdited(func=flt.command("sh") & flt.user(SUDOERS) & ~flt.forwarded)
)
@tbot.on_message(flt.command("sh") & flt.user(SUDOERS) & ~flt.forwarded)
async def shellrunner(event):
    if len(event.text.split()) < 2:
        return await event.reply("**Give some commands like:**\n/sh git pull")

    text = event.text.split(None, 1)[1]
    output = ""

    if "\n" in text:
        commands = text.split("\n")
        for cmd in commands:
            r = await tbot.run(cmd)
            output += f"<b>Command:</b> {cmd}\n"
            if r.stdout:
                output += f"<b>Output:</b>\n<pre>{r.stdout}</pre>\n"
            if r.stderr:
                output += f"<b>Error:</b>\n<pre>{r.stderr}</pre>\n"
    else:
        r = await tbot.run(text)
        if r.stdout:
            output += f"<b>Output:</b>\n<pre>{r.stdout}</pre>\n"
        if r.stderr:
            output += f"<b>Error:</b>\n<pre>{r.stderr}</pre>\n"

    if not output.strip():
        output = "<b>OUTPUT :</b>\n<code>None</code>"

    if len(output) > 4096:
        with open("output.txt", "w") as file:
            file.write(output)
        await event.reply(
            file="output.txt",
            message="<code>Output</code>",
        )
        os.remove("output.txt")
    else:
        await event.reply(output, parse_mode="HTML")

    raise events.StopPropagation

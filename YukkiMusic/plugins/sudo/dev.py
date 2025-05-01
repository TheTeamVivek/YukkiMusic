#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

# This aeval and sh module is taken from < https://github.com/TheHamkerCat/WilliamButcherBot >
# Credit goes to TheHamkerCat.
#
import contextlib
import io
import textwrap
import traceback
from time import time

from telethon import Button, events

from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.misc import SUDOERS


def cleanup_code(code):
    if code.startswith("```") and code.endswith("```"):
        return "\n".join(code.strip("`").split("\n")[1:-1])
    return code.strip("` \n")


async def aexec(code, event):
    local_vars = {
        "__builtins__": globals()["__builtins__"],
        "tbot": tbot,
        "client": tbot,
        "app": tbot,
        "event": event,
        "rmsg": await event.get_reply_message(),
    }
    to_compile = f"async def __aexec_func():\n{textwrap.indent(code, '  ')}"
    exec(
        to_compile,
        local_vars,
    )
    func = local_vars["__aexec_func"]
    return await func()


@tbot.on(
    events.MessageEdited(func=flt.command(["ev", "eval"]) & SUDOERS & ~flt.forwarded)
)
@tbot.on_message(flt.command(["ev", "eval"]) & SUDOERS & ~flt.forwarded)
async def executor(event):
    if len(event.text.split()) < 2:
        return await event.reply("**Give me something to execute**")
    try:
        cmd = event.raw_text.split(None, 1)[1]
        cmd = cleanup_code(cmd)
    except IndexError:
        return await event.delete()

    t1 = time()
    redirected_output = io.StringIO()
    redirected_error = io.StringIO()
    (
        stdout,
        stderr,
        exc,
        result,
    ) = (
        None,
        None,
        None,
        None,
    )

    with (
        contextlib.redirect_stdout(redirected_output),
        contextlib.redirect_stderr(redirected_error),
    ):
        try:
            result = await aexec(cmd, event)
        except Exception:
            exc = traceback.format_exc()

    stdout = redirected_output.getvalue()
    stderr = redirected_error.getvalue()
    template = "<b>{0}:</b>\n<pre class='python'>{1}</pre>"

    final_output = ""
    if stdout:
        final_output += template.format("Output", stdout)
    if stderr:
        final_output += template.format("Error", stderr)
    if exc:
        final_output += template.format("Exception", exc)
    if result is not None:
        final_output += template.format("Returns", str(result))

    if not final_output:
        final_output = template.format("Result", "Success")

    t2 = time()

    if len(final_output) > 3000:
        text = ""
        if stdout:
            text += "OUTPUT\n" + stdout
        if stderr:
            text += "ERROR\n" + stderr
        if exc:
            text += "EXCEPTION\n" + exc
        if result is not None:
            text += "RETURNS\n" + result
        with io.BytesIO(str(text).encode()) as f:
            f.name = "output.txt"
            keyboard = [
                [
                    Button.inline(
                        text="⏳",
                        data=f"runtime {t2 - t1} Seconds",
                    )
                ]
            ]

            await event.reply(
                file=f,
                message=f"<b>EVAL :</b>\n<code>{cmd[0:980]}</code>\n\n<b>Results:</b>\nAttached Document",
                parse_mode="HTML",
                buttons=keyboard,
            )
        await event.delete()
    else:
        keyboard = [
            [
                Button.inline(
                    text="⏳",
                    data=f"runtime {round(t2 - t1, 3)} Seconds",
                ),
                Button.inline(
                    text="🗑",
                    data=f"forceclose abc|{event.sender_id}",
                ),
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


@tbot.on(events.MessageEdited(func=flt.command("sh") & SUDOERS & ~flt.forwarded))
@tbot.on_message(flt.command("sh") & SUDOERS & ~flt.forwarded)
async def shellrunner(event):
    if len(event.text.split()) < 2:
        return await event.reply("**Give some commands like:**\n/sh git pull")

    text = event.raw_text.split(None, 1)[1]
    output = ""

    if "\n" in text:
        commands = text.split("\n")
        for cmd in commands:
            r = await tbot.run_shell_command(cmd)
            output += f"<b>Command:</b> {cmd}\n"
            if r.stdout:
                output += f"<b>Output:</b>\n<pre>{r.stdout}</pre>\n"
            if r.stderr:
                output += f"<b>Error:</b>\n<pre>{r.stderr}</pre>\n"
    else:
        r = await tbot.run_shell_command(text)
        if r.stdout:
            output += f"<b>Output:</b>\n<pre>{r.stdout}</pre>\n"
        if r.stderr:
            output += f"<b>Error:</b>\n<pre>{r.stderr}</pre>\n"

    if not output.strip():
        output = "<b>OUTPUT :</b>\n<code>None</code>"

    if len(output) > 3000:
        with io.BytesIO(str(output).encode()) as f:
            f.name = "output.txt"

            await event.reply(
                file=f,
                message="<code>Output</code>",
                parse_mode="HTML",
            )

    else:
        await event.reply(output, parse_mode="HTML")

    raise events.StopPropagation

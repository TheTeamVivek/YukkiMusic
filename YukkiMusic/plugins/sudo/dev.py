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

import asyncio
import contextlib
import io
import traceback
from time import time

from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS


def cleanup_code(code):
    if code.startswith("```") and code.endswith("```"):
        return "\n".join(code.strip("`").split("\n")[1:-1])
    return code.strip("` \n")


def get_output(stdout, stderr, exc, result, fmt=False):
    data = {
        "StdOut": stdout,
        "StdError": stderr,
        "Exception": exc,
        "Result": result,
    }

    if fmt:
        template = "**{0}:**\n```python\n{1}\n```"
        output = [template.format(k, v) for k, v in data.items() if v]
        if not output:
            output.append(template.format("Result", "Success"))
        return "".join(output)
    return "".join(f"{k}\n{v}" for k, v in data.items() if v)


async def aexec(code, client, message):
    local_vars = {
        "__builtins__": __builtins__,  # DON'T REMOVE THIS
        "client": client,
        "app": client,
        "message": message,
        "m": message,
        "c": client,
        "rmsg": message.reply_to_message,
    }
    # pylint: disable-next=exec-used
    exec(
        "async def __aexec(): " + "".join(f"\n {a}" for a in code.split("\n")),
        local_vars,
    )
    return await local_vars["__aexec"]()


@app.on_edited_message(
    filters.command(["ev", "eval"]) & SUDOERS & ~filters.forwarded & ~filters.via_bot
)
@app.on_message(
    filters.command(["ev", "eval"]) & SUDOERS & ~filters.forwarded & ~filters.via_bot
)
async def executor(client: app, message: Message):
    if message.edit_hide:
        return
    if len(message.command) < 2:
        return await message.reply(text="<b>Give me something to exceute</b>")
    try:
        cmd = message.text.markdown.split(" ", maxsplit=1)[1]
        cmd = cleanup_code(cmd)
    except IndexError:
        return await message.delete()
    t1 = time()
    (exc,) = (None,)
    with (
        contextlib.redirect_stdout(io.StringIO()) as stdout,
        contextlib.redirect_stderr(io.StringIO()) as stderr,
    ):
        try:
            result = await aexec(
                cmd, client, message
            )  # pylint: disable-next=broad-exception-caught
        except Exception:
            exc = traceback.format_exc()
    t2 = time()
    final_output = get_output(stdout.getvalue(), stderr.getvalue(), exc, result, True)

    if len(final_output) > 3000:
        text = get_output(stdout.getvalue(), stderr.getvalue(), exc, result)

        with io.BytesIO(text.encode()) as f:
            f.name = "output.txt"

            keyboard = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            text="⏳",
                            callback_data=f"runtime {t2 - t1} Seconds",
                        )
                    ]
                ]
            )
            await message.reply_document(
                document=f,
                caption=(
                    f"<b>EVAL :</b>\n<code>{cmd[:980]}</code>"
                    "\n\n<b>Results:</b>\nAttached Document",
                ),
                reply_markup=keyboard,
            )
            await message.delete()
    else:
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="⏳",
                        callback_data=f"runtime {round(t2 - t1, 3)} Seconds",
                    ),
                    InlineKeyboardButton(
                        text="🗑",
                        callback_data=f"forceclose abc|{message.from_user.id}",
                    ),
                ]
            ]
        )
        await message.reply(text=final_output, reply_markup=keyboard)
        await message.stop_propagation()


@app.on_callback_query(filters.regex(r"runtime"))
async def runtime_func_cq(_, cq):
    runtime = cq.data.split(None, 1)[1]
    await cq.answer(runtime, show_alert=True)


@app.on_callback_query(filters.regex("forceclose"))
async def forceclose_command(_, query):
    callback_data = query.data.strip()
    callback_request = callback_data.split(None, 1)[1]
    query, user_id = callback_request.split("|")
    if query.from_user.id != int(user_id):
        return await query.answer(
            "This is not for you stay away from here", show_alert=True
        )
    await query.message.delete()
    await query.answer()


@app.on_edited_message(
    filters.command("sh") & SUDOERS & ~filters.forwarded & ~filters.via_bot
)
@app.on_message(filters.command("sh") & SUDOERS & ~filters.forwarded & ~filters.via_bot)
async def shellrunner(_, message: Message):
    if message.edit_hide:
        return
    if len(message.command) < 2:
        return await message.reply("<b>Give some commands like:</b>\n/sh git pull")

    text = message.text.markdown.split(None, 1)[1]
    output = ""

    async def run_command(command: str, timeout: int = 30):
        try:
            process = await asyncio.create_subprocess_shell(
                command,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )

            stdout, stderr = await asyncio.wait_for(
                process.communicate(), timeout=timeout
            )
            return stdout.decode().strip(), stderr.decode().strip()

        except asyncio.TimeoutError:
            process.kill()
            await process.wait()
            return (
                None,
                "Command timed out after 30 seconds.",
            )  # pylint: disable-next=broad-exception-caught
        except Exception:
            return None, traceback.format_exc()

    if "\n" in text:
        commands = text.split("\n")
        for cmd in commands:
            stdout, stderr = await run_command(cmd)
            output += f"<b>Command:</b> {cmd}\n"
            if stdout:
                output += f"<b>Output:</b>\n<pre>{stdout}</pre>\n"
            if stderr:
                output += f"<b>Error:</b>\n<pre>{stderr}</pre>\n"
    else:
        stdout, stderr = await run_command(text)
        if stdout:
            output += f"<b>Output:</b>\n<pre>{stdout}</pre>\n"
        if stderr:
            output += f"<b>Error:</b>\n<pre>{stderr}</pre>\n"

    if not output.strip():
        output = "<b>OUTPUT :</b>\n<code>None</code>"

    if len(output) > 4000:
        with io.BytesIO(str(output).encode()) as f:
            f.name = "output.txt"
            await app.send_document(
                message.chat.id,
                f,
                reply_to_message_id=message.id,
                caption="<code>Output</code>",
            )
    else:
        await message.reply(text=output)

    await message.stop_propagation()

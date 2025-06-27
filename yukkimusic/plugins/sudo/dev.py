#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/yukkimusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/yukkimusic/blob/master/LICENSE >
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
from pyrogram.enums import ParseMode
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from yukkimusic import app
from yukkimusic.misc import SUDOERS


def cleanup_code(code):
    if code.startswith("```") and code.endswith("```"):
        lines = code.strip("`").split("\n")
        # If the first line after the opening ``` contains a language tag, skip it
        if lines[0].strip() == "" or lines[0].isalpha():
            return "\n".join(lines[1:-1])
        else:
            return "\n".join(lines[:-1])
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
        return "\n\n".join(output)
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
    exc = None
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
                    f"**EVAL :**\n`{cmd[:980]}`" "\n\n**Results:**\nAttached Document"
                ),
                parse_mode=ParseMode.MARKDOWN,  # disable html formatting result can include some tags that can break formatting
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
        await message.reply(
            text=final_output, parse_mode=ParseMode.MARKDOWN, reply_markup=keyboard
        )
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
            output += f"**Command:** {cmd}\n"
            if stdout:
                output += f"**Output:**\n```\n{stdout}```\n"
            if stderr:
                output += f"**Error:**\n```\n{stderr}```\n"
    else:
        stdout, stderr = await run_command(text)
        if stdout:
            output += f"**Output:**\n```\n{stdout}\n```\n"
        if stderr:
            output += f"**Error:**\n```\n{stderr}\n```\n"

    if not output.strip():
        output = "**OUTPUT :**\n`None`"

    if len(output) > 4000:
        with io.BytesIO(str(output).encode()) as f:
            f.name = "output.txt"
            await app.send_document(
                message.chat.id,
                f,
                reply_to_message_id=message.id,
                caption="**Output**",
                parse_mode=ParseMode.MARKDOWN,
            )
    else:
        await message.reply(text=output, parse_mode=ParseMode.MARKDOWN)

    await message.stop_propagation()

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
import os
import traceback
import io
from time import time

import aiofiles
from pyrogram import filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message

from YukkiMusic import app
from YukkiMusic.misc import SUDOERS

def cleanup_code(code):
    if code.startswith("```") and code.endswith("```"):
        return "\n".join(code.strip("`").split("\n")[1:-1])
    return code.strip("` \n")

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
    exec(
        "async def __aexec(): "
        + "".join(f"\n {a}" for a in code.split("\n")),
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
            result = await aexec(cmd, client, message)
        except Exception:
            exc = traceback.format_exc()
    stdout = redirected_output.getvalue()
    stderr = redirected_error.getvalue()
    template = "**{0}:**\n```python\n{1}\n```"
    t2 = time()
    
    final_output = ""
    if stdout:
        final_output += template.format("StdOut", stdout)
    if stderr:
        final_output += template.format("StdError", stderr)
    if exc:
        final_output += template.format("Exception", exc)
    if result is not None:
        final_output += template.format("Result", str(result))

    if not final_output:
        final_output = template.format("Result", "Success")
        
    if len(final_output) > 3000:
        text = ""
        if stdout:
            text += "StdOut\n" + stdout
        if stderr:
            text += "StdError\n" + stderr
        if exc:
            text += "Exception\n" + exc
        if result is not None:
            text += "Result\n" + result
        with io.BytesIO(str(text).encode()) as f:
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
                caption=f"<b>EVAL :</b>\n<code>{cmd[0:980]}</code>\n\n<b>Results:</b>\nAttached Document",
        
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
        try:
            return await query.answer(
                "This is not for you stay away from here", show_alert=True
            )
        except Exception:
            return
    await query.message.delete()
    try:
        await query.answer()
    except Exception:
        return


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

            stdout, stderr = await asyncio.wait_for(process.communicate(), timeout=timeout)
            return stdout.decode().strip(), stderr.decode().strip()
    
        except asyncio.TimeoutError:
            process.kill()
            await process.wait()
            return None, "Command timed out after 30 seconds."

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
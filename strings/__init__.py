#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved

import re
import os
import sys
import yaml
from typing import Dict, List, Union

from pyrogram import filters
from pyrogram.types import Message
from pyrogram import Client
from pyrogram.enums import ChatType

from YukkiMusic.misc import SUDOERS


from YukkiMusic.utils.database import get_lang, is_maintenance

languages = {}
commands = {}
helpers = {}
languages_present = {}


def load_yaml_file(file_path: str) -> dict:
    with open(file_path, "r", encoding="utf8") as file:
        return yaml.safe_load(file)


def get_command(lang: str = "en") -> Union[str, List[str]]:
    if lang not in commands:
        lang = "en"
    return commands[lang]


def get_string(lang: str):
    return languages[lang]


def get_helpers(lang: str):
    return helpers[lang]


# Load English commands first and set English keys
commands["en"] = load_yaml_file(r"./strings/cmds/en.yml")
english_keys = set(commands["en"].keys())

for filename in os.listdir(r"./strings/cmds/"):
    if filename.endswith(".yml") and filename != "en.yml":
        language_code = filename[:-4]
        commands[language_code] = load_yaml_file(
            os.path.join(r"./strings/cmds/", filename)
        )

        missing_keys = english_keys - set(commands[language_code].keys())
        if missing_keys:
            print(
                f"Error: Missing keys in strings/cmds/{language_code}.yml: {', '.join(missing_keys)}"
            )
            sys.exit(1)

for filename in os.listdir(r"./strings/helpers/"):
    if filename.endswith(".yml"):
        language_code = filename[:-4]
        helpers[language_code] = load_yaml_file(
            os.path.join(r"./strings/helpers/", filename)
        )

if "en" not in languages:
    languages["en"] = load_yaml_file(r"./strings/langs/en.yml")
    languages_present["en"] = languages["en"]["name"]

for filename in os.listdir(r"./strings/langs/"):
    if filename.endswith(".yml") and filename != "en.yml":
        language_name = filename[:-4]
        languages[language_name] = load_yaml_file(
            os.path.join(r"./strings/langs/", filename)
        )

        for item in languages["en"]:
            if item not in languages[language_name]:
                languages[language_name][item] = languages["en"][item]

        try:
            languages_present[language_name] = languages[language_name]["name"]
        except KeyError:
            print(
                "There is an issue with the language file. Please report it to TheTeamvk at @TheTeamvk on Telegram"
            )
            sys.exit()

if not commands:
    print(
        "There's a problem loading the command files. Please report it to TheTeamVivek at @TheTeamVivek on Telegram"
    )
    sys.exit()


def command(
    commands: Union[str, List[str]],
    prefixes: Union[str, List[str], None] = "/",
    case_sensitive: bool = False,
):
    async def func(flt, client: Client, message: Message):
        lang_code = await get_lang(message.chat.id)
        try:
            _ = get_string(lang_code)
        except Exception:
            _ = get_string("en")

        if not await is_maintenance():
            if (
                message.from_user and message.from_user.id not in SUDOERS
            ) or not message.from_user:
                if message.chat.type == ChatType.PRIVATE:
                    await message.reply_text(_["maint_4"])
                    return False
                return False

        if isinstance(commands, str):
            commands_list = [commands]
        else:
            commands_list = commands

        # Get localized and English commands
        localized_commands = []
        en_commands = []
        for cmd in commands_list:
            localized_cmd = get_command(lang_code)[cmd]
            if isinstance(localized_cmd, str):
                localized_commands.append(localized_cmd)
            elif isinstance(localized_cmd, list):
                localized_commands.extend(localized_cmd)

            en_cmd = get_command("en")[cmd]
            if isinstance(en_cmd, str):
                en_commands.append(en_cmd)
            elif isinstance(en_cmd, list):
                en_commands.extend(en_cmd)

        username = client.me.username or ""
        text = message.text or message.caption
        message.command = None

        if not text:
            return False

        def match_command(cmd, text, with_prefix=True):
            if with_prefix and flt.prefixes:
                for prefix in flt.prefixes:
                    if text.startswith(prefix):
                        without_prefix = text[len(prefix) :]
                        if re.match(
                            rf"^(?:{cmd}(?:@?{username})?)(?:\s|$)",
                            without_prefix,
                            flags=re.IGNORECASE if not flt.case_sensitive else 0,
                        ):
                            return prefix + cmd
            else:
                # Match without prefix
                if re.match(
                    rf"^(?:{cmd}(?:@?{username})?)(?:\s|$)",
                    text,
                    flags=re.IGNORECASE if not flt.case_sensitive else 0,
                ):
                    return cmd
            return None

        all_commands = []

        # Add English commands with prefix only
        if lang_code == "en":
            all_commands.extend((cmd, True) for cmd in en_commands)  # Only with prefix
        else:
            # For non-English languages, add commands both with and without prefix
            all_commands.extend(
                (cmd, True) for cmd in en_commands
            )  # English commands with prefix
            all_commands.extend(
                (cmd, True) for cmd in localized_commands
            )  # Non-English commands with prefix
            all_commands.extend(
                (cmd, False) for cmd in localized_commands
            )  # Non-English commands without prefix

        for cmd, with_prefix in all_commands:
            matched_cmd = match_command(cmd, text, with_prefix)
            if matched_cmd:
                without_command = re.sub(
                    rf"{matched_cmd}(?:@?{username})?\s?",
                    "",
                    text,
                    count=1,
                    flags=re.IGNORECASE if not flt.case_sensitive else 0,
                )
                message.command = [matched_cmd] + [
                    re.sub(r"\\([\"'])", r"\1", m.group(2) or m.group(3) or "")
                    for m in re.finditer(
                        r'([^\s"\']+)|"([^"]*)"|\'([^\']*)\'', without_command
                    )
                ]
                return True

        return False

    if prefixes == "" or prefixes is None:
        prefixes = set()
    else:
        prefixes = set(prefixes) if isinstance(prefixes, list) else {prefixes}

    return filters.create(
        func,
        "MultilingualCommandFilter",
        commands=commands,
        prefixes=prefixes,
        case_sensitive=case_sensitive,
    )

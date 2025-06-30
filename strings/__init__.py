#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved

# pylint: disable=missing-module-docstring, missing-function-docstring
import os
import random
import re
import sys

import yaml

from .attrdict import AttrDict

languages = AttrDict()
languages_present = AttrDict()
commands: AttrDict
print("yep print working")


def get_command(key, lang=None):
    data = commands.get(key)
    if not data:
        return []
    if lang:
        return list({cmd.lower() for cmd in data.get(lang, data.get("en", []))})
    all_commands = set()
    for lang_commands in data.values():
        all_commands.update(cmd.lower() for cmd in lang_commands)
    return list(all_commands)


def pick_commands(key, lang=None):
    commands_list = get_command(key, lang)
    if commands_list:
        return "/" + random.choice(commands_list)
    return None


def command(
    commands: str | list[str],  # pylint: disable=W0621
    prefixes: str | list[str] | None = "/",
    case_sensitive: bool = False,
):
    if not isinstance(prefixes, list):
        prefixes = [prefixes]
    prefixes.append("")  # Command can work with and without prefix

    if not isinstance(commands, list):
        commands = [commands]
    cmds = []
    for key in commands:
        cmds.extend(get_command(key))

    from pyrogram import filters  # pylint: disable=import-outside-toplevel

    return filters.command(cmds, prefixes=prefixes, case_sensitive=case_sensitive)


def load_yaml(file_path: str) -> dict:
    with open(file_path, encoding="utf8") as file:
        return yaml.safe_load(file)


def get_string(lang: str):
    return languages.get(lang, languages["en"])


def format_value(value, is_command=False):
    if isinstance(value, list):
        if is_command:
            return " ".join(f"/{cmd}" for cmd in value)
        return " ".join(str(v) for v in value)
    return f"/{value}" if is_command else value


def replace_placeholders(text: str, lang_data: dict, lang_code: str = "en") -> str:
    if not isinstance(text, str):
        return text

    pattern = re.compile(r"\{(\w+)(?:\[(\d+)\])?\}")

    def replacer(match):
        key = match.group(1)
        index = match.group(2)

        is_command = key.endswith("_COMMAND")

        if is_command:
            cmds = get_command(key, lang_code)
            if not cmds:
                return match.group(0)

            if index is not None:
                i = int(index)
                return (
                    f"/{cmds[i]}" if 0 <= i < len(cmds) else f"/{random.choice(cmds)}"
                )
            return format_value(cmds, is_command=True)

        return format_value(lang_data.get(key, match.group(0)), is_command=False)

    return pattern.sub(replacer, text)


def update_helpers(data: dict, lang_code: str = "en"):
    if not isinstance(data, dict):
        return data
    for dict_key, value in data.items():
        if isinstance(value, dict):
            data[dict_key] = update_helpers(value, lang_code)
        elif isinstance(value, str):
            data[dict_key] = replace_placeholders(value, data, lang_code)
    return data


commands = AttrDict(load_yaml(os.path.join("strings", "commands.yml")))

if "en" not in languages:
    languages["en"] = load_yaml(os.path.join("strings", "langs", "en.yml"))
    languages_present["en"] = languages["en"]["name"]

languages["en"] = update_helpers(languages["en"], "en")

for filename in os.listdir(os.path.join("strings", "langs")):
    if filename.endswith(".yml") and filename != "en.yml":
        lang_name = filename[:-4]
        lang_path = os.path.join("strings", "langs", filename)
        languages[lang_name] = load_yaml(lang_path)

        for key_holder in languages["en"]:
            if key_holder not in languages[lang_name]:
                languages[lang_name][key_holder] = languages["en"][key_holder]

        try:
            languages_present[lang_name] = languages[lang_name]["name"]
        except KeyError:
            print("There is an issue with the language file. Please report it.")
            sys.exit()

        languages[lang_name] = update_helpers(languages[lang_name], lang_name)

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved

import os
import re
import sys

import yaml

languages = {}
languages_present = {}
commands = {}


def get_command(command, lang=None):
    data = commands.get(command)
    if not data:
        return []
    if lang:
        return list({cmd.lower() for cmd in data.get(lang, data.get("en", []))})
    all_commands = set()
    for lang_commands in data.values():
        all_commands.update(cmd.lower() for cmd in lang_commands)
    return list(all_commands)


def command(
    commands: str | list[str],
    prefixes: str | list[str] | None = "/",
    case_sensitive: bool = False,
):
    if not isinstance(prefixes, list):
        prefixes = [prefixes]
    prefixes.append("") # Command can work with without prefix

    if not isinstance(commands, list):
        commands = [commands]
    cmds = []
    for key in commands:
        cmds.extend(get_command(key))

    from pyrogram import filters

    return filters.command(cmds, prefixes=prefixes, case_sensitive=case_sensitive)


def load_yaml(file_path: str) -> dict:
    with open(file_path, encoding="utf8") as file:
        return yaml.safe_load(file)


def get_string(lang: str):
    return languages.get(lang, languages["en"])


def format_value(value):
    if isinstance(value, list):
        return " ".join(f"/{cmd}" for cmd in value)
    return value


def replace_placeholders(
    text: str, lang_data: dict, outer_key: str = "", lang_code: str = "en"
) -> str:
    if not isinstance(text, str):
        return text

    pattern = re.compile(r"\{(\w+)\}")

    def replacer(match):
        key = match.group(1)
        if key.endswith("_COMMAND"):
            return format_value(get_command(key, lang_code))
        return format_value(lang_data.get(key, match.group(0)))

    return pattern.sub(replacer, text)


def update_helpers(data: dict, lang_code: str = "en"):
    if not isinstance(data, dict):
        return data
    for dict_key, value in data.items():
        if isinstance(value, dict):
            data[dict_key] = update_helpers(value, lang_code)
        elif isinstance(value, str):
            data[dict_key] = replace_placeholders(value, data, dict_key, lang_code)
    return data


commands.update(load_yaml(os.path.join("strings", "commands.yml")))

if "en" not in languages:
    languages["en"] = load_yaml(os.path.join("strings", "langs", "en.yml"))
    languages_present["en"] = languages["en"]["name"]

languages["en"] = update_helpers(languages["en"], "en")

for filename in os.listdir(os.path.join("strings", "langs")):
    if filename.endswith(".yml") and filename != "en.yml":
        lang_name = filename[:-4]
        lang_path = os.path.join("strings", "langs", filename)
        languages[lang_name] = load_yaml(lang_path)
        for key in languages["en"]:
            if key not in languages[lang_name]:
                languages[lang_name][key] = languages["en"][key]
        try:
            languages_present[lang_name] = languages[lang_name]["name"]
        except KeyError:
            print("There is an issue with the language file. Please report it.")
            sys.exit()
        languages[lang_name] = update_helpers(languages[lang_name], lang_name)

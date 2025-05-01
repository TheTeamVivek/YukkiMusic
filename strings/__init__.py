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


def load_yaml(file_path: str) -> dict:
    with open(file_path, encoding="utf8") as file:
        return yaml.safe_load(file)


def get_string(lang: str):
    return languages.get(lang, "en")


def format_value(value):
    if isinstance(value, list):
        return " ".join(f"/{cmd}" for cmd in value)
    return value


def replace_placeholders(text: str, lang_data: dict) -> str:
    if not isinstance(text, str):
        return text
    pattern = re.compile(r"\{(\w+)\}")
    return pattern.sub(
        lambda m: format_value(lang_data.get(m.group(1), m.group(0))), text
    )


def update_helpers(data: dict):
    if not isinstance(data, dict):
        return data
    for dict_key, value in data.items():
        if isinstance(value, dict):
            data[dict_key] = update_helpers(value)
        elif isinstance(value, str):
            data[dict_key] = replace_placeholders(value, data)
    return data


if "en" not in languages:
    languages["en"] = load_yaml(os.path.join("strings", "langs", "en.yml"))
    languages_present["en"] = languages["en"]["name"]

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
        languages[lang_name] = update_helpers(languages[lang_name])

languages["en"] = update_helpers(languages["en"])
commands = load_yaml(os.path.join("strings", "commands.yaml"))


def get_command(command, lang=None):
    data = commands.get(command)
    if not data:
        return []
    if lang:
        return data.get(lang, [])
    all_commands = []
    for lang_commands in data.values():
        all_commands.extend(lang_commands)
    return all_commands

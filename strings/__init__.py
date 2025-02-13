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

helpers_key = [
    "Active",
    "Admin",
    "Auth",
    "Blist",
    "Bot",
    "Dev",
    "Gcast",
    "PList",
    "Play",
]


def load_yaml(file_path: str) -> dict:
    with open(file_path, encoding="utf8") as file:
        return yaml.safe_load(file)


def get_string(lang: str):
    return languages[lang]


def replace_helpers(text: str, lang_data: dict) -> str:
    if not isinstance(text, str):
        return text

    for helper_key in helpers_key:
        pattern = rf"\{{\s*{re.escape(helper_key)}\s*\}}"
        text = re.sub(pattern, lang_data.get(helper_key, helper_key), text)

    return text


def update_helpers(data: dict):
    if not isinstance(data, dict):
        return data

    for dict_key, value in data.items():
        if isinstance(value, dict):
            data[dict_key] = update_helpers(value)
        elif isinstance(value, str):
            data[dict_key] = replace_helpers(value, data)

    return data

if "en" not in languages:
    languages["en"] = load_yaml(r"./strings/langs/en.yml")
    languages_present["en"] = languages["en"]["name"]

for filename in os.listdir(r"./strings/langs/"):
    if filename.endswith(".yml") and filename != "en.yml":
        lang_name = filename[:-4]
        languages[lang_name] = load_yaml(os.path.join(r"./strings/langs/", filename))

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

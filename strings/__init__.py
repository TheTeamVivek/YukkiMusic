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
from typing import Dict, List, Union

import yaml
from pyrogram import Client, filters
from pyrogram.enums import ChatType
from pyrogram.types import Message

from YukkiMusic.misc import SUDOERS
from YukkiMusic.utils.database import get_lang, is_maintenance

languages = {}
languages_present = {}


def load_yaml_file(file_path: str) -> dict:
    with open(file_path, encoding="utf8") as file:
        return yaml.safe_load(file)


def get_string(lang: str):
    return languages[lang]


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

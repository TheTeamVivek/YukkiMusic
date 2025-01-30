#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

import json
import random
from os.path import realpath

from YukkiMusic.core.request import Request


class Carbon:
    def __init__(self):
        self.language = "auto"
        self.drop_shadow = True
        self.drop_shadow_blur = "68px"
        self.drop_shadow_offset = "20px"
        self.font_family = "JetBrains Mono"
        self.width_adjustment = True
        self.watermark = False
        self.themes = [
            "3024-night",
            "a11y-dark",
            "blackboard",
            "base16-dark",
            "base16-light",
            "cobalt",
            "duotone-dark",
            "dracula-pro",
            "hopscotch",
            "lucario",
            "material",
            "monokai",
            "nightowl",
            "nord",
            "oceanic-next",
            "one-light",
            "one-dark",
            "panda-syntax",
            "parasio-dark",
            "seti",
            "shades-of-purple",
            "solarized+dark",
            "solarized+light",
            "synthwave-84",
            "twilight",
            "verminal",
            "vscode",
            "yeti",
            "zenburn",
        ]
        self.colour = [
            "#FF0000",
            "#FF5733",
            "#FFFF00",
            "#008000",
            "#0000FF",
            "#800080",
            "#A52A2A",
            "#FF00FF",
            "#D2B48C",
            "#00FFFF",
            "#808000",
            "#800000",
            "#00FFFF",
            "#30D5C8",
            "#00FF00",
            "#008080",
            "#4B0082",
            "#EE82EE",
            "#FFC0CB",
            "#000000",
            "#FFFFFF",
            "#808080",
        ]

    async def generate(self, text: str, user_id):
        params = {
            "code": text,
            "backgroundColor": random.choice(self.colour),
            "theme": random.choice(self.themes),
            "dropShadow": self.drop_shadow,
            "dropShadowOffsetY": self.drop_shadow_offset,
            "dropShadowBlurRadius": self.drop_shadow_blur,
            "fontFamily": self.font_family,
            "language": self.language,
            "watermark": self.watermark,
            "widthAdjustment": self.width_adjustment,
        }

        resp = await Request.post_raw(
            "https://carbonara.solopov.dev/api/cook",
            data=json.dumps(params),
            headers={"Content-Type": "application/json"},
        )
        with open(f"cache/carbon{user_id}.jpg", "wb") as f:
            f.write(resp)
        return realpath(f.name)

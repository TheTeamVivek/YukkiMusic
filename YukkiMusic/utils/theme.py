import random
from YukkiMusic.utils.database import get_theme

themes = [
    "alexa1",
    "alexa2",
    "alexa3",
    "alexa4",
    "alexa5",
    "alexa6",
    "alexa7",
    "alexa8",
]


async def check_theme(chat_id: int):
    _theme = await get_theme(chat_id, "theme")
    if not _theme:
        theme = random.choice(themes)
    else:
        theme = _theme["theme"]
        if theme == "Random":
            theme = random.choice(themes)
    return theme
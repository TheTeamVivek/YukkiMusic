import random
from YukkiMusic.utils.database import get_theme

themes = [
    "yukki1",
    "yukki2",
    "yukki3",
    "yukki4",
    "yukki5",
    "yukki6",
    "yukki7",
    "yukki8",
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
#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import random
import re
import string

import lyricsgenius as lg
from telethon import Button

from config import lyrical
from YukkiMusic import BANNED_USERS, tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.utils import language

api_key = "Vd9FvPMOKWfsKJNG9RbZnItaTNIRFzVyyXFdrGHONVsGqHcHBoj3AI3sIlNuqzuf0ZNG8uLcF9wAd5DXBBnUzA"
y = lg.Genius(
    api_key,
    skip_non_songs=True,
    excluded_terms=["(Remix)", "(Live)"],
    remove_section_headers=True,
)
y.verbose = False


@tbot.on_message(flt.command("LYRICS_COMMAND", True) & ~flt.user(BANNED_USERS))
@language
async def lrsearch(event, _):
    if len(event.text.split()) < 2:
        return await event.reply(_["lyrics_1"])
    title = event.message.text.split(None, 1)[1]
    mystic = await event.reply(_["lyrics_2"])
    song = y.search_song(title, get_full_info=False)
    if song is None:
        return await mystic.edit(_["lyrics_3"].format(title))
    ran_hash = "".join(random.choices(string.ascii_uppercase + string.digits, k=10))
    lyric = song.lyrics
    if "Embed" in lyric:
        lyric = re.sub(r"\d*Embed", "", lyric)
    lyrical[ran_hash] = lyric
    upl = [
        [
            Button.inline(
                text=_["L_B_1"],
                url=f"https://t.me/{tbot.username}?start=lyrics_{ran_hash}",
            ),
        ]
    ]
    await mystic.edit(_["lyrics_4"], buttons=upl)

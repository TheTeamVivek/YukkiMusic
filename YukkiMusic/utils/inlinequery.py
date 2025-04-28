#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.


from telethon.tl.custom.inlinebuilder import InlineBuilder
from telethon.tl.types import DocumentAttributeImageSize, InputWebDocument

__all__ = ["answer"]


def thumb(url):
    return InputWebDocument(
        url=url,
        size=0,
        mime_type="image/jpeg",
        attributes=[DocumentAttributeImageSize(w=0, h=0)],
    )


def answer():
    from YukkiMusic import tbot
    builder = InlineBuilder(tbot)

    answer = [
        [
            builder.article(
                title="Pause Stream",
                description="Pause the current playing song on voice chat.",
                thumb=thumb("https://telegra.ph/file/c0a1c789def7b93f13745.png"),
                text="/pause",
            ),
            builder.article(
                title="Resume Stream",
                description="Resume the paused song on voice chat.",
                thumb=thumb("https://telegra.ph/file/02d1b7f967ca11404455a.png"),
                text="/resume",
            ),
            builder.article(
                title="Mute Stream",
                description="Mute the ongoing song on voice chat.",
                thumb=thumb("https://telegra.ph/file/66516f2976cb6d87e20f9.png"),
                text="/vcmute",
            ),
            builder.article(
                title="Unmute Stream",
                description="Unmute the ongoing song on voice chat.",
                thumb=thumb("https://telegra.ph/file/3078794f9341ffd582e18.png"),
                text="/vcunmute",
            ),
            builder.article(
                title="Skip Stream",
                description=(
                    "Skip to next track. | Skip to next track. | "
                    "For specific track number: /skip [number]"
                ),
                thumb=thumb("https://telegra.ph/file/98b88e52bc625903c7a2f.png"),
                text="/skip",
            ),
            builder.article(
                title="End Stream",
                description="Stop the ongoing song on group voice chat.",
                thumb=thumb("https://telegra.ph/file/d2eb03211baaba8838cc4.png"),
                text="/stop",
            ),
            builder.article(
                title="Shuffle Stream",
                description="Shuffle the queued tracks list.",
                thumb=thumb("https://telegra.ph/file/7f6aac5c6e27d41a4a269.png"),
                text="/shuffle",
            ),
            builder.article(
                title="Seek Stream",
                description="Seek the ongoing stream to a specific duration.",
                thumb=thumb("https://telegra.ph/file/cd25ec6f046aa8003cfee.png"),
                text="/seek 10",
            ),
            builder.article(
                title="Loop Stream",
                description=(
                    "Loop the current playing music. Usage: /loop [enable|disable]"
                ),
                thumb=thumb("https://telegra.ph/file/081c20ce2074ea3e9b952.png"),
                text="/loop 3",
            ),
        ]
    ]
    return answer

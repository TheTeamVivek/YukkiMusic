#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

from uuid import uuid4

from telethon.tl.types import (
    DocumentAttributeImageSize,
    InputBotInlineMessageText,
    InputBotInlineResult,
    InputWebDocument,
)


def article(title, description, thumb_url, input_message_content):
    return InputBotInlineResult(
        id=str(uuid4()),
        type="article",
        send_message=InputBotInlineMessageText(message=input_message_content),
        title=title,
        description=description,
        thumb=InputWebDocument(
            url=thumb_url,
            size=0,
            mime_type="image/jpeg",
            attributes=[DocumentAttributeImageSize(w=0, h=0)],
        ),
    )


answer = []

answer.extend(
    [
        article(
            title="Pause Stream",
            description="Pause the current playing song on voice chat.",
            thumb_url="https://telegra.ph/file/c0a1c789def7b93f13745.png",
            input_message_content="/pause",
        ),
        article(
            title="Resume Stream",
            description="Resume the paused song on voice chat.",
            thumb_url="https://telegra.ph/file/02d1b7f967ca11404455a.png",
            input_message_content="/resume",
        ),
        article(
            title="Mute Stream",
            description="Mute the ongoing song on voice chat.",
            thumb_url="https://telegra.ph/file/66516f2976cb6d87e20f9.png",
            input_message_content="/vcmute",
        ),
        article(
            title="Unmute Stream",
            description="Unmute the ongoing song on voice chat.",
            thumb_url="https://telegra.ph/file/3078794f9341ffd582e18.png",
            input_message_content="/vcunmute",
        ),
        article(
            title="Skip Stream",
            description=(
                "Skip to next track. | Skip to next track. | "
                "For specific track number: /skip [number]"
            ),
            thumb_url="https://telegra.ph/file/98b88e52bc625903c7a2f.png",
            input_message_content="/skip",
        ),
        article(
            title="End Stream",
            description="Stop the ongoing song on group voice chat.",
            thumb_url="https://telegra.ph/file/d2eb03211baaba8838cc4.png",
            input_message_content="/stop",
        ),
        article(
            title="Shuffle Stream",
            description="Shuffle the queued tracks list.",
            thumb_url="https://telegra.ph/file/7f6aac5c6e27d41a4a269.png",
            input_message_content="/shuffle",
        ),
        article(
            title="Seek Stream",
            description="Seek the ongoing stream to a specific duration.",
            thumb_url="https://telegra.ph/file/cd25ec6f046aa8003cfee.png",
            input_message_content="/seek 10",
        ),
        article(
            title="Loop Stream",
            description=(
                "Loop the current playing music. Usage: /loop [enable|disable]"
            ),
            thumb_url="https://telegra.ph/file/081c20ce2074ea3e9b952.png",
            input_message_content="/loop 3",
        ),
    ]
)

#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from strings import get_command
from YukkiMusic import tbot
from YukkiMusic.core import filters as flt
from YukkiMusic.core.call import Yukki
from YukkiMusic.misc import BANNED_USERS, db
from YukkiMusic.utils import seconds_to_min
from YukkiMusic.utils.database import get_loop
from YukkiMusic.utils.decorators import admin_rights_check
from YukkiMusic.utils.inline.play import play_markup
from YukkiMusic.utils.stream.autoclear import auto_clean

SKIP_COMMAND = get_command("SKIP_COMMAND")


@tbot.on_message(flt.command(SKIP_COMMAND) & flt.group & ~BANNED_USERS)
@admin_rights_check
async def skip(event, _, chat_id):
    mention = await tbot.create_mention(await event.get_sender())
    if not len(event.text.split()) < 2:
        loop = await get_loop(chat_id)
        if loop != 0:
            return await event.reply(_["admin_12"])
        state = event.text.split(None, 1)[1].strip()
        if state.isnumeric():
            state = int(state)
            check = db.get(chat_id)
            if check:
                count = len(check)
                if count > 2:
                    count = int(count - 1)
                    if 1 <= state <= count:
                        for x in range(state):
                            popped = None
                            try:
                                popped = check.pop(0)
                                if popped.get("mystic"):
                                    try:
                                        await popped.get("mystic").delete()
                                    except Exception:
                                        pass
                            except Exception:
                                return await event.reply(_["admin_16"])
                            if popped:
                                await auto_clean(popped)
                            if not check:
                                try:
                                    await event.reply(
                                        _["admin_10"].format(mention),
                                        link_preview=False,
                                    )
                                    await Yukki.stop_stream(chat_id)
                                except Exception:
                                    return
                                break
                    else:
                        return await event.reply(_["admin_15"].format(count))
                else:
                    return await event.reply(_["admin_14"])
            else:
                return await event.reply(_["queue_2"])
        else:
            return await event.reply(_["admin_13"])
    else:
        check = db.get(chat_id)
        popped = None
        try:
            popped = check.pop(0)
            if popped:
                await auto_clean(popped)
                if popped.get("mystic"):
                    try:
                        await popped.get("mystic").delete()
                    except Exception:
                        pass
            if not check:
                await event.reply(
                    _["admin_10"].format(mention),
                    link_preview=False,
                )
                try:
                    return await Yukki.stop_stream(chat_id)
                except Exception:
                    return
        except Exception:
            try:
                await event.reply(
                    _["admin_10"].format(mention),
                    link_preview=False,
                )
                return await Yukki.stop_stream(chat_id)
            except Exception:
                return
    track = check[0]["track"]
    user = check[0]["by"]
    url = (
        f"https://t.me/{tbot.username}?start=info_{track.vidid}"
        if track.is_youtube
        else track.link
    )
    db[chat_id][0]["played"] = 0
    mystic = await event.reply(_["call_8"], link_preview=False)

    try:
        file_path = await track.download()
        await Yukki.skip_stream(chat_id, file_path, video=track.video)

    except Exception as e:
        await tbot.handle_error(e, event)
        return await mystic.edit(_["call_7"])

    what, button = play_markup(_, chat_id, track)

    run = await event.respond(
        file=track.thumb,
        message=_["stream_1"].format(
            track.title[:27],
            url,
            seconds_to_min(track.duration),
            user,
        ),
        buttons=button,
    )
    db[chat_id][0]["mystic"] = run
    db[chat_id][0]["markup"] = what
    await mystic.delete()

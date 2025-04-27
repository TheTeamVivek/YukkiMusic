from telethon import Button as _btn


def downlod_markup(_):
    """return [
        [
            Button.inline(
                text=_["TG_B_1"],
                data="stop_downloading",
            ),
        ]
    ]"""
    return _btn.inline(text=_["TG_B_1"], data="stop_downloading")

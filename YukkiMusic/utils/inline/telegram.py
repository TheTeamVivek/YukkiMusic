from telethon import Button


# TG_B_1: "🚦 Cancel downloading"
def downlod_markup(_):
    """return [
        [
            Button.inline(
                text=_["TG_B_1"],
                data="stop_downloading",
            ),
        ]
    ]"""
    return Button.inline(text=_["TG_B_1"], data="stop_downloading")

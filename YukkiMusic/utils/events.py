from telethon import events
import re

def NewMessage(*args, **kwargs):
    kwargs["incoming"] = kwargs.get("incoming") or True
    pattern = kwargs.pop("pattern", None)
    if isinstance(pattern, str):
        if not pattern.startswith("^"):
            pattern = rf"^(?:/)?{re.escape(pattern)}"
        kwargs["pattern"] = re.compile(pattern)
    elif isinstance(pattern, list):
        cmd_pattern = '|'.join(re.escape(cmd) for cmd in pattern)
        kwargs["pattern"] = re.compile(rf"^(?:/)?(?:{cmd_pattern})")
    elif isinstance(pattern, re.Pattern):
        kwargs["pattern"] = pattern

    return events.NewMessage(*args, **kwargs)

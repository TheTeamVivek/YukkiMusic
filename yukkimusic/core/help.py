HELP_DATA = {}


class ModuleHelp:
    """
    Register help entries and localized names for a module.
    """

    def __init__(self, module: str):
        """
        Initialize the module entry in HELP_DATA if not exists.

        Args:
            module (str): Internal name of the module (e.g., "Admins")
        """
        self.module = module
        HELP_DATA.setdefault(module, {"names": {}, "entries": {}})

    def __call__(self, module: str):
        """
        Allows calling the instance to create a new ModuleHelp for another module.

        Args:
            module (str): New module name to register

        Returns:
            ModuleHelp: New instance for the new module
        """
        return ModuleHelp(module)

    def name(self, lang: str, text: str):
        """
        Set the localized display name for this module.

        Args:
            lang (str): Language code (e.g., "en", "hi")
            text (str): Localized name for UI use

        Returns:
            ModuleHelp: For method chaining
        """
        HELP_DATA[self.module]["names"][lang] = text
        return self

    def add(self, lang: str, text: str, priority: int = 0):
        """
        Add a help entry in a specific language.

        Args:
            lang (str): Language code (e.g., "en")
            text (str): Help message (e.g., "/ban - Ban a user")
            priority (int): Sort order (higher = earlier)

        Returns:
            ModuleHelp: For method chaining
        """
        HELP_DATA[self.module]["entries"].setdefault(lang, [])
        if not any(
            entry["text"] == text for entry in HELP_DATA[self.module]["entries"][lang]
        ):
            HELP_DATA[self.module]["entries"][lang].append(
                {
                    "text": text,
                    "priority": priority,
                }
            )
        return self


def get_help(module: str, lang: str = "en", sort: bool = True) -> list[dict] | None:
    """
    Get help entries for a module.

    Args:
        module (str): Module name
        lang (str): Preferred language
        sort (bool): Sort by priority descending

    Returns:
        list[dict] | None: Help entries or None
    """
    data = HELP_DATA.get(module, {}).get("entries", {}).get(lang) or HELP_DATA.get(
        module, {}
    ).get("entries", {}).get("en")
    if not data:
        return None
    return sorted(data, key=lambda x: -x["priority"]) if sort else data


def render_help(module: str, lang: str = "en") -> str | None:
    """
    Render help as a formatted string.

    Args:
        module (str): Module name
        lang (str): Preferred language

    Returns:
        str | None: Newline-separated help or None
    """
    entries = get_help(module, lang)
    if not entries:
        return None
    return "\n".join(entry["text"] for entry in entries)

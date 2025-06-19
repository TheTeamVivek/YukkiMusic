HELP_DATA = {}


class ModuleHelp:
    """
    Register help entries for a specific module with multi-language support.
    """

    def __init__(self, module: str):
        """
        Args:
            module (str): Internal module/category name (e.g., "Admins")
        """
        self.module = module
        HELP_DATA.setdefault(module, {})

    def add(self, lang: str, text: str, priority: int = 0):
        """
        Add a help entry for a given language and priority.

        Args:
            lang (str): Language code (e.g., "en", "hi")
            text (str): Help description (e.g., "/play - play a song")
            priority (int): Sorting priority (higher comes first)

        Returns:
            ModuleHelp: to allow method chaining
        """
        HELP_DATA[self.module].setdefault(lang, [])
        if not any(entry["text"] == text for entry in HELP_DATA[self.module][lang]):
            HELP_DATA[self.module][lang].append(
                {
                    "text": text,
                    "priority": priority,
                }
            )
        return self


def get_help(module: str, lang: str = "en", sort: bool = True) -> list[dict] | None:
    """
    Retrieve help entries for a module in a specific language.

    Args:
        module (str): Module/category name
        lang (str): Language code (fallback to 'en' if not found)
        sort (bool): Whether to sort help entries by priority descending

    Returns:
        list[dict] | None: List of help entries or None if not found
    """
    data = HELP_DATA.get(module, {}).get(lang) or HELP_DATA.get(module, {}).get("en")
    if not data:
        return None
    return sorted(data, key=lambda x: -x["priority"]) if sort else data


def render_help(module: str, lang: str = "en") -> str | None:
    """
    Render help entries into a formatted string for display.

    Args:
        module (str): Module name
        lang (str): Language code

    Returns:
        str | None: Newline-separated help text or None if empty
    """
    entries = get_help(module, lang)
    if not entries:
        return None
    return "\n".join(entry["text"] for entry in entries)

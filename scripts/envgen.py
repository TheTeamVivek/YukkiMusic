#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
# pylint: disable=missing-module-docstring, missing-function-docstring, missing-class-docstring

import os
import sys
import termios
import tty
from dataclasses import dataclass


def is_bool(value):
    v = str(value).lower()
    truth_map = {
        "true": True,
        "yes": True,
        "y": True,
        "false": False,
        "no": False,
        "n": False,
    }

    if v in truth_map:
        return True, truth_map[v]

    print_hint('Accepted values: "yes", "no"')
    return False, None


def parse_list(sep=","):
    def _parse(text):
        if not text:
            return False, None
        h = "space-separated" if not sep or not sep.strip() else "comma-separated"
        values = [v.strip() for v in text.strip("'\"").split(sep) if v.strip()]
        if not values:
            print_hint(f"Provide {h} values")
            return False, None
        return True, values

    return _parse


def isint(text):
    v = text
    if not v:
        return False, None
    try:
        v = int(v)
    except ValueError:
        return False, None
    return True, v


def _str(t):
    return True, str(t)


required_vars = {
    "API_ID": ("Enter your API_ID from my.telegram.org", True, isint),
    "API_HASH": ("Enter your API_HASH from my.telegram.org", True, _str),
    "BOT_TOKEN": ("Enter your bot token from @BotFather", True, _str),
    "MONGO_DB_URI": ("MongoDB URI for database connection", True, _str),
    "LOG_GROUP_ID": ("Group ID for logging bot errors and events", True, isint),
    "OWNER_ID": ("Owner Telegram user ID(s), space-separated", True, parse_list(None)),
    "STRING_SESSIONS": (
        "List of Pyrogram string sessions (comma-separated)",
        False,
        parse_list(),
    ),
}

recommended_vars = {
    "COOKIE_LINK": (
        "List of batbin.me cookie links must be seperated by a space",
        False,
        parse_list(None),
    ),
    "SUPPORT_CHANNEL": ("Support channel link (starts with https://)", False, _str),
    "SUPPORT_GROUP": ("Support group link (starts with https://)", False, _str),
    "AUTO_LEAVING_ASSISTANT": ("Enable assistant auto-leave?", False, is_bool),
    "AUTO_LEAVE_ASSISTANT_TIME": ("Time before assistant leaves", False, isint),
    "UPSTREAM_REPO": ("Upstream repository URL", False, _str),
    "UPSTREAM_BRANCH": ("Upstream branch name", False, _str),
    "GITHUB_REPO": (
        "Your GitHub repo URL that will show in start as Source",
        False,
        _str,
    ),
    "GIT_TOKEN": ("GitHub token (if UPSTREAM_RPEO is private)", False, _str),
    "SPOTIFY_CLIENT_ID": ("Spotify Client ID", False, _str),
    "SPOTIFY_CLIENT_SECRET": ("Spotify Client Secret", False, _str),
}

optional_vars = {
    "DURATION_LIMIT": ("Maximum playback duration in minutes", False, isint),
    "SONG_DOWNLOAD_DURATION_LIMIT": ("Max download duration in minutes", False, isint),
    "VIDEO_STREAM_LIMIT": ("Max number of video streams", False, isint),
    "PRIVATE_BOT_MODE": ("Enable private bot mode?", False, is_bool),
    "CLEANMODE_MINS": ("Interval for clean mode (in mins)", False, isint),
    "YOUTUBE_EDIT_SLEEP": ("Sleep between yt-dl edit messages", False, isint),
    "TELEGRAM_EDIT_SLEEP": ("Sleep between TG edit messages", False, isint),
    "SERVER_PLAYLIST_LIMIT": ("Limit for saving server playlists", False, isint),
    "PLAYLIST_FETCH_LIMIT": ("Max fetch count from playlist URLs", False, isint),
    "TG_AUDIO_FILESIZE_LIMIT": ("Max Telegram audio size (bytes)", False, isint),
    "TG_VIDEO_FILESIZE_LIMIT": ("Max Telegram video size (bytes)", False, isint),
    "SET_CMDS": ("Should setup commands automatically?", False, is_bool),
    "START_IMG_URL": ("Start image URL", False, _str),
    "PING_IMG_URL": ("Ping image URL", False, _str),
    "PLAYLIST_IMG_URL": ("Playlist image URL", False, _str),
    "GLOBAL_IMG_URL": ("Global image URL", False, _str),
    "STATS_IMG_URL": ("Stats image URL", False, _str),
    "TELEGRAM_AUDIO_URL": ("Audio image URL", False, _str),
    "TELEGRAM_VIDEO_URL": ("Video image URL", False, _str),
    "STREAM_IMG_URL": ("Stream image URL", False, _str),
    "SOUNCLOUD_IMG_URL": ("SoundCloud image URL", False, _str),
    "YOUTUBE_IMG_URL": ("YouTube image URL", False, _str),
    "SPOTIFY_ARTIST_IMG_URL": ("Spotify artist image URL", False, _str),
    "SPOTIFY_ALBUM_IMG_URL": ("Spotify album image URL", False, _str),
    "SPOTIFY_PLAYLIST_IMG_URL": ("Spotify playlist image URL", False, _str),
}


@dataclass
class Colors:
    HEADER = "\033[95m"
    OKBLUE = "\033[94m"
    OKCYAN = "\033[96m"
    OKGREEN = "\033[92m"
    WARNING = "\033[93m"
    FAIL = "\033[91m"
    ENDC = "\033[0m"
    BOLD = "\033[1m"
    UNDERLINE = "\033[4m"


def getch():
    fd = sys.stdin.fileno()
    old_settings = termios.tcgetattr(fd)
    try:
        tty.setraw(fd)
        return sys.stdin.read(1)
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)


def ask_yes_no(prompt="Do you want to continue? (y/n): "):
    print(prompt, end="", flush=True)
    while True:
        ch = getch().lower()
        if ch in ("y", "n"):
            print(ch)
            return ch == "y"


def cprint(msg, color="", end="\n"):
    print(color + str(msg) + Colors.ENDC, end=end)


def print_warning(msg):
    cprint(f"[!] {msg}", Colors.WARNING)


def print_error(msg):
    cprint(f"[✖] {msg}", Colors.FAIL)


def print_success(msg):
    cprint(f"[✓] {msg}", Colors.OKGREEN)


def print_hint(msg):
    cprint(f"[*] {msg}", Colors.OKBLUE)


def print_normal(msg):
    cprint(f"[-] {msg}", Colors.OKCYAN)


def input_colored(prompt, color=""):
    return input(color + Colors.BOLD + prompt + Colors.ENDC)


def print_section(title):
    border = f"{Colors.OKGREEN}{'=' * (len(title) + 10)}{Colors.ENDC}"
    print(f"\n{border}")
    print(f"{Colors.BOLD}{Colors.OKGREEN}     {title}{Colors.ENDC}")
    print(f"{border}\n")


def get_input(var_name, prompt_text, required=True, validator=None, default=None):
    while True:
        print_hint(f"⮞ Configuring: {var_name}")
        suffix = f" (default: {default})" if default else ""
        try:
            user_input = input_colored(f"{prompt_text}{suffix}\n> ", Colors.OKCYAN)
        except KeyboardInterrupt:
            print_error("\n[!] Input cancelled by user.")
            sys.exit(0)

        if not user_input:
            if default is not None:
                user_input = str(default)
                print_normal(f"Using default for {var_name}: {user_input}")
            elif required:
                print_warning(
                    f'Required value for "{var_name}" was not provided. Please enter it.'
                )
                continue
            else:
                print_warning(f'"{var_name}" not provided. Skipping.')
                return None

        if validator:
            is_valid, value = validator(user_input)
            if not is_valid:
                print_error(f'Invalid input for "{var_name}". Please try again.')
                continue
            print_success(f"✓ {var_name} set to: {value}")
            return value
        print_success(f"✓ {var_name} set to: {user_input}")
        return user_input


def load_config(path="config/config.py"):
    loc = {}
    if os.path.exists(path):
        with open(path, encoding="utf-8") as f:  # noqa
            exec(f.read(), loc)  # pylint: disable=W0122
        print_success(f"Loaded default values from {path}")
    else:
        print_warning(f"No config file found at {path}. Continuing without defaults.")
    return loc


def main():
    loc = load_config()
    env_lines = []
    print_section("Configure Required Variables")
    for key, (prompt_text, required, validator) in required_vars.items():
        default = loc.get(key)
        value = get_input(key, prompt_text, required, validator, default)
        if isinstance(value, list):
            value = ",".join(map(str, value))
        env_lines.append(f"{key}={value}")

    print_section("Configure Optional Variables")
    if ask_yes_no("Do you want to configure optional variables? (y/n): "):
        optional_to_configure = recommended_vars.copy()
        if ask_yes_no("Configure all optional variables? (y/n): "):
            optional_to_configure.update(optional_vars)
        for key, (prompt_text, required, validator) in optional_to_configure.items():
            default = loc.get(key)
            value = get_input(key, prompt_text, required, validator, default)
            if value is not None:
                if isinstance(value, list):
                    value = ",".join(map(str, value))
                env_lines.append(f"{key}={value}")

    with open(".env", "w", encoding="utf-8") as f:
        f.write("\n".join(env_lines) + "\n")

    print_section("Configuration Complete")
    print_success("✅ Environment variables saved to .env")


if __name__ == "__main__":
    main()
